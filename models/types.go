package models

import (
	"encoding/json"
	"net"
	"sync"
	"time"
	"log"
)

type Control struct {
	Controller chan struct{}
}

func (c *Control) Close() {
	select {
	case <- c.Controller:
	default:
		close(c.Controller)
	}
}

func (c *Control) Wait() {
	<- c.Controller
	<-time.NewTimer(6).C
}

type Server struct {
	IP string
	Port uint
}

type ServerError struct {
	Error   error
	Message string
	Code    int
}

type Response struct {
	Status int    `json:"status"`
	Descr  string `json:"descr"`
}

type DevConfig struct {
	TurnedOn    bool   `json:"turnedOn"`
	StreamOn    bool   `json:"streamOn"`
	CollectFreq int64  `json:"collectFreq"`
	SendFreq    int64  `json:"sendFreq"`
	MAC         string `json:"mac"`
}

type DevMeta struct {
	Type string `json:"type"`
	Name string `json:"name"`
	MAC  string `json:"mac"`
	IP   string `json:"ip"`
}

type Request struct {
	Action string          `json:"action"`
	Time   int64           `json:"time"`
	Meta   DevMeta         `json:"meta"`
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

type DevData struct {
	Site string              `json:"site"`
	Meta DevMeta             `json:"meta"`
	Data map[string][]string `json:"data"`
}

//Connections pool for configTCPServer
type ConnectionPool struct {
	sync.Mutex
	conn map[string]net.Conn
}

func (pool *ConnectionPool) AddConn(conn net.Conn, key string) {
	pool.Lock()
	pool.conn[key] = conn
	defer pool.Unlock()
}

func (pool *ConnectionPool) GetConn(key string) net.Conn {
	pool.Lock()
	defer pool.Unlock()
	return pool.conn[key]
}

func (pool *ConnectionPool) RemoveConn(key string)  {
	pool.Lock()
	defer pool.Unlock()
	 delete(pool.conn, key)
}

func (pool *ConnectionPool) Init() {
	pool.Lock()
	defer pool.Unlock()

	pool.conn = make(map[string]net.Conn)
}


