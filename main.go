package main

import "github.com/KharkivGophers/center-smart-house/server/webSocket"

func main() {

	wg.Add(4)

	// db connection
	dbClient, err := RunDBConnection()
	CheckError("Main: RunDBConnection", err)
	defer dbClient.Close()

	// http connection with browser
	go runDynamicServer()

	// web socket server
	server := webSocket.NewWebSocketServer(connHost,"2540", connHost,6379)
	go server.StartWebsocketServer()

	//-----TCP-Config
	go runConfigServer(configConnType, configHost, configPort)
	//-----TCP
	go runTCPServer()

	wg.Wait()
}
