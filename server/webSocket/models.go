package webSocket

import (
	"sync"
	"github.com/gorilla/websocket"
	log "github.com/Sirupsen/logrus"
)

//For work with web socket
type listConnection struct {
	sync.Mutex
	connections []*websocket.Conn
}

func (list *listConnection) Add(conn *websocket.Conn) {
	list.Lock()
	list.connections = append(list.connections, conn)
	list.Unlock()

}
func (list *listConnection) Remove(conn *websocket.Conn) bool {
	list.Lock()
	defer list.Unlock()
	position := 0
	for _, v := range list.connections {
		if v == conn {
			list.connections = append(list.connections[:position], list.connections[position+1:]...)
			log.Info("Web sockets connection deleted: ", conn.RemoteAddr())
			return true
		}
		position++
	}
	return false
}
