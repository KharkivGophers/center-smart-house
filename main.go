package main

import (
	"net"
	log "github.com/Sirupsen/logrus"
	"menteslibres.net/gosexy/redis"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

var (
	dbHost   = "127.0.0.1"
	dbPort   = uint(6379)
	dbClient *redis.Client
	//for SS network
	connHost = "192.168.104.23"
	//connHost = "127.0.0.1"
	devConnPort         = "3030"
	devConnType         = "tcp"
	httpStaticConnPort  = "8100"
	httpDynamicConnPort = "8101"
	//for web socket
	wsConnPort = "2540"
)

func main() {
	log.Info("Server has started")
	//db connection
	dbClient = redis.New()
	if err := dbClient.Connect(dbHost, dbPort); err != nil {
		log.Fatalf("Database: connection has failed: %s\n", err.Error())
		return
	}



	defer dbClient.Quit()

	//http dynamic connection with browser
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/devices", httpDevHandler)
		srv := &http.Server{
			Handler: r,
			Addr:    "127.0.0.1" + ":" + httpDynamicConnPort,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		go log.Fatal(srv.ListenAndServe())
	}()

	//http static connection with browser
	go func() {
		r := mux.NewRouter()
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./View/")))
		srv := &http.Server{
			Handler: r,
			Addr:    "127.0.0.1" + ":" + httpStaticConnPort,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		go log.Fatal(srv.ListenAndServe())
	}()

	//http web socket connection
	go CloseWebsocket()
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/devWS", webSocketHandler)

		srv := &http.Server{
			Handler: r,
			Addr:    "127.0.0.1" + ":" + wsConnPort,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		go log.Fatal(srv.ListenAndServe())
	}()

	//tcp connection with devices
	ln, err := net.Listen(devConnType, "127.0.0.1"+":"+devConnPort)
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
