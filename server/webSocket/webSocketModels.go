package webSocket

import (
	"github.com/gorilla/websocket"
	"sync"
	log "github.com/Sirupsen/logrus"
)

type WSConnectionsMap struct {
	ConnChanCloseWS chan *websocket.Conn
	StopCloseWS     chan string
	MapConn         map[string]*ListConnection
	sync.Mutex
}

type PubSub struct {
	RoomIDForWSPubSub string
	StopSub           chan bool
	SubWSChannel      chan []string
}

//For work with web socket
type ListConnection struct {
	sync.Mutex
	Connections []*websocket.Conn
}

func (list *ListConnection) Add(conn *websocket.Conn) {
	list.Lock()
	list.Connections = append(list.Connections, conn)
	list.Unlock()

}
func (list *ListConnection) Remove(conn *websocket.Conn) bool {
	list.Lock()
	defer list.Unlock()
	position := 0
	for _, v := range list.Connections {
		if v == conn {
			list.Connections = append(list.Connections[:position], list.Connections[position+1:]...)
			log.Info("Web sockets connection deleted: ", conn.RemoteAddr())
			return true
		}
		position++
	}
	return false
}

func (connMap *WSConnectionsMap) Remove(mac string){
	connMap.Lock()
	delete(connMap.MapConn, mac)
	connMap.Unlock()
}
///////////------------------------------------------------------------------------------------>
func (connMap *WSConnectionsMap) MapCollector(mac chan  string, ){
	connMap.Lock()
	delete(connMap.MapConn, mac)
	connMap.Unlock()
}
