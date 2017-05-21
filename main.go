package main

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/vpakhuchyi/device-smart-house/models"

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
	// go runTCPConfig()

	//-----TCP
	go runTCPServer()
	wg.Wait()
}

func switcher() bool {
	if state {
		state = false
		return state
	}
	return true
}

func SendRequest(turned bool, collFreq int, sendFreq int, conn net.Conn) {
	var req models.ConfigRequest
	var resp models.Response

	req.Turned = turned
	req.CollectFreq = collFreq
	req.SendFreq = sendFreq

	err := json.NewEncoder(conn).Encode(&req)
	if err != nil {
		log.Errorln(err)
	}

	err = json.NewDecoder(conn).Decode(&resp)
	if err != nil {
		log.Errorln(err)
	}

	log.Println("response: ", resp)
	log.Println("SendRequest: [done]")

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
		go requestHandler(conn)
	}
}

func runDBConnection() *redis.Client {
	var reconnect *time.Ticker
	dbClient = redis.New()

	// if err := dbClient.Connect(dbHost, dbPort); err != nil {
	// 	log.Errorln("Database: connection has failed: %s\n", err.Error())
	// 	return
	// }
	err := dbClient.Connect(dbHost, dbPort)
	for err != nil {
		log.Errorln("Database: connection has failed: %s\n", err)
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			err := dbClient.Connect(dbHost, dbPort)
			log.Errorln("Database: connection has failed: %s\n", err)
		}
	}
	//produce connection error; replaced to main.go
	// defer dbClient.Quit()
	return dbClient
}

func runViewServer() {
	r := mux.NewRouter()

	r.HandleFunc("/devices", staticHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./View/")))

	http.Handle("/", r)
	http.ListenAndServe(":8100", nil)

}

func runTCPConfig() {
	connType := "tcp"
	host := "localhost"
	port := "3000"
	var reconnect *time.Ticker
	ticker := time.NewTicker(time.Second * 5)

	defer reconnect.Stop()

	conn, err := net.Dial(connType, host+":"+port)
	for err != nil {
		log.Errorln(err)
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			conn, _ = net.Dial(connType, host+":"+port)
		}
	}

	for {
		select {
		case <-ticker.C:
			go SendRequest(switcher(), 1, 3, conn)

		}
	}

}
