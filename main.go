
package main

import (
	"net"
	log "github.com/logrus"
	"menteslibres.net/gosexy/redis"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"github.com/gorilla/websocket"
)

var (
	dbHost   = "127.0.0.1"
	dbPort   = uint(6379)
	dbClient *redis.Client
	//for SS network
	connHost = "192.168.104.23"
	//connHost = "127.0.0.1"
	devConnPort = "3030"
	devConnType = "tcp"
	httpConnPort = "8100"
)




func main() {
	//db connection
	dbClient = redis.New()
	if err := dbClient.Connect(dbHost, dbPort); err != nil {
		log.Fatalf("Database: connection has failed: %s\n", err.Error())
		return
	}
	defer dbClient.Quit()

	//http connection with browser
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/devWS", webSocketHandler)
		r.HandleFunc("/devices", httpDevHandler)
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./View/")))

		srv := &http.Server{
			Handler: r,
			Addr:    "127.0.0.1" + ":" + httpConnPort,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		go log.Fatal(srv.ListenAndServe())
	} ()

	//tcp connection with devices
	ln, err := net.Listen(devConnType, connHost + ":" + devConnPort)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorln(err)
			continue
		}

		go requestHandler(conn)
	}
}