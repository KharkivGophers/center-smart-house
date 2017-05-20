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
	go runDBConnection()
	log.Println("DB connected 1")
	//-----TCP
	go runTCPServer()
	log.Println("TCP server connected 2")
	//-----View
	go runViewServer()
	log.Println("View server connected 3")
	//-----TCP-Config
	//	go runTCPConfig()
	log.Println("TCP config server connected 4")

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
	reconnect := time.NewTicker(time.Second * 1)
	defer reconnect.Stop()

	ln, err := net.Listen(connType, connHost+":"+connPort)

	for err != nil {
		log.Println("2: net.Listen err != nil")
		for range reconnect.C {
			ln, _ = net.Listen(connType, connHost+":"+connPort)
			i++
			log.Println("2: net.Listen for range", i)
		}
	}
	log.Println("2: net.Listen connected after loop")

	for {
		log.Println("2: ln.Accept inf for")
		conn, err := ln.Accept()
		if err != nil {
			log.Println("2: ln.Accept inf for err!=nil")
			log.Errorln(err)
			continue
		}
		log.Println("2: before requestHandler")
		go requestHandler(conn)
		log.Println("2: after requestHandler")
	}
}

func runDBConnection() {
	reconnect := time.NewTicker(time.Second * 1)
	dbClient = redis.New()

	// if err := dbClient.Connect(dbHost, dbPort); err != nil {
	// 	log.Errorln("Database: connection has failed: %s\n", err.Error())
	// 	return
	// }
	err := dbClient.Connect(dbHost, dbPort)
	for err != nil {
		for range reconnect.C {
			err := dbClient.Connect(dbHost, dbPort)
			log.Errorln(err)
		}
	}
	//produce connection error
	// defer dbClient.Quit()
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

	ticker := time.NewTicker(time.Second * 5)
	reconnect := time.NewTicker(time.Second * 1)
	defer reconnect.Stop()

	conn, err := net.Dial(connType, host+":"+port)
	log.Warningln("Config Dial 1st conn")
	for err != nil {
		for range reconnect.C {
			conn, _ = net.Dial(connType, host+":"+port)
		}
	}
	log.Warningln("Config Dial after loop conn")
	go func() {
		for {
			log.Warningln("SenRequest for intro")
			select {
			case <-ticker.C:
				log.Warningln("SenRequest case intro")
				go SendRequest(switcher(), 1, 3, conn)
				log.Warningln("SenRequest case out")

			}
		}
	}()

}
