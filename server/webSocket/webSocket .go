package webSocket

import (
	"time"
	"net/http"
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"menteslibres.net/gosexy/redis"
	"github.com/gorilla/websocket"
	"strings"
	"github.com/KharkivGophers/center-smart-house/dao"
)

type WebSocketServer struct {
	host string
	port            string
	roomIDForDevWSPublish string
	connChanal   chan *websocket.Conn
	stopCloseWS  chan string
	stopSub      chan bool
	subWSChannel chan []string
	mapConn      map[string]*listConnection
	upgrader     websocket.Upgrader
}

func NewWebSocketServer(host, port string) *WebSocketServer {
	var (
		roomIDForDevWSPublish = "devWS"
		connChanal   = make(chan *websocket.Conn)
		stopCloseWS  = make(chan string)
		stopSub      = make(chan bool)
		subWSChannel = make(chan []string)
		mapConn      = make(map[string]*listConnection)
		upgrader     = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				if r.Host == host+":"+port {
					return true
				}
				return true
			},
		}
	)
	return &WebSocketServer{host, port, roomIDForDevWSPublish ,
		connChanal, stopCloseWS, stopSub,
		subWSChannel, mapConn, upgrader}
}

//http web socket connection
func (server *WebSocketServer) websocketServer() {

	var dbWorker dao.DbWorker
	myRedis, err := dao.MyRedis{}.RunDBConnection()
	CheckError("webSocket: runDBConnection", err)

	go CloseWebsocket(server.connChanal, server.stopCloseWS)
	go WSSubscribe(myRedis, server.roomIDForDevWSPublish, server.subWSChannel, server.connChanal, server.stopSub)

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}", webSocketHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         server.host + ":" + server.port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go log.Fatal(srv.ListenAndServe())
}

//-------------------WEB Socket--------------------------------------------------------------------------------------------
func webSocketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	//http://..../device/type/name/mac
	uri := strings.Split(r.URL.String(), "/")

	if _, ok := mapConn[uri[2]]; !ok {
		mapConn[uri[2]] = new(listConnection)
	}
	mapConn[uri[2]].Add(conn)
}

/**
Delete connections in mapConn
*/
func CloseWebsocket(connChan chan *websocket.Conn, stopCloseWS chan string) {
	for {
		select {
		case connAddres := <-connChan:
			for _, val := range mapConn {
				if ok := val.Remove(connAddres); ok {
					break
				}
			}
		case <-stopCloseWS:
			log.Info("CloseWebsocket closed")
			return
		}
	}
}

/*
Listens changes in database. If they have, we will send to all websocket which working with them.
*/
func WSSubscribe(client *redis.Client, roomID string, chanForTheSub chan []string, connChan chan *websocket.Conn, stopWSSub chan bool) {
	subscribe(client, roomID, chanForTheSub)
	for {
		select {
		case msg := <-chanForTheSub:
			if msg[0] == "message" {
				go checkAndSendInfoToWSClient(msg, connChan)
			}
		case <-stopWSSub:
			log.Info("WSSubscribe closed")
			return
		}
	}
}

//We are check mac in our mapConnections.
// If we have mac in the map we will send message to all connections.
// Else we do nothing
func checkAndSendInfoToWSClient(msg []string, connChan chan *websocket.Conn) {
	r := new(Request)
	err := json.Unmarshal([]byte(msg[2]), &r)
	if checkError("checkAndSendInfoToWSClient", err) != nil {
		return
	}
	if _, ok := mapConn[r.Meta.MAC]; ok {
		sendInfoToWSClient(r.Meta.MAC, msg[2], connChan)
		return
	}
	log.Infof("mapConn dont have this MAC: %v. Len map is %v", r.Meta.MAC, len(mapConn))
}

//Send message to all connections which we have in map, and which pertain to mac
func sendInfoToWSClient(mac, message string, connChan chan *websocket.Conn) {
	mapConn[mac].Lock()
	for _, val := range mapConn[mac].connections {
		err := val.WriteMessage(1, []byte(message))
		if err != nil {
			log.Errorf("Connection %v closed", val.RemoteAddr())
			go getToChanal(val, connChan)
		}
	}
	mapConn[mac].Unlock()
}

func getToChanal(conn *websocket.Conn, connChan chan *websocket.Conn) {
	connChan <- conn
}

func publishWS(req Request) {
	pubReq, err := json.Marshal(req)
	checkError("Marshal for publish.", err)
	go publishMessage(pubReq, roomIDForDevWSPublish)
}
