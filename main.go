package main

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"menteslibres.net/gosexy/redis"
)
func main() {
	wg.Add(4)
	//db connection
	go func() {
		dbClient = runDBConnection()
	}()
	defer dbClient.Close()

	//http connection with browser

	go runStaticServer()
	go runDynamicServer()

	//websocket server
	go websocketServer()

	//-----TCP-Config
	go runConfigServer(configConnType, configHost, configPort)
	//-----TCP
	go runTCPServer()
	wg.Wait()
}

//http web socket connection
func websocketServer() {
	go CloseWebsocket()
	go WSSubscribe(wsDBClient, roomIDForDevWSPublish, subWSChannel)

	r := mux.NewRouter()
	r.HandleFunc("/devices/{type}/{name}/{mac}", webSocketHandler)
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
	r.HandleFunc("/devices/{id}/config", postDevConfigHandler).Methods("POST")
	r.HandleFunc("/devices/{type}/{name}/{mac}", getDevDataHandler).Methods("GET")
	r.HandleFunc("/devices", getDevicesHandler)
	srv := &http.Server{
		Handler: r,
		Addr:    connHost + ":" + httpDynamicConnPort,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//CORS provides Cross-Origin Resource Sharing middleware
	http.ListenAndServe(connHost+":"+httpDynamicConnPort, handlers.CORS()(r))

	go log.Fatal(srv.ListenAndServe())
}

//http static connection with browser
func runStaticServer() {
	r := mux.NewRouter()
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./View/")))
	srv := &http.Server{
		Handler: r,
		Addr:    connHost + ":" + httpStaticConnPort,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go log.Fatal(srv.ListenAndServe())

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

func runTCPServer() {
	var i int
	var reconnect *time.Ticker

	ln, err := net.Listen(connType, connHost+":"+tcpConnPort)

	for err != nil {
		reconnect = time.NewTicker(time.Second * 1)
		for range reconnect.C {
			ln, _ = net.Listen(connType, connHost+":"+tcpConnPort)
			i++
			log.Println("2: net.Listen for range", i)
		}
		reconnect.Stop()
	}

	for {
		conn, err := ln.Accept()
		CheckError("TCP conn Accept", err)

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
		dbClient = runDBConnection()
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
		CheckError("TCP config conn Accept", err)
		go sendDefaultConfiguration(&conn, &pool)
	}

}

func sendNewConfiguration(config Config, pool *ConectionPool) {
	var resp Response
	conn := pool.getConn(config.MAC)

	err := json.NewEncoder(*conn).Encode(&config)
	CheckError("sendNewConfiguration JSON Encod", err)
	err = json.NewDecoder(*conn).Decode(&resp)
	CheckError("sendNewConfiguration JSON Decod", err)
}

func sendDefaultConfiguration(conn *net.Conn, pool *ConectionPool) {
	// Send Default Configuration to Device
	var req Request

	err := json.NewDecoder(*conn).Decode(&req)
	CheckError("sendDefaultConfiguration JSON Decod", err)
	pool.addConn(conn, req.Meta.MAC)

	Time := time.Now().UnixNano() / int64(time.Millisecond)
	configInfo := req.Meta.MAC + ":" + "params" // key

	// Save default configuration to DB
	defaultConfig := Config{
		TurnedOn:    true,
		CollectFreq: 5,
		SendFreq:    10,
	}

	dbClient.SAdd("Config", configInfo)

	_, err = dbClient.HMSet(req.Meta.MAC, "ConfigTime", Time)
	CheckError("DB error", err)
	_, err = dbClient.SAdd(configInfo, "TurnedOn", "CollectFreq", "SendFreq")
	CheckError("DB error", err)
	_, err = dbClient.ZAdd(configInfo+":"+"TurnedOn", defaultConfig.TurnedOn)
	CheckError("DB error", err)
	_, err = dbClient.ZAdd(configInfo+":"+"CollectFreq", defaultConfig.CollectFreq)
	CheckError("DB error", err)
	_, err = dbClient.ZAdd(configInfo+":"+"SendFreq", defaultConfig.SendFreq)
	CheckError("DB error", err)

	// Send to Device
	err = json.NewEncoder(*conn).Encode(&defaultConfig)
	CheckError("sendDefaultConfiguration JSON enc", err)
	log.Warningln("default config has been sent")
}