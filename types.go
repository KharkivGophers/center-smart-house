package main

import "encoding/json"
import "sync"
import "net"

type ServerError struct {
	Error   error
	Message string
	Code    int
}

type Response struct {
	Status int    `json:"status"`
	Descr  string `json:"descr"`
	Config Config `json:"config"`
}

type Config struct {
	State       bool   `json:"state"`
	CollectFreq int    `json:"collectFreq"`
	SendFreq    int    `json:"sendFreq"`
	MAC         string `json:"mac"`
}

type Metadata struct {
	Type string `json:"type"`
	Name string `json:"name"`
	MAC  string `json:"mac"`
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
type ViewDevice struct {
	Location  string   `json:location`
	Name      string   `json:name`
	Type      string   `json:type`
	Functions []string `json:functions`
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
