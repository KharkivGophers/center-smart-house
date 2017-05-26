package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"menteslibres.net/gosexy/redis"
)

var (
	dbHost   = "127.0.0.1"
	dbPort   = uint(6379)
	dbClient *redis.Client
	//for SS network
	//connHost = "192.168.104.23"
	connHost = "192.168.104.76"
	connPort = "3030"
	connType = "tcp"

	httpStaticConnPort  = "8100"
	httpDynamicConnPort = "8101"

	wsConnPort            = "2540"
	wsDBClient            *redis.Client
	roomIDForDevWSPublish = "devWS"

	wg    sync.WaitGroup
	state bool
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

	//websocket server
	go websocketServer()

	//-----TCP-Config
	go runConfigServer()
	//-----TCP
	go runTCPServer()
	wg.Wait()
}

//http web socket connection
func websocketServer() {
	go CloseWebsocket()
	go Subscribe(wsDBClient, roomIDForDevWSPublish)

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

		go tcpDataHandler(&conn)
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
	checkDBError(err)
	_, err = dbClient.SAdd(configInfo, "TurnedOn", "CollectFreq", "SendFreq")
	checkDBError(err)
	_, err = dbClient.ZAdd(configInfo+":"+"TurnedOn", defaultConfig.TurnedOn)
	checkDBError(err)
	_, err = dbClient.ZAdd(configInfo+":"+"CollectFreq", defaultConfig.CollectFreq)
	checkDBError(err)
	_, err = dbClient.ZAdd(configInfo+":"+"SendFreq", defaultConfig.SendFreq)
	checkDBError(err)

	// Send to Device
	err = json.NewEncoder(*conn).Encode(&defaultConfig)
	if err != nil {
		log.Errorln(err)
	}

	log.Warningln("default config has been sent")
	return nil
}

func checkDBError(err error) error {
	if err != nil {
		log.Errorln("Failed operation with DB", err)
		return err
	}
	return nil
}
