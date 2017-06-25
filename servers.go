package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"menteslibres.net/gosexy/redis"
	"github.com/gorilla/websocket"
)

//http web socket connection
func websocketServer() {
	var (
		connChanal   = make(chan *websocket.Conn)
		stopCloseWS  = make(chan string)
		stopSub      = make(chan bool)
		subWSChannel = make(chan []string)
	)

	wsDBClient, err := runDBConnection()
	checkError("webSocket: runDBConnection", err)

	go CloseWebsocket(connChanal, stopCloseWS)
	go WSSubscribe(wsDBClient, roomIDForDevWSPublish, subWSChannel, connChanal, stopSub)

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}", webSocketHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         connHost + ":" + wsConnPort,
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

func runDBConnection() (*redis.Client, error) {
	client := redis.New()
	err := client.Connect(dbHost, dbPort)
	checkError("runDBConnection error: ", err)
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
		if checkError("TCP conn Accept", err) == nil {
			go tcpDataHandler(conn)
		}
	}
}

func runConfigServer(connType string, host string, port string) {

	messages := make(chan []string)
	var dbClient *redis.Client
	var reconnect *time.Ticker
	var pool ConnectionPool
	pool.init()
	dbClient, err := runDBConnection()
	checkError("runConfigServer: runDBConnection", err)
	go func() {
		var reconnect *time.Ticker
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
		go sendDefaultConfiguration(conn, &pool)
	}
}

func sendNewConfiguration(config DevConfig, pool *ConnectionPool) {

	connection := pool.getConn(config.MAC)
	if connection == nil {
		log.Error("Has not connection with mac:config.MAC  in connectionPool")
		return
	}

	// log.Println("mac in pool sendNewConfig", config.MAC)
	err := json.NewEncoder(connection).Encode(&config)

	if err != nil {
		pool.removeConn(config.MAC)
	}
	checkError("sendNewConfig", err)
}

func sendDefaultConfiguration(conn net.Conn, pool *ConnectionPool) {
	// Send Default Configuration to Device

	dbClient, _ := runDBConnection()
	var req Request
	var config DevConfig
	err := json.NewDecoder(conn).Decode(&req)
	checkError("sendDefaultConfiguration JSON Decod", err)
	log.Println(req)

	pool.addConn(conn, req.Meta.MAC)
	// log.Println("MAC in pool", req.Meta.MAC)

	configInfo := req.Meta.MAC + ":" + "config" // key

	if ok, _ := dbClient.Exists(configInfo); ok {
		state, err := dbClient.HMGet(configInfo, "TurnedOn")
		checkError("Get from DB error1: TurnedOn ", err)

		if strings.Join(state, " ") != "" {
			// log.Warningln("New Config")
			sendFreq, _ := dbClient.HMGet(configInfo, "SendFreq")
			checkError("Get from DB error2: SendFreq ", err)
			collectFreq, _ := dbClient.HMGet(configInfo, "CollectFreq")
			checkError("Get from DB error3: CollectFreq ", err)
			streamOn, _ := dbClient.HMGet(configInfo, "StreamOn")
			checkError("Get from DB error4: StreamOn ", err)

			stateBool, _ := strconv.ParseBool(strings.Join(state, " "))
			sendFreqInt, _ := strconv.Atoi(strings.Join(sendFreq, " "))
			collectFreqInt, _ := strconv.Atoi(strings.Join(collectFreq, " "))
			streamOnBool, _ := strconv.ParseBool(strings.Join(streamOn, " "))

			config = DevConfig{
				TurnedOn:    stateBool,
				CollectFreq: int64(collectFreqInt),
				SendFreq:    int64(sendFreqInt),
				StreamOn:    streamOnBool,
			}
			log.Println("Old Device with MAC: ", req.Meta.MAC, "detected.")
			log.Println("Configuration from DB: ", state, sendFreq, collectFreq)
		}
	} else {
		log.Warningln("New Device with MAC: ", req.Meta.MAC, "detected.")
		log.Warningln("Default Config will be sent.")
		config = DevConfig{
			TurnedOn:    true,
			StreamOn:    true,
			CollectFreq: 1000,
			SendFreq:    5000,
		}

		// Save default configuration to DB
		_, err = dbClient.HMSet(configInfo, "TurnedOn", config.TurnedOn)
		checkError("DB error1: TurnedOn", err)
		_, err = dbClient.HMSet(configInfo, "CollectFreq", config.CollectFreq)
		checkError("DB error2: CollectFreq", err)
		_, err = dbClient.HMSet(configInfo, "SendFreq", config.SendFreq)
		checkError("DB error3: SendFreq", err)
		_, err = dbClient.HMSet(configInfo, "StreamOn", config.StreamOn)
		checkError("DB error4: StreamOn", err)
	}

	err = json.NewEncoder(conn).Encode(&config)
	checkError("sendDefaultConfiguration JSON enc", err)
	log.Warningln("Configuration has been successfully sent")
}
