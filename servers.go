package main

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"menteslibres.net/gosexy/redis"

)


func RunDBConnection() (*redis.Client, error) {
	client := redis.New()
	err := client.Connect(dbHost, dbPort)
	CheckError("RunDBConnection error: ", err)
	for err != nil {
		log.Errorln("1 Database: connection has failed:", err)
		time.Sleep(time.Second)
		err = client.Connect(dbHost, dbPort)
		if err == nil {
			continue
		}
		log.Errorln("err not nil")
	}

	return client, err
}

func runTCPServer() {
	var reconnect *time.Ticker

	ln, err := net.Listen(connType, connHost+":"+tcpConnPort)

	for err != nil {
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			ln, _ = net.Listen(connType, connHost+":"+tcpConnPort)
		}
		reconnect.Stop()
	}

	for {
		conn, err := ln.Accept()
		if CheckError("TCP conn Accept", err) == nil {
			go tcpDataHandler(conn)
		}
	}
}



