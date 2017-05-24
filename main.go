package main

import (
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"sync"

	"encoding/json"

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
	//go runViewServer()

	//-----TCP-Config
	go runConfigServer()

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

func runConfigServer() {
	connType := "tcp"
	host := "localhost"
	port := "3000"
	var reconnect *time.Ticker
	var pool ConectionPool
	pool.init()

	ln, err := net.Listen(connType, host+":"+port)

	for err != nil {
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			ln, _ = net.Listen(connType, connHost+":"+connPort)
		}
		reconnect.Stop()
	}
	//connection with DB; listens for PubSubs
	// go redisPubSubConn() {
	// for {
	// take new config from DB and maps it to COnfig structure
	// run sendNewConfiguration() with this Congif struct as param
	// if publish {
	// 	sendNewConfiguration()
	// }
	// }
	// }

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorln("ln.Accept", err)
		}
		//save conn to the map[MAC]net.conn
		go sendDefaultConfiguration(&conn, &pool)
	}

}

func sendNewConfiguration(config Config, pool *ConectionPool) error {
	var resp Response
	conn := pool.getConn(config.MAC)

	err := json.NewEncoder(*conn).Encode(&config)
	if err != nil {
		log.Errorln("sendNewConfiguration JSON Encod", err)
		return err
	}

	err = json.NewDecoder(*conn).Decode(&resp)
	if err != nil {
		log.Errorln("sendNewConfiguration JSON Decod", err)
		return err
	}

	return nil
}

func sendDefaultConfiguration(conn *net.Conn, pool *ConectionPool) error {
	// Send Default Configuration to Device
	var req Request
	log.Warningln("received default config request")
	err := json.NewDecoder(*conn).Decode(&req)
	if err != nil {
		log.Errorln("sendDefaultConfiguration JSON Decod", err)
		return err
	}
	pool.addConn(conn, req.Meta.MAC)

	//for now we just send hard-coded config
	//but we have to send default config from DB
	configuration := Config{
		State:       true,
		CollectFreq: 1,
		SendFreq:    10,
	}
	err = json.NewEncoder(*conn).Encode(&configuration)
	if err != nil {
		log.Errorln("sendDefaultConfiguration JSON Encod", err)
		return err
	}
	log.Warningln("default config has been sent")
	return nil

}

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
