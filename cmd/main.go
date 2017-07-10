package main

import (
	"github.com/KharkivGophers/center-smart-house/servers/webSocket"
	"github.com/KharkivGophers/center-smart-house/servers/http"
	. "github.com/KharkivGophers/center-smart-house/models"
	"time"
	"github.com/KharkivGophers/center-smart-house/servers/tcp"
	"github.com/KharkivGophers/center-smart-house/dao"
)

func main() {

	var dbClient dao.DbClient = &dao.RedisClient{}
	dbClient.SetDBServer(dbServer)
	var controller RoutinesController

	httpServer := http.NewHTTPServer(Server{IP: centerIP, Port: httpConnPort}, controller, dbClient)
	go httpServer.Run()

	webSocketServer := webSocket.NewWebSocketServer(Server{IP: centerIP, Port: wsPort}, controller,dbClient)
	go webSocketServer.Run()

	tcpDevConfigServer := tcp.NewTCPDevConfigServerDefault(Server{IP: centerIP, Port: tcpDevConfigPort},
		 controller, dbClient)
	go tcpDevConfigServer.Run()

	reconnect := time.NewTicker(time.Second * 1)
	tcpDevDataServer := tcp.NewTCPDevDataServer(Server{IP: centerIP, Port: tcpDevDataPort},
		 reconnect, controller,dbClient)
	go tcpDevDataServer.Run()

	controller.Wait()
}
