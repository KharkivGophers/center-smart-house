package main

import (
	"github.com/KharkivGophers/center-smart-house/servers/webSocket"
	"github.com/KharkivGophers/center-smart-house/servers/http"
	. "github.com/KharkivGophers/center-smart-house/models"
	"time"
	"github.com/KharkivGophers/center-smart-house/servers/tcp"
)

func main() {

	var controller RoutinesController

	httpServer := http.NewHTTPServer(Server{IP: centerIP, Port: httpConnPort}, dbServer, controller)
	go httpServer.Run()

	webSocketServer := webSocket.NewWebSocketServer(Server{IP: centerIP, Port: wsPort}, dbServer, controller)
	go webSocketServer.Run()

	reconnect := time.NewTicker(time.Second * 1)
	messages := make(chan []string)
	tcpDevConfigServer := tcp.NewTCPDevConfigServer(Server{IP: centerIP, Port: tcpDevConfigPort}, dbServer, reconnect, messages, controller)
	go tcpDevConfigServer.Run()

	tcpDevDataServer := tcp.NewTCPDevDataServer(Server{IP: centerIP, Port: tcpDevDataPort}, dbServer, reconnect, controller)
	go tcpDevDataServer.Run()

	controller.Wait()
}
