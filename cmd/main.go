package main

import (
	"github.com/KharkivGophers/center-smart-house/server/webSocket"
	"github.com/KharkivGophers/center-smart-house/server/myHTTP"
	"github.com/KharkivGophers/center-smart-house/common/models"
	"github.com/KharkivGophers/center-smart-house/server/tcp/tcpConfig"
	"time"
	"github.com/KharkivGophers/center-smart-house/server/tcp/tcpData"
)

func main() {

	wg.Add(4)

	// myHTTP connection with browser
	httpserver := myHTTP.NewHTTPServer(connHost, httpConnPort, models.DBURL{connHost, dbPort})
	go httpserver.RunDynamicServer()

	// web socket server
	server := webSocket.NewWebSocketServer(connHost, "2540", connHost, 6379)
	go server.StartWebSocketServer()

	//-----TCP-Config
	reconnect := time.NewTicker(time.Second * 1)
	messages := make(chan []string)
	tcpConfigServer := tcpConfig.NewTCPConfigServer(connHost, configPort,
		models.DBURL{connHost, dbPort}, reconnect, messages)
	go tcpConfigServer.RunConfigServer()
	//-----TCP

	tcpDataserver := tcpData.NewTCPDataServer(connHost, tcpConnPort, models.DBURL{connHost, dbPort}, reconnect)
	go tcpDataserver.RunTCPServer()

	wg.Wait()
}
