package main

import (
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"sync"

	"menteslibres.net/gosexy/redis"
)

var (
	dbHost   = "127.0.0.1"
	dbPort   = uint(6379)
	dbClient *redis.Client

	connHost = "localhost"
	connPort = "3030"
	connType = "tcp"

	wg    sync.WaitGroup
	state bool
)

func main() {
	wg.Add(4)

	//-----DB
	go func() {
		dbClient = runDBConnection()
	}()
	defer dbClient.Close()

	//-----View
	go runViewServer()

	//-----TCP-Config
	// go runConfigServer()

	//-----TCP
	go runTCPServer()
	wg.Wait()
}

func runTCPServer() {
	var i int
	var reconnect *time.Ticker

	ln, err := net.Listen(connType, connHost+":"+connPort)

	for err != nil {
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			ln, _ = net.Listen(connType, connHost+":"+connPort)
			i++
			log.Println("2: net.Listen for range", i)
		}
		reconnect.Stop()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorln("ln.Accept", err)
		}
		go requestHandler(&conn)
	}
}

// func runConfigServer() {
// 	connType := "tcp"
// 	host := "localhost"
// 	port := "3000"
// 	var reconnect *time.Ticker

// 	ln, err := net.Listen(connType, host+":"+port)

// 	for err != nil {
// 		reconnect = time.NewTicker(time.Second * 1)
// 		for range reconnect.C {
// 			ln, _ = net.Listen(connType, connHost+":"+connPort)
// 		}
// 		reconnect.Stop()
// 	}
// 	//перекинуть в мейн девайса
// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			log.Errorln("ln.Accept", err)
// 		}
// 		go requestHandler(&conn)
// 	}

// }

func runDBConnection() *redis.Client {
	var reconnect *time.Ticker
	dbClient = redis.New()

	err := dbClient.Connect(dbHost, dbPort)
	for err != nil {
		log.Errorln("Database: connection has failed: %s\n", err)
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			err := dbClient.Connect(dbHost, dbPort)
			log.Errorln("Database: connection has failed: %s\n", err)
		}
	}

	return dbClient
}

func runViewServer() {
	r := mux.NewRouter()

	r.HandleFunc("/devices", staticHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./View/")))

	http.Handle("/", r)
	http.ListenAndServe(":8100", nil)

}
