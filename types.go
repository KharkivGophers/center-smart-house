package main

import "encoding/json"

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
	Turned      bool `json:"turned"`
	CollectFreq int  `json:"collectFreq"`
	SendFreq    int  `json:"sendFreq"`
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
