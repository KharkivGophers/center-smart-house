package main

import (
	"github.com/KharkivGophers/center-smart-house/servers/webSocket"
	"github.com/KharkivGophers/center-smart-house/servers/http"
	"github.com/KharkivGophers/center-smart-house/models"
	"time"
	"fmt"
	"sync"
	"github.com/KharkivGophers/center-smart-house/servers/tcp"
)

func main() {

	var wg sync.WaitGroup

	// http connection with browser
	httpServer := http.NewHTTPServer(models.Server{IP: centerIP, Port: httpConnPort}, dbServer)
	go httpServer.RunHTTPServer(wg)

	// web socket servers
	webSocketServer := webSocket.NewWebSocketServer(centerIP, fmt.Sprint(wsPort), centerIP, dbServer.Port)
	go webSocketServer.RunWebSocketServer(wg)

	// tcp config
	reconnect := time.NewTicker(time.Second * 1)
	messages := make(chan []string)
	tcpConfigServer := tcp.NewTCPConfigServer(models.Server{IP: centerIP, Port: tcpDevConfigPort}, dbServer, reconnect, messages)
	go tcpConfigServer.RunConfigServer(wg)

	// tcp data
	tcpDataServer := tcp.NewTCPDataServer(models.Server{IP: centerIP, Port: tcpDevDataPort}, dbServer, reconnect)
	go tcpDataServer.RunTCPServer(wg)

	wg.Wait()
}
