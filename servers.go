package main

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"menteslibres.net/gosexy/redis"

	log "github.com/Sirupsen/logrus"
)

//http web socket connection
func websocketServer() {
	go CloseWebsocket()
	go WSSubscribe(wsDBClient, roomIDForDevWSPublish, subWSChannel)

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}", webSocketHandler)
	r.HandleFunc("/devWS", webSocketHandler)

	srv := &http.Server{
		Handler: r,
		Addr:    connHost + ":" + wsConnPort,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go log.Fatal(srv.ListenAndServe())
}

//http dynamic connection with browser
func runDynamicServer() {
	r := mux.NewRouter()
	r.HandleFunc("/devices", getDevicesHandler).Methods("GET")
	r.HandleFunc("/devices/{id}/data", getDevDataHandler).Methods("GET")
	r.HandleFunc("/devices/{id}/config", getDevConfigHandler).Methods("GET")

	r.HandleFunc("/devices/{id}/config", patchDevConfigHandler).Methods("PATCH")

	//provide static html pages
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./view/")))

	srv := &http.Server{
		Handler: r,
		Addr:    connHost + ":" + httpConnPort,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//CORS provides Cross-Origin Resource Sharing middleware
	http.ListenAndServe(connHost+":"+httpConnPort, handlers.CORS()(r))

	go log.Fatal(srv.ListenAndServe())
}

func runDBConnection() *redis.Client {
	var reconnect *time.Ticker
	dbClient = redis.New()

	err := dbClient.Connect(dbHost, dbPort)
	checkError("runDBConnection error: ", err)
	for err != nil {
		log.Errorln("1 Database: connection has failed:", err)
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			err = dbClient.Connect(dbHost, dbPort)
			if err == nil {
				continue
			}
			log.Errorln("err not nil")
		}
		//make sleep

		// for dbClient == err {
		// 	reconnect = time.NewTicker(time.Second * 1)
		// 	for range reconnect.C {
		// 		err := dbClient.Connect(dbHost, dbPort)
		// 		log.Errorln("Database: connection has failed: %s\n", err)
		// 	}
		// }
		// reconnect.Stop()
	}
	return dbClient
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
		checkError("TCP conn Accept", err)

		go tcpDataHandler(&conn)
	}
}

func runConfigServer(connType string, host string, port string) {

	messages := make(chan []string)
	var dbClient *redis.Client
	var reconnect *time.Ticker
	var pool ConectionPool
	pool.init()

	go func() {
		var reconnect *time.Ticker
		dbClient = runDBConnection()
		for dbClient == nil {
			reconnect = time.NewTicker(time.Second * 1)
			for range reconnect.C {
				err := dbClient.Connect(dbHost, dbPort)
				log.Errorln("Database: connection has failed: %s\n", err)
			}
			return
		}
	}()
	defer dbClient.Close()

	ln, err := net.Listen(connType, host+":"+port)

	for err != nil {
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			ln, _ = net.Listen(connType, connHost+":"+tcpConnPort)
		}
		reconnect.Stop()
	}
	go configSubscribe(dbClient, "configChan", messages, &pool)

	for {
		conn, err := ln.Accept()
		checkError("TCP config conn Accept", err)
		go sendDefaultConfiguration(&conn, &pool)
	}
}

func sendNewConfiguration(config DevConfig, pool *ConectionPool) {
	var resp Response
	conn := pool.getConn(config.MAC)

	err := json.NewEncoder(*conn).Encode(&config)
	checkError("sendNewConfiguration JSON Encod", err)
	err = json.NewDecoder(*conn).Decode(&resp)
	checkError("sendNewConfiguration JSON Decod", err)
}

func sendDefaultConfiguration(conn *net.Conn, pool *ConectionPool) {
	// Send Default Configuration to Device
	var req Request

	err := json.NewDecoder(*conn).Decode(&req)
	checkError("sendDefaultConfiguration JSON Decod", err)
	pool.addConn(conn, req.Meta.MAC)

	Time := time.Now().UnixNano() / int64(time.Millisecond)
	configInfo := req.Meta.MAC + ":" + "params" // key

	// Save default configuration to DB
	defaultConfig := DevConfig{
		TurnedOn:    true,
		CollectFreq: 1,
		SendFreq:    5,
	}

	dbClient.SAdd("Config", configInfo)

	_, err = dbClient.HMSet(req.Meta.MAC, "ConfigTime", Time)
	checkError("DB error 1", err)
	_, err = dbClient.SAdd(configInfo, "TurnedOn", "CollectFreq", "SendFreq")
	checkError("DB error 2", err)
	_, err = dbClient.ZAdd(configInfo+":"+"TurnedOn", Time, defaultConfig.TurnedOn)
	checkError("DB error 3", err)
	_, err = dbClient.ZAdd(configInfo+":"+"CollectFreq", Time, defaultConfig.CollectFreq)
	checkError("DB error 4", err)
	_, err = dbClient.ZAdd(configInfo+":"+"SendFreq", Time, defaultConfig.SendFreq)
	checkError("DB error 5", err)

	// Send to Device
	err = json.NewEncoder(*conn).Encode(&defaultConfig)
	checkError("sendDefaultConfiguration JSON enc", err)
	log.Warningln("default config has been sent")
}
