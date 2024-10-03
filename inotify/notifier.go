package inotify

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/kyaxcorp/go-helper/file"
	"github.com/kyaxcorp/go-helper/filesystem"
	"github.com/kyaxcorp/go-helper/filesystem/fsnotify"
	server "github.com/kyaxcorp/go-websocket"
	"github.com/rs/zerolog"
)

type ListeningPath struct {
	Path string
	// This is the file regex
	FileRegex string
	// Delete all existing files before starting listening!
	ClearBeforeListen bool
}

type OnNotificationRead func(h *server.Hub, data interface{}, plainData string)

type WSNotifier struct {
	WebSocketServer             *server.Server
	ListeningPaths              []ListeningPath
	WSHub                       *server.Hub
	DelayRead                   time.Duration
	OnNotificationReadGoRoutine bool
	OnNotificationRead          OnNotificationRead
	OnError                     func(msg string)
}

func (wsNotifier *WSNotifier) onError(msg string) {
	if wsNotifier.OnError != nil {
		wsNotifier.OnError(msg)
	}
}

func New(wsNotifier *WSNotifier) *WSNotifier {
	wsNotifier.WSHub = wsNotifier.WebSocketServer.NewHub(func(h *server.Hub) {
		warn := func() *zerolog.Event {
			return wsNotifier.WebSocketServer.LWarnF("ws_notifier")
		}
		_error := func() *zerolog.Event {
			return wsNotifier.WebSocketServer.LErrorF("ws_notifier")
		}
		// Get the data!
		if h.StopCalled.Get() {
			return
		}

		// Create the file notifier!
		fsNotifier, err := fsnotify.New()
		if err != nil {
			errorMsg := "Failed to create notifier..."
			wsNotifier.onError(errorMsg)
			_error().Err(err).Msg(errorMsg)
			return
		}

		defer func() {
			fsNotifier.Stop()
			fsNotifier = nil
		}()

		for _, listeningPath := range wsNotifier.ListeningPaths {
			if listeningPath.ClearBeforeListen {
				// Clear before starting...
				scanDir, _ := filesystem.RealPath(listeningPath.Path)
				if scanDir != "" {
					// Delete only old notifications
					//matches, err := filepath.Glob(scanDir + filesystem.DirSeparator() + "*.notif")
					matches, globErr := filepath.Glob(scanDir + filesystem.DirSeparator() + listeningPath.FileRegex)
					if globErr == nil {
						for _, match := range matches {
							file.Unlink(match)
						}
					}
				}
			}

			// Add to watcher!
			fsNotifier.Watch(listeningPath.Path, fsnotify.WatchConfig{
				CallbackGoRoutine: wsNotifier.OnNotificationReadGoRoutine,
				Callback: func(e fsnotify.EventData) {

					if wsNotifier.DelayRead > 0 {
						time.Sleep(wsNotifier.DelayRead)
					}

					// Check if it's a file!
					if !file.IsFile(e.Path) {
						warnMsg := "Failed to create notifier..."
						warn().Msg(warnMsg)
						return
					}

					data, getErr := file.GetContents(e.Path)

					// check if it's a json!

					if getErr != nil {
						errorMsg := "failed reading file"
						wsNotifier.onError(errorMsg)
						_error().Err(getErr).Str("file_path", e.Path).Msg(errorMsg)
						return
					}

					var d json.RawMessage
					jErr := json.Unmarshal([]byte(data), &d)
					if jErr != nil {
						wsNotifier.onError(jErr.Error() + fmt.Sprintf("=> %s => %s", e.Path, data))
						_error().Err(jErr).Msg("failed to unmarshal")
						return
					}

					//if !str.IsJSON(data) {
					//	errorMsg := "file data is not a json"
					//	wsNotifier.onError(errorMsg)
					//	_error().Str("data", data).Msg(errorMsg)
					//	return
					//}

					var dataDecoded interface{}
					err = json.Unmarshal([]byte(data), dataDecoded)

					if wsNotifier.OnNotificationRead != nil {
						// The programmer can change the values of the data before sending!
						// The programmer should take care of sending the data!
						wsNotifier.OnNotificationRead(h, dataDecoded, data)
					} else {
						// It will broadcast to all!
						if data != "" {
							h.BroadcastText(data)
						}
					}

					// Usually the data is already in Json Format!

					/*
						How do we decide to whom we are sending the message! Of course if we want that
						to be done...
						The notification can contain information about how to process it!
						Also we have the ability to change the content of it!
						Besides that we should filter the clients/receivers...
					*/

					// DO not delete instantly...because sometimes write can take some time?!
					time.AfterFunc(time.Second*10, func() {
						// log.Println("deleting file...", e.Path)
						if file.Exists(e.Path) {
							file.Unlink(e.Path)
						}
					})
				},
				Op: []fsnotify.Op{fsnotify.Create},
			})
		}

		// Start listening!
		fsNotifier.Start()

		for {
			time.Sleep(time.Second)
			if h.StopCalled.Get() {
				break
			}
		}

	})

	return wsNotifier
}

func (wsNotifier *WSNotifier) Start() *WSNotifier {
	wsNotifier.WSHub.Start()
	return wsNotifier
}

func (wsNotifier *WSNotifier) Stop() *WSNotifier {
	wsNotifier.WSHub.Stop()
	return wsNotifier
}
