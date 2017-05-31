package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"menteslibres.net/gosexy/redis"
)

//var from main.go
var (
	//Data base
	dbHost     = "127.0.0.1"
	dbPort     = uint(6379)
	dbClient   *redis.Client
	wsDBClient *redis.Client

	//General
	//connHost = "192.168.104.23"
	//connHost = "192.168.104.76"
	connHost = "localhost"

	//tcp conn with devices
	connType    = "tcp"
	tcpConnPort = "3030"

	//http connection
	httpConnPort  = "8100"

	//for TCP config
	configConnType = "tcp"
	configHost     = "localhost"
	configPort     = "3000"

	//Web-socket connections
	wsConnPort            = "2540"
	roomIDForDevWSPublish = "devWS"
	subWSChannel          = make(chan []string)

	wg    sync.WaitGroup
	state bool

	//var from handler.go
	//This var needed to websocket for using
	connChanal  = make(chan *websocket.Conn)
	stopCloseWS = make(chan string)
	stopSub     = make(chan bool)
	mapConn     = make(map[string]*listConnection)

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Host == connHost+":"+wsConnPort {
				return true
			}
			return false
		},
	}
)
