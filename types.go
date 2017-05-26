package main

import (
	"encoding/json"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

type ServerError struct {
	Error   error
	Message string
	Code    int
}

type Response struct {
	Status int    `json:"status"`
	Descr  string `json:"descr"`
}

type Config struct {
	TurnedOn    bool    `json:"turnedOn"`
	StreamOn    bool    `json:"streamOn"`
	CollectFreq float64 `json:"collectFreq"`
	SendFreq    float64 `json:"sendFreq"`
	MAC         string  `json:"mac"`
}

type Metadata struct {
	Type string `json:"type"`
	Name string `json:"name"`
	MAC  string `json:"mac"`
	IP   string `json:"ip"`
}

type Request struct {
	Action string          `json:"action"`
	Time   int64           `json:"time"`
	Meta   Metadata        `json:"meta"`
	Data   json.RawMessage `json:"data"`
}

type FridgeData struct {
	TempCam1 map[int64]float32 `json:"tempCam1"`
	TempCam2 map[int64]float32 `json:"tempCam2"`
}

type WasherData struct {
	Mode   string
	Drying string
	Temp   map[int64]float32
}

type DeviceView struct {
	Site string              `json:"site"`
	Meta Metadata            `json:"meta"`
	Data map[string][]string `json:"data"`
}

type DetailedDevData struct {
	Site   string              `json:"site"`
	Meta   Metadata            `json:"meta"`
	Config Config              `json:"config"`
	Data   map[string][]string `json:"data"`
}

type ConectionPool struct {
	sync.Mutex
	conn map[string]*net.Conn
}

func (pool *ConectionPool) addConn(conn *net.Conn, key string) {
	pool.Lock()
	pool.conn[key] = conn
	defer pool.Unlock()
}

func (pool *ConectionPool) getConn(key string) *net.Conn {
	pool.Lock()
	defer pool.Unlock()
	return pool.conn[key]
}
func (pool *ConectionPool) init() {
	pool.Lock()
	defer pool.Unlock()

	pool.conn = make(map[string]*net.Conn)
}

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
