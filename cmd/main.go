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

	wg.Add(4)

	// http connection with browser
	httpServer := http.NewHTTPServer(models.Server{IP: centerIP, Port: httpConnPort}, dbServer)
	go httpServer.RunDynamicServer()

	// web socket servers
	webSockerServer := webSocket.NewWebSocketServer(centerIP, fmt.Sprint(wsPort), centerIP, dbServer.Port)
	go webSockerServer.StartWebSocketServer()

	// tcp config
	reconnect := time.NewTicker(time.Second * 1)
	messages := make(chan []string)

	tcpConfigServer := tcp.NewTCPConfigServer(models.Server{IP: centerIP, Port: tcpDevConfigPort}, dbServer, reconnect, messages)
	go tcpConfigServer.RunConfigServer()

	// tcp data
	tcpDataServer := tcp.NewTCPDataServer(models.Server{IP: centerIP, Port: tcpDevDataPort}, dbServer, reconnect)
	go tcpDataServer.RunTCPServer()

	wg.Wait()
}
