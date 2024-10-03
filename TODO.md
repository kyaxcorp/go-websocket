1. add statistics for each client
2. See with broadcasting for multiple clients... we need here ASYNC, because sometimes it can be a problem writing data
for some long distanced clients... we should check... if the writer is somehow delayed or taking too much time writing 
   to these kind of clients....
async meaning we should create goroutines, but anyway... if we have pipes... writePipes for the clients, and they 
   are autonomous, then everything is alright! Just check there if everything is ok!
   

3. Check http recover and connection_id generation... for client, when the session dies, the http recovers, and
we have an incorrect connection id attached for the client!...
4. add termination context
5. disconnect connections which are not authorized... -> in this case we simply use middleware before upgrading to websocket!