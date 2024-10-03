package websocket

import (
	"math"
	"sync"

	"github.com/kyaxcorp/go-helper/array"
)

type FindClientsFilter struct {
	All bool
	// Users (ID's)
	Users []string
	// Devices (ID's)
	Devices []string
	// AuthTokens (Tokens)
	AuthTokens []string
	// Connections (ID's)
	Connections []uint64
	// IP Addresses (Addresses)
	IPAddresses []string
	// Route Paths
	RequestPaths []string
	// Exception List (usually used in tandem with All param) if sending to everyone
	ExceptConnections  []uint64
	ExceptUsers        []string
	ExceptDevices      []string
	ExceptAuthTokens   []string
	ExceptIPAddresses  []string
	ExceptRequestPaths []string

	// Exception List Maps (converted automatically when the struct it's being created)
	exceptConnectionsMap  map[uint64]int
	exceptUsersMap        map[string]int
	exceptDevicesMap      map[string]int
	exceptAuthTokensMap   map[string]int
	exceptIPAddressesMap  map[string]int
	exceptRequestPathsMap map[string]int

	// Exception List Checks if are enabled
	isExceptConnections  bool
	isExceptUsers        bool
	isExceptDevices      bool
	isExceptAuthTokens   bool
	isExceptIPAddresses  bool
	isExceptRequestPaths bool
}

func prepareFilter(filter *FindClientsFilter) *FindClientsFilter {
	// TODO: use Goroutines here

	var awaitGroup sync.WaitGroup
	awaitGroup.Add(6)

	// Create the exceptions filters map for better indexing and performance
	go func() {
		defer func() {
			awaitGroup.Done()
		}()
		if len(filter.ExceptConnections) > 0 {
			filter.isExceptConnections = true
			filter.exceptConnectionsMap = array.ConvUint64ValuesToMapKey(filter.ExceptConnections)
		}
	}()
	go func() {
		defer func() {
			awaitGroup.Done()
		}()
		if len(filter.ExceptUsers) > 0 {
			filter.isExceptUsers = true
			filter.exceptUsersMap = array.ConvStringValuesToMapKey(filter.ExceptUsers)
		}
	}()
	go func() {
		defer func() {
			awaitGroup.Done()
		}()
		if len(filter.ExceptDevices) > 0 {
			filter.isExceptDevices = true
			filter.exceptDevicesMap = array.ConvStringValuesToMapKey(filter.ExceptDevices)
		}
	}()
	go func() {
		defer func() {
			awaitGroup.Done()
		}()
		if len(filter.ExceptAuthTokens) > 0 {
			filter.isExceptAuthTokens = true
			filter.exceptAuthTokensMap = array.ConvStringValuesToMapKey(filter.ExceptAuthTokens)
		}
	}()
	go func() {
		defer func() {
			awaitGroup.Done()
		}()
		if len(filter.ExceptIPAddresses) > 0 {
			filter.isExceptIPAddresses = true
			filter.exceptIPAddressesMap = array.ConvStringValuesToMapKey(filter.ExceptIPAddresses)
		}
	}()
	go func() {
		defer func() {
			awaitGroup.Done()
		}()
		if len(filter.ExceptRequestPaths) > 0 {
			filter.isExceptRequestPaths = true
			filter.exceptRequestPathsMap = array.ConvStringValuesToMapKey(filter.ExceptRequestPaths)
		}
	}()
	awaitGroup.Wait()
	return filter
}

type clientsData struct {
	// Locks
	clientsLock sync.RWMutex

	// Data
	clients map[*Client]bool
	// Indexes
	clientsIndex   ClientsIndex
	enableIndexing bool
}

func (c *clientsData) GetNrOfClients() uint {
	defer func() {
		c.clientsLock.RUnlock()
	}()
	c.clientsLock.RLock()
	return uint(len(c.clients))
}

func (c *clientsData) createUsersIndexMap(client *Client) {
	if _, ok := c.clientsIndex.Users[client.authDetails.UserDetails.UserID]; !ok {
		// Create the map!
		c.clientsIndex.Users[client.authDetails.UserDetails.UserID] = make(map[uint64]*Client)
	}
}

func (c *clientsData) createDevicesIndexMap(client *Client) {
	if _, ok := c.clientsIndex.Devices[client.authDetails.DeviceDetails.DeviceID]; !ok {
		// Create the map!
		c.clientsIndex.Devices[client.authDetails.DeviceDetails.DeviceID] = make(map[uint64]*Client)
	}
}

func (c *clientsData) createAuthTokensIndexMap(client *Client) {
	if _, ok := c.clientsIndex.AuthTokens[client.authDetails.AuthTokenDetails.Token]; !ok {
		// Create the map!
		c.clientsIndex.AuthTokens[client.authDetails.AuthTokenDetails.Token] = make(map[uint64]*Client)
	}
}

func (c *clientsData) createIPAddressesIndexMap(client *Client) {
	if _, ok := c.clientsIndex.IPAddresses[client.connDetails.ClientIPAddress]; !ok {
		// Create the map!
		c.clientsIndex.IPAddresses[client.connDetails.ClientIPAddress] = make(map[uint64]*Client)
	}
}

func (c *clientsData) createRequestPathIndexMap(client *Client) {
	if _, ok := c.clientsIndex.RequestPath[client.connDetails.RequestPath]; !ok {
		// Create the map!
		c.clientsIndex.RequestPath[client.connDetails.RequestPath] = make(map[uint64]*Client)
	}
}

func (c *clientsData) createIndexes(client *Client) {
	connectionID := client.connectionID

	// ------------------Add to indexes for faster finding!-------------------\\
	// By Connection ID
	c.clientsIndex.connectionsLock.Lock()
	c.clientsIndex.Connections[connectionID] = client
	c.clientsIndex.connectionsLock.Unlock()

	// Check if authDetails exists!
	if client.authDetails != nil {
		// Users
		if client.authDetails.UserDetails.UserID != "" {
			c.clientsIndex.usersLock.Lock()
			c.createUsersIndexMap(client)
			// Add the reference to index
			c.clientsIndex.Users[client.authDetails.UserDetails.UserID][connectionID] = client
			c.clientsIndex.usersLock.Unlock()
		}

		// Devices
		if client.authDetails.DeviceDetails.DeviceID != "" {
			c.clientsIndex.devicesLock.Lock()
			c.createDevicesIndexMap(client)
			// Add the reference to index
			c.clientsIndex.Devices[client.authDetails.DeviceDetails.DeviceID][connectionID] = client
			c.clientsIndex.devicesLock.Unlock()
		}

		// Auth Tokens
		if client.authDetails.AuthTokenDetails.Token != "" {
			c.clientsIndex.authTokensLock.Lock()
			c.createAuthTokensIndexMap(client)
			// Add the reference to index
			c.clientsIndex.AuthTokens[client.authDetails.AuthTokenDetails.Token][connectionID] = client
			c.clientsIndex.authTokensLock.Unlock()
		}
	}

	if client.connDetails != nil {
		// IP Addresses
		if client.connDetails.ClientIPAddress != "" {
			c.clientsIndex.ipAddressesLock.Lock()
			c.createIPAddressesIndexMap(client)
			// Add the reference to index
			c.clientsIndex.IPAddresses[client.connDetails.ClientIPAddress][connectionID] = client
			c.clientsIndex.ipAddressesLock.Unlock()
		}

		// Request Path / Request URI / ROUTE PATH
		if client.connDetails.RequestPath != "" {
			c.clientsIndex.requestPathLock.Lock()
			c.createRequestPathIndexMap(client)
			// Add the reference to index
			c.clientsIndex.RequestPath[client.connDetails.RequestPath][connectionID] = client
			c.clientsIndex.requestPathLock.Unlock()
		}
	}
	// ------------------Add to indexes for faster finding!-------------------\\
}

func (c *clientsData) unsetIndexes(client *Client) {
	connectionID := client.connectionID

	// Delete from connections
	c.clientsIndex.connectionsLock.Lock()
	if _, ok := c.clientsIndex.Connections[connectionID]; ok {
		delete(c.clientsIndex.Connections, connectionID)
	}
	c.clientsIndex.connectionsLock.Unlock()

	// Delete from users
	if client.authDetails != nil {

		if client.authDetails.UserDetails.UserID != "" {
			c.clientsIndex.usersLock.Lock()
			c.createUsersIndexMap(client)
			if _, ok := c.clientsIndex.Users[client.authDetails.UserDetails.UserID][connectionID]; ok {
				delete(c.clientsIndex.Users[client.authDetails.UserDetails.UserID], connectionID)
			}
			c.clientsIndex.usersLock.Unlock()
		}

		// Devices
		if client.authDetails.DeviceDetails.DeviceID != "" {
			c.clientsIndex.devicesLock.Lock()
			c.createDevicesIndexMap(client)
			if _, ok := c.clientsIndex.Devices[client.authDetails.DeviceDetails.DeviceID][connectionID]; ok {
				delete(c.clientsIndex.Devices[client.authDetails.DeviceDetails.DeviceID], connectionID)
			}
			c.clientsIndex.devicesLock.Unlock()
		}

		// Auth Tokens
		if client.authDetails.AuthTokenDetails.Token != "" {
			c.clientsIndex.authTokensLock.Lock()
			c.createAuthTokensIndexMap(client)
			if _, ok := c.clientsIndex.AuthTokens[client.authDetails.AuthTokenDetails.Token][connectionID]; ok {
				delete(c.clientsIndex.AuthTokens[client.authDetails.AuthTokenDetails.Token], connectionID)
			}
			c.clientsIndex.authTokensLock.Unlock()
		}
	}

	if client.connDetails != nil {
		// IP Addresses
		if client.connDetails.ClientIPAddress != "" {
			c.clientsIndex.ipAddressesLock.Lock()
			c.createIPAddressesIndexMap(client)
			if _, ok := c.clientsIndex.IPAddresses[client.connDetails.ClientIPAddress][connectionID]; ok {
				delete(c.clientsIndex.IPAddresses[client.connDetails.ClientIPAddress], connectionID)
			}
			c.clientsIndex.ipAddressesLock.Unlock()
		}

		// Request Paths
		if client.connDetails.RequestPath != "" {
			c.clientsIndex.requestPathLock.Lock()
			c.createRequestPathIndexMap(client)
			if _, ok := c.clientsIndex.RequestPath[client.connDetails.RequestPath][connectionID]; ok {
				delete(c.clientsIndex.RequestPath[client.connDetails.RequestPath], connectionID)
			}
			c.clientsIndex.requestPathLock.Unlock()
		}
	}
}

func (c *clientsData) GetClientByConnID(connectionID uint64) *Client {
	c.clientsIndex.connectionsLock.RLock()
	client, ok := c.clientsIndex.Connections[connectionID]
	c.clientsIndex.connectionsLock.RUnlock()
	if !ok {
		return nil
	}
	return client
}

func (c *clientsData) GetClients() map[*Client]bool {
	defer func() {
		c.clientsLock.RUnlock()
	}()
	// We are copying the clients into a new map because
	// if returning directly the map, the programmer should handle the locks
	// But if you don't handle locks, then we should copy the map to be store in a diff address space
	// Returning Maps is like returning pointers
	nmap := make(map[*Client]bool)
	c.clientsLock.RLock()
	for k, v := range c.clients {
		nmap[k] = v
	}
	return nmap
}

func (c *clientsData) GetClientsOrderedByConnectionID() map[int64]*Client {
	defer func() {
		c.clientsLock.RUnlock()
	}()
	nmap := make(map[int64]*Client)
	c.clientsLock.RLock()
	for cl, _ := range c.clients {
		nmap[int64(cl.connectionID)] = cl
	}
	return nmap
}

func (c *clientsData) GetClientsInChunks(nrOfChunks uint16) []map[*Client]bool {
	defer c.clientsLock.RUnlock()
	c.clientsLock.RLock()
	return GetClientsInChunks(c.clients, nrOfChunks)
}

// the programmer should handle locks before!!
// func GetClientsInChunks(clients interface{}, nrOfChunks uint16) []map[*Client]bool {
func GetClientsInChunks(clients map[*Client]bool, nrOfChunks uint16) []map[*Client]bool {
	/*reflect.TypeOf(clients)

	switch clients.(type) {
	case map[*Client]bool:

	case map[uint64]*Client:

	}*/

	nrOfConnections := len(clients)
	if nrOfConnections == 0 {
		return nil
	}
	if nrOfChunks == 0 {
		return nil
	}

	chunks := make([]map[*Client]bool, nrOfChunks)
	nr := int(nrOfChunks)
	// Create the chunks first! why? because we don't need to make multiple verifications
	// in the clients loop! so it's better to do it here!
	for i := 0; i < nr; i++ {
		nmap := make(map[*Client]bool)
		chunks[i] = nmap
	}
	itemsPerChunk := int(math.Ceil(float64(nrOfConnections) / float64(nr)))
	currentChunkNr := 0
	added := 0
	for k, _ := range clients {
		chunks[currentChunkNr][k] = true
		added++ // How many have being added into the current chunk!
		if added == itemsPerChunk {
			added = 0        // Reset back!
			currentChunkNr++ // Next Chunk
		}
	}
	return chunks
}

func GetClientsInChunksWithConn(clients map[uint64]*Client, nrOfChunks uint16) []map[uint64]*Client {
	nrOfConnections := len(clients)
	if nrOfConnections == 0 {
		return nil
	}
	if nrOfChunks == 0 {
		return nil
	}

	chunks := make([]map[uint64]*Client, nrOfChunks)

	nr := int(nrOfChunks)
	// Create the chunks first! why? because we don't need to make multiple verifications
	// in the clients loop! so it's better to do it here!
	for i := 0; i < nr; i++ {
		nmap := make(map[uint64]*Client)
		chunks[i] = nmap
	}
	itemsPerChunk := int(math.Ceil(float64(nrOfConnections) / float64(nr)))
	currentChunkNr := 0
	added := 0
	for connId, c := range clients {
		chunks[currentChunkNr][connId] = c
		added++ // How many have being added into the current chunk!
		if added == itemsPerChunk {
			added = 0        // Reset back!
			currentChunkNr++ // Next Chunk
		}
	}
	return chunks
}

func (c *clientsData) GetClientsList() map[uint64]*Client {

	c.clientsIndex.connectionsLock.RLock()
	nmap := make(map[uint64]*Client)
	for k, v := range c.clientsIndex.Connections {
		nmap[k] = v
	}
	c.clientsIndex.connectionsLock.RUnlock()
	return nmap
}

func (c *clientsData) GetClientsByUserID(userID string) map[uint64]*Client {
	c.clientsIndex.usersLock.RLock()
	clients, ok := c.clientsIndex.Users[userID]
	c.clientsIndex.usersLock.RUnlock()
	if !ok || len(clients) == 0 {
		return nil
	}
	return clients
}

func (c *clientsData) GetClientsByDeviceID(deviceID string) map[uint64]*Client {
	c.clientsIndex.devicesLock.RLock()
	clients, ok := c.clientsIndex.Devices[deviceID]
	c.clientsIndex.devicesLock.RUnlock()
	if !ok || len(clients) == 0 {
		return nil
	}
	return clients
}

func (c *clientsData) GetClientsByAuthToken(authToken string) map[uint64]*Client {
	c.clientsIndex.authTokensLock.RLock()
	clients, ok := c.clientsIndex.AuthTokens[authToken]
	c.clientsIndex.authTokensLock.RUnlock()
	if !ok || len(clients) == 0 {
		return nil
	}
	return clients
}

func (c *clientsData) GetClientsByIPAddress(ipAddress string) map[uint64]*Client {
	c.clientsIndex.ipAddressesLock.RLock()
	clients, ok := c.clientsIndex.IPAddresses[ipAddress]
	c.clientsIndex.ipAddressesLock.RUnlock()
	if !ok || len(clients) == 0 {
		return nil
	}
	return clients
}

func (c *clientsData) GetClientsByRequestPath(requestPath string) map[uint64]*Client {
	c.clientsIndex.requestPathLock.RLock()
	clients, ok := c.clientsIndex.RequestPath[requestPath]
	c.clientsIndex.requestPathLock.RUnlock()
	if !ok || len(clients) == 0 {
		return nil
	}
	return clients
}

func (c *clientsData) registerClient(client *Client) *clientsData {
	go func() {
		c.clientsLock.Lock()
		c.clients[client] = true
		c.clientsLock.Unlock()

		if c.enableIndexing {
			go c.createIndexes(client)
		}
	}()
	return c
}
func (c *clientsData) IsClientExist(client *Client) bool {
	c.clientsLock.RLock()
	defer c.clientsLock.RUnlock()
	if _, ok := c.clients[client]; ok {
		return true
	}
	return false
}

func (c *clientsData) unregisterClient(client *Client) *clientsData {
	go func() {
		if c.enableIndexing {
			go c.unsetIndexes(client)
		}

		c.clientsLock.Lock()
		if _, ok := c.clients[client]; ok {
			// Remove the element from map
			delete(c.clients, client)
		}
		c.clientsLock.Unlock()
	}()
	return c
}

func (c *clientsData) getClientsByFilter(filter FindClientsFilter) map[uint64]*Client {
	var nrOfLaunchedSearches = 0
	var awaitGroup sync.WaitGroup

	// Prepare the filter..
	prepareFilter(&filter)

	// TODO: we can also save where we have found the ClientsStatus
	// We are searching through 6 indexes!

	// There are no overheads on local slices when using len function!
	foundClientsChan := make(chan map[uint64]*Client)

	if filter.All {
		// Send to all, but there are also exceptions!
		// Get all available connections
		// Iterate all connections
		// Check one by one for all exceptions

		clientsMap := c.GetClients()
		clients := make(map[uint64]*Client)
		for client, _ := range clientsMap {
			// Check in filters

			// So here we should check for each user:
			/*
				- Connection ID
				- User ID
				- Device ID
				- IP Address
				- Request Path
				- Auth Token
			*/

			connectionId := client.GetConnectionID()

			// TODO: also we should convert all exception lists to maps for better performance
			// But if we convert to maps, many will be overwritten... like usernames, auth tokens... that's ok!

			if filter.isExceptDevices {
				if _, ok := filter.exceptDevicesMap[client.GetDeviceID()]; ok {
					// If found, then skip it!
					continue
				}
			}
			if filter.isExceptUsers {
				if _, ok := filter.exceptUsersMap[client.GetUserID()]; ok {
					// If found, then skip it!
					continue
				}
			}
			if filter.isExceptConnections {
				if _, ok := filter.exceptConnectionsMap[connectionId]; ok {
					// If found, then skip it!
					continue
				}
			}
			if filter.isExceptIPAddresses {
				if _, ok := filter.exceptIPAddressesMap[client.GetIPAddress()]; ok {
					// If found, then skip it!
					continue
				}
			}
			if filter.isExceptAuthTokens {
				if _, ok := filter.exceptAuthTokensMap[client.GetAuthToken()]; ok {
					// If found, then skip it!
					continue
				}
			}
			if filter.isExceptRequestPaths {
				if _, ok := filter.exceptRequestPathsMap[client.GetRequestPath()]; ok {
					// If found, then skip it!
					continue
				}
			}
			clients[connectionId] = client
		}
		return clients
	}

	// Check if defined - Users
	if len(filter.Users) > 0 {
		awaitGroup.Add(1)
		nrOfLaunchedSearches++
		go func() {
			defer func() {
				awaitGroup.Done()
			}()
			local := make(map[uint64]*Client)
			for _, userID := range filter.Users {
				if userID != "" {
					if filter.isExceptUsers {
						// Check if it's not excluded
						if _, ok := filter.exceptUsersMap[userID]; ok {
							continue
						}
					}

					tmpClients := c.GetClientsByUserID(userID)
					if tmpClients != nil {
						for connectionID, client := range tmpClients {
							if _, ok := local[connectionID]; !ok {
								local[connectionID] = client
							}
						}
					}
				}
			}
			foundClientsChan <- local
		}()
	}

	// Devices
	if len(filter.Devices) > 0 {
		awaitGroup.Add(1)
		nrOfLaunchedSearches++
		go func() {
			defer func() {
				awaitGroup.Done()
			}()

			local := make(map[uint64]*Client)
			for _, deviceID := range filter.Devices {
				if deviceID != "" {
					if filter.isExceptDevices {
						// Check if it's not excluded
						if _, ok := filter.exceptDevicesMap[deviceID]; ok {
							continue
						}
					}
					tmpClients := c.GetClientsByDeviceID(deviceID)
					if tmpClients != nil {
						for connectionID, client := range tmpClients {
							if _, ok := local[connectionID]; !ok {
								local[connectionID] = client
							}
						}
					}
				}
			}
			foundClientsChan <- local
		}()
	}

	// By auth tokens
	if len(filter.AuthTokens) > 0 {
		awaitGroup.Add(1)
		nrOfLaunchedSearches++
		go func() {
			defer func() {
				awaitGroup.Done()
			}()
			local := make(map[uint64]*Client)
			for _, authToken := range filter.AuthTokens {
				if authToken != "" {
					if filter.isExceptAuthTokens {
						// Check if it's not excluded
						if _, ok := filter.exceptAuthTokensMap[authToken]; ok {
							continue
						}
					}
					tmpClients := c.GetClientsByAuthToken(authToken)
					if tmpClients != nil {
						for connectionID, client := range tmpClients {
							if _, ok := local[connectionID]; !ok {
								local[connectionID] = client
							}
						}
					}
				}
			}
			foundClientsChan <- local
		}()

	}

	// By IP Addresses
	if len(filter.IPAddresses) > 0 {
		awaitGroup.Add(1)
		nrOfLaunchedSearches++
		go func() {
			defer func() {
				awaitGroup.Done()
			}()
			local := make(map[uint64]*Client)

			for _, ipAddress := range filter.IPAddresses {
				if ipAddress != "" {
					if filter.isExceptIPAddresses {
						// Check if it's not excluded
						if _, ok := filter.exceptIPAddressesMap[ipAddress]; ok {
							continue
						}
					}
					tmpClients := c.GetClientsByIPAddress(ipAddress)
					if tmpClients != nil {
						for connectionID, client := range tmpClients {
							if _, ok := local[connectionID]; !ok {
								local[connectionID] = client
							}
						}
					}
				}
			}
			foundClientsChan <- local
		}()
	}

	// By Route Paths
	if len(filter.RequestPaths) > 0 {
		awaitGroup.Add(1)
		nrOfLaunchedSearches++
		go func() {
			defer func() {
				awaitGroup.Done()
			}()
			local := make(map[uint64]*Client)

			for _, requestPath := range filter.RequestPaths {
				if requestPath != "" {
					if filter.isExceptRequestPaths {
						// Check if it's not excluded
						if _, ok := filter.exceptRequestPathsMap[requestPath]; ok {
							continue
						}
					}
					tmpClients := c.GetClientsByRequestPath(requestPath)
					if tmpClients != nil {
						for connectionID, client := range tmpClients {
							if _, ok := local[connectionID]; !ok {
								local[connectionID] = client
							}
						}
					}
				}
			}
			foundClientsChan <- local
		}()
	}

	// By Connection ID's
	if len(filter.Connections) > 0 {
		awaitGroup.Add(1)
		nrOfLaunchedSearches++
		go func() {
			defer func() {
				awaitGroup.Done()
			}()
			local := make(map[uint64]*Client)

			for _, connectionID := range filter.Connections {
				if connectionID != 0 {
					if filter.isExceptConnections {
						// Check if it's not excluded
						if _, ok := filter.exceptConnectionsMap[connectionID]; ok {
							continue
						}
					}
					tmpClient := c.GetClientByConnID(connectionID)
					if tmpClient != nil {
						if _, ok := local[connectionID]; !ok {
							local[connectionID] = tmpClient
						}
					}
				}
			}

			foundClientsChan <- local
		}()
	}

	clients := make(map[uint64]*Client)
	doneAll := make(chan bool)
	go func() {
		// Wait for all goroutines to finish
		awaitGroup.Wait()
		// Notify the main process that everything has being finished!
		doneAll <- true
	}()

	finishedAll := false
	receivedData := 0
	for {
		select {
		case foundClients := <-foundClientsChan:
			// Add to clients var
			receivedData++
			for connectionID, client := range foundClients {
				if _, ok := clients[connectionID]; !ok {
					clients[connectionID] = client
				}
			}
		case <-doneAll:
			// Break from loop and return clients!
			finishedAll = true
		default:
			if receivedData == nrOfLaunchedSearches {
				break
			}
		}

		if finishedAll && receivedData == nrOfLaunchedSearches {
			break
		}
	}

	return clients
}

func NewClientsInstance() *clientsData {
	return &clientsData{
		// Here we store the c
		clients: make(map[*Client]bool),
		// Creating map of ClientsStatus indexes
		clientsIndex: ClientsIndex{
			Users:       make(map[string]map[uint64]*Client),
			Devices:     make(map[string]map[uint64]*Client),
			Connections: make(map[uint64]*Client),
			AuthTokens:  make(map[string]map[uint64]*Client),
			IPAddresses: make(map[string]map[uint64]*Client),
			RequestPath: make(map[string]map[uint64]*Client),
		},
		// Allow indexing
		enableIndexing: true,
	}
}

func (s *Server) GetClientsByFilter(filter FindClientsFilter) map[uint64]*Client {
	return s.c.getClientsByFilter(filter)
}
