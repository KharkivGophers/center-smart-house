package main

import (
	"github.com/KharkivGophers/center-smart-house/server/webSocket"
	"github.com/KharkivGophers/center-smart-house/server/myHTTP"
	"github.com/KharkivGophers/center-smart-house/common/models"
)

func main() {

	wg.Add(4)

	// db connection
	dbClient, err := RunDBConnection()
	CheckError("Main: RunDBConnection", err)
	defer dbClient.Close()

	// myHTTP connection with browser
	httpserver := myHTTP.NewHTTPServer(connHost, httpConnPort, models.DBURL{connHost, dbPort})
	go httpserver.RunDynamicServer()

	// web socket server
	server := webSocket.NewWebSocketServer(connHost,"2540", connHost,6379)
	go server.StartWebSocketServer()

	//-----TCP-Config
	go runConfigServer(configConnType, configHost, configPort)
	//-----TCP
	go runTCPServer()

	wg.Wait()
}
