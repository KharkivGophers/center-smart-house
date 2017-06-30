package main

import (
	"github.com/KharkivGophers/center-smart-house/servers/webSocket"
	"github.com/KharkivGophers/center-smart-house/servers/http"
	"github.com/KharkivGophers/center-smart-house/models"
	"github.com/KharkivGophers/center-smart-house/servers/tcp/config"
	"time"
	"github.com/KharkivGophers/center-smart-house/servers/tcp/data"
	"fmt"
	"sync"
)

func main() {

	var wg sync.WaitGroup

	wg.Add(4)

	// http connection with browser
	httpServer := http.NewHTTPServer(centerIP, httpConnPort, dbServer)
	go httpServer.RunDynamicServer()

	// web socket servers
	webSockerServer := webSocket.NewWebSocketServer(centerIP, fmt.Sprint(wsPort), centerIP, dbServer.Port)
	go webSockerServer.StartWebSocketServer()

	// tcp config
	reconnect := time.NewTicker(time.Second * 1)
	messages := make(chan []string)

	tcpConfigServer := config.NewTCPConfigServer(models.Server{IP: centerIP, Port: tcpDevConfigPort}, dbServer, reconnect, messages)
	go tcpConfigServer.RunConfigServer()

	// tcp data
	tcpDataServer := data.NewTCPDataServer(models.Server{IP: centerIP, Port: tcpDevDataPort}, dbServer, reconnect)
	go tcpDataServer.RunTCPServer()

	wg.Wait()
}
