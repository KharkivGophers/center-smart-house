package webSocket

import (
	"time"
	"net/http"
	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"


	"github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/common"
	. "github.com/KharkivGophers/center-smart-house/common/models"
)

type WSServer struct {
	Host             string
	Port             string
	WSConnectionsMap WSConnectionsMap
	PubSub           PubSub
	DBURL            DBURL
	Upgrader         websocket.Upgrader
}

func NewWebSocketConnections() *WSConnectionsMap {
	var (
		connChanCloseWS = make(chan *websocket.Conn)
		stopCloseWS     = make(chan string)
		mapConn         = make(map[string]*ListConnection)
	)
	return &WSConnectionsMap{ConnChanCloseWS: connChanCloseWS, StopCloseWS: stopCloseWS, MapConn: mapConn}
}

func NewPubSub(roomIDForWSPubSub string, stopSub chan bool, subWSChannel chan []string) *PubSub {
	return &PubSub{RoomIDForWSPubSub: roomIDForWSPubSub, StopSub: stopSub, SubWSChannel: subWSChannel}
}

func NewWebSocketServer(wsHost, wsPort, dbhost string, dbPort uint) *WSServer {
	var (
		roomIDForDevWSPublish = "devWS"
		stopSub               = make(chan bool)
		subWSChannel          = make(chan []string)
	)

	dburl := DBURL{DbHost: dbhost, DbPort: dbPort}
	return NewWSServer(wsHost, wsPort, *NewPubSub(roomIDForDevWSPublish, stopSub, subWSChannel), dburl,
		*NewWebSocketConnections())

}

// Return referenced address on the WSServer with default Upgrader where:
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *myHTTP.Request) bool {
//			return true
//	}
func NewWSServer(host, port string, pubSub PubSub, dburi DBURL, wsConnections WSConnectionsMap) *WSServer {
	var (
		upgrader = websocket.Upgrader{
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
	return &WSServer{host, port, wsConnections, pubSub,
			 dburi, upgrader}
}

//myHTTP web socket connection
func (server *WSServer) StartWebSocketServer() {

	myRedis, err := dao.MyRedis{Host: server.Host, Port: server.DBURL.DbPort}.RunDBConnection()
	CheckError("webSocket: runDBConnection", err)

	go server.CloseWebsocket()
	go server.WSSubscribe(myRedis)

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}", server.WebSocketHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         server.Host + ":" + server.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go log.Fatal(srv.ListenAndServe())
}

//-------------------WEB Socket--------------------------------------------------------------------------------------------
func (server *WSServer) WebSocketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := server.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	//http://..../device/type/name/mac
	server.WSConnectionsMap.Lock()
	uri := strings.Split(r.URL.String(), "/")

	if _, ok := server.WSConnectionsMap.MapConn[uri[2]]; !ok {
		server.WSConnectionsMap.MapConn[uri[2]] = new(ListConnection)
	}
	server.WSConnectionsMap.MapConn[uri[2]].Add(conn)
	server.WSConnectionsMap.Unlock()

}

/**
Delete connections in mapConn
*/
func (server *WSServer) CloseWebsocket() {
	for {
		select {
		case connAddres := <-server.WSConnectionsMap.ConnChanCloseWS:
			for _, val := range server.WSConnectionsMap.MapConn {
				if ok := val.Remove(connAddres); ok {
					break
				}
			}
		case <-server.WSConnectionsMap.StopCloseWS:
			log.Info("CloseWebsocket closed")
			return
		}
	}
}

/*
Listens changes in database. If they have, we will send to all websocket which working with them.
*/
func (server *WSServer) WSSubscribe(dbWorker dao.DbWorker) {
	dbWorker.Subscribe(server.PubSub.SubWSChannel, server.PubSub.RoomIDForWSPubSub)
	for {
		select {
		case msg := <-server.PubSub.SubWSChannel:
			if msg[0] == "message" {
				go server.checkAndSendInfoToWSClient(msg)
			}
		case <-server.PubSub.StopSub:
			log.Info("WSSubscribe closed")
			return
		}
	}
}



//We are check mac in our mapConnections.
// If we have mac in the map we will send message to all connections.
// Else we do nothing
func (server *WSServer) checkAndSendInfoToWSClient(msg []string) {
	r := new(Request)
	err := json.Unmarshal([]byte(msg[2]), &r)
	if CheckError("checkAndSendInfoToWSClient", err) != nil {
		return
	}
	if _, ok := server.WSConnectionsMap.MapConn[r.Meta.MAC]; ok {
		server.sendInfoToWSClient(r.Meta.MAC, msg[2])
		return
	}
	log.Infof("mapConn dont have this MAC: %v. Len map is %v", r.Meta.MAC, len(server.WSConnectionsMap.MapConn))
}

//Send message to all connections which we have in map, and which pertain to mac
func (server *WSServer) sendInfoToWSClient(mac, message string) {
	server.WSConnectionsMap.MapConn[mac].Lock()
	for _, val := range server.WSConnectionsMap.MapConn[mac].Connections {
		err := val.WriteMessage(1, []byte(message))
		if err != nil {
			log.Errorf("Connection %v closed", val.RemoteAddr())
			go getToChanal(val, server.WSConnectionsMap.ConnChanCloseWS)
		}
	}
	server.WSConnectionsMap.MapConn[mac].Unlock()
}

func getToChanal(conn *websocket.Conn, connChanCloseWS chan *websocket.Conn) {
	connChanCloseWS <- conn
}
