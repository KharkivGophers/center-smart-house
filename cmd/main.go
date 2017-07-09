package main

import (
	"github.com/KharkivGophers/center-smart-house/servers/webSocket"
	"github.com/KharkivGophers/center-smart-house/servers/http"
	. "github.com/KharkivGophers/center-smart-house/models"
	"github.com/KharkivGophers/center-smart-house/servers/tcp"
)

func main() {

	var controller RoutinesController

	httpServer := http.NewHTTPServer(Server{IP: centerIP, Port: httpConnPort}, dbServer, controller)
	go httpServer.Run()

	webSocketServer := webSocket.NewWebSocketServer(Server{IP: centerIP, Port: wsPort}, dbServer, controller)
	go webSocketServer.Run()

	tcpDevConfigServer := tcp.NewTCPDevConfigServerDefault(Server{IP: centerIP, Port: tcpDevConfigPort},
		dbServer, controller)
	go tcpDevConfigServer.Run()

	tcpDevDataServer := tcp.NewTCPDevDataServerDefault(Server{IP: centerIP, Port: tcpDevDataPort},
		dbServer, controller)
	go tcpDevDataServer.Run()

	controller.Wait()
}
