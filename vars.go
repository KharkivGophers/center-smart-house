package main

import (
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
	"menteslibres.net/gosexy/redis"
	"os"
	"strconv"
)

//var from main.go
var (
	//Database
	dbHost     = getEnvDbHost("REDIS_PORT_6379_TCP_ADDR")
	dbPort     = getEnvDbPort("REDIS_PORT_6379_TCP_PORT")
	dbClient   *redis.Client
	wsDBClient *redis.Client

	//General
	//connHost = "192.168.104.23"
	connHost = "localhost"
	//connHost = "10.4.25.73"
	//connHost = "192.168.104.60"

	//tcp conn with devices
	connType    = "tcp"
	tcpConnPort = "3030"

	//http connection
	httpConnPort = "8100"

	//for TCP config
	configConnType = "tcp"
	configHost     = "localhost"
	configPort     = "3000"
	configSubChan  = make(chan []string)

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

func getEnvDbPort(key string) uint {
	parsed64, _ := strconv.ParseUint(os.Getenv(key), 10, 64)
	port := uint(parsed64)
	if port == 0 {
		return uint(6379)
	}
	return port
}

func getEnvDbHost(key string) string {
	host := os.Getenv(key)
	if len(host) == 0 {
		return "127.0.0.1"
	}
	return host
}

