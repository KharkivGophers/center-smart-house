package main

import (
	"github.com/KharkivGophers/center-smart-house/servers/webSocket"
	"github.com/KharkivGophers/center-smart-house/servers/http"
	. "github.com/KharkivGophers/center-smart-house/models"
	"time"
	"fmt"
	"github.com/KharkivGophers/center-smart-house/servers/tcp"
)

func main() {

	var control Control

	// http connection with browser
	httpServer := http.NewHTTPServer(Server{IP: centerIP, Port: httpConnPort}, dbServer)
	go httpServer.RunHTTPServer(control)

	// web socket servers
	webSocketServer := webSocket.NewWebSocketServer(centerIP, fmt.Sprint(wsPort), centerIP, dbServer.Port)
	go webSocketServer.RunWebSocketServer(control)

	// tcp config
	reconnect := time.NewTicker(time.Second * 1)
	messages := make(chan []string)
	tcpConfigServer := tcp.NewTCPConfigServer(Server{IP: centerIP, Port: tcpDevConfigPort}, dbServer, reconnect, messages)
	go tcpConfigServer.RunConfigServer(control)

	// tcp data
	tcpDataServer := tcp.NewTCPDataServer(Server{IP: centerIP, Port: tcpDevDataPort}, dbServer, reconnect)
	go tcpDataServer.RunTCPServer(control)

	control.Wait()
}
