package main

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"menteslibres.net/gosexy/redis"
)

var (
	dbHost   = "127.0.0.1"
	dbPort   = uint(6379)
	dbClient *redis.Client

	connHost = "localhost"
	connPort = "3030"
	connType = "tcp"
)

func main() {
	dbClient = redis.New()
	if err := dbClient.Connect(dbHost, dbPort); err != nil {
		log.Fatalf("Database: connection has failed: %s\n", err.Error())
		return
	}
	defer dbClient.Quit()

	ln, _ := net.Listen(connType, connHost+":"+connPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorln(err)
		}
		go requestHandler(conn)
	}
}
