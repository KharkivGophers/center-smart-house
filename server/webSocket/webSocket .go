package webSocket

import (
	"time"
	"net/http"
	"encoding/json"


	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"strings"
	"github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/server/common"

	"github.com/KharkivGophers/device-smart-house/models"
)

type WebSocketServer struct {
	host string
	port            string
	dbPort uint
	roomIDForDevWSPublish string
	connChanal   chan *websocket.Conn
	stopCloseWS  chan string
	stopSub      chan bool
	subWSChannel chan []string
	mapConn      map[string]*listConnection
	upgrader     websocket.Upgrader
}

func NewWebSocketServer(host, port string, dbPort uint) *WebSocketServer {
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
	return &WebSocketServer{host, port, dbPort, roomIDForDevWSPublish ,
		connChanal, stopCloseWS, stopSub,
		subWSChannel, mapConn, upgrader}
}

//http web socket connection
func (server *WebSocketServer) StartWebsocketServer() {

	myRedis, err := dao.MyRedis{Host:server.host, Port:server.dbPort}.RunDBConnection()
	CheckError("webSocket: runDBConnection", err)

	go server.CloseWebsocket()
	go server.WSSubscribe(myRedis)

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}", server.WebSocketHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         server.host + ":" + server.port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go log.Fatal(srv.ListenAndServe())
}

//-------------------WEB Socket--------------------------------------------------------------------------------------------
func (server *WebSocketServer)WebSocketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	//http://..../device/type/name/mac
	uri := strings.Split(r.URL.String(), "/")

	if _, ok := server.mapConn[uri[2]]; !ok {
		server.mapConn[uri[2]] = new(listConnection)
	}
	server.mapConn[uri[2]].Add(conn)
}

/**
Delete connections in mapConn
*/
func (server *WebSocketServer)CloseWebsocket() {
	for {
		select {
		case connAddres := <-server.connChanal:
			for _, val := range server.mapConn {
				if ok := val.Remove(connAddres); ok {
					break
				}
			}
		case <-server.stopCloseWS:
			log.Info("CloseWebsocket closed")
			return
		}
	}
}

/*
Listens changes in database. If they have, we will send to all websocket which working with them.
*/
func  (server *WebSocketServer) WSSubscribe(dbWorker dao.DbWorker) {
	dbWorker.Subscribe(server.subWSChannel,server.roomIDForDevWSPublish)
	for {
		select {
		case msg := <-server.subWSChannel:
			if msg[0] == "message" {
				go server.checkAndSendInfoToWSClient(msg)
			}
		case <-server.stopSub:
			log.Info("WSSubscribe closed")
			return
		}
	}
}

//We are check mac in our mapConnections.
// If we have mac in the map we will send message to all connections.
// Else we do nothing
func (server *WebSocketServer) checkAndSendInfoToWSClient(msg []string) {
	r := new(models.Request)
	err := json.Unmarshal([]byte(msg[2]), &r)
	if CheckError("checkAndSendInfoToWSClient", err) != nil {
		return
	}
	if _, ok := server.mapConn[r.Meta.MAC]; ok {
		server.sendInfoToWSClient(r.Meta.MAC, msg[2])
		return
	}
	log.Infof("mapConn dont have this MAC: %v. Len map is %v", r.Meta.MAC, len(server.mapConn))
}

//Send message to all connections which we have in map, and which pertain to mac
func  (server *WebSocketServer) sendInfoToWSClient(mac, message string) {
	server.mapConn[mac].Lock()
	for _, val := range server.mapConn[mac].connections {
		err := val.WriteMessage(1, []byte(message))
		if err != nil {
			log.Errorf("Connection %v closed", val.RemoteAddr())
			go getToChanal(val, server.connChanal)
		}
	}
	server.mapConn[mac].Unlock()
}

func getToChanal(conn *websocket.Conn, connChan chan *websocket.Conn) {
	connChan <- conn
}


