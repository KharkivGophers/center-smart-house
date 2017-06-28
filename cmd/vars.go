package main

import (

	"sync"
	"os"
	"strconv"
)

//var from main.go
var (
	//Database
	dbHost     = getEnvDbHost("REDIS_PORT_6379_TCP_ADDR")
	dbPort     = getEnvDbPort("REDIS_PORT_6379_TCP_PORT")
	//General
	connHost = "0.0.0.0"

	//tcp conn with devices
	connType    = "tcp"
	tcpConnPort = "3030"

	//myHTTP connection
	httpConnPort = "8100"

	//for TCP config
	configConnType = "tcp"
	configHost     = "0.0.0.0"
	configPort     = "3000"

	//Web-socket connections
	wsConnPort            = "2540"
	roomIDForDevWSPublish = "devWS"


	wg    sync.WaitGroup


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

