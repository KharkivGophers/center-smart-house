package main

import (
	"encoding/json"
	"net"
	"strings"
	"net/http"
	"strconv"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"fmt"
)

var (
	updDevConfigList = "updDevConfigList"
	updDevDataList   = "updDevDataList"
)

//For work with web socket
var (
	connection = make(map[*websocket.Conn]bool)
	upgrader   = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func requestHandler(conn net.Conn) {
	var req Request
	var res Response

	err := json.NewDecoder(conn).Decode(&req)
	if err != nil {
		log.Errorln(err)
	}

	defer conn.Close()

	go devTypeHandler(req)

	res = Response{
		Status: http.StatusOK,
		Descr:  "Data have been delivered successfully",
	}

	err = json.NewEncoder(conn).Encode(&res)
	if err != nil {
		log.Errorln(err)
	}
}

func devTypeHandler(req Request) {
	switch req.Action {
	case "update":
		switch req.Meta.Type {
		case "fridge":
			if err := req.fridgeDataHandler(); err != nil {
				log.Errorf("%v", err.Error)
			}
		case "washer":
			if err := req.washerDataHandler(); err != nil {
				log.Errorf("%v", err.Error)
			}

		default:
			log.Println("Device request: unknown device type")
		}

	default:
		log.Println("Device request: unknown action")
	}
}

func (req *Request) fridgeDataHandler() *ServerError {
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	devKey := "device" + ":" + devType + ":" + devName + ":" + mac
	devDataKey := devKey + ":" + "data"
	//updListKey := updDevDataList + ":" + devKey

	dbClient.SAdd("devDataKeys", devDataKey)
	dbClient.HMSet(devKey, "ReqTime", devReqTime)
	dbClient.SAdd(devDataKey, "TempCam1", "TempCam2")
	dbClient.LPush(updDevDataList, devKey)

	var devData FridgeData
	json.Unmarshal([]byte(req.Data), &devData)

	for time, value := range devData.TempCam1 {
		dbClient.ZAdd(devDataKey+":"+"TempCam1",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))

		/*dbClient.ZAdd(updListKey + ":" + "TempCam1",
			Int64ToString(time), Int64ToString(time) + ":" + Float32ToString(float64(value))) */
	}

	for time, value := range devData.TempCam2 {
		dbClient.ZAdd(devDataKey+":"+"TempCam2",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))

		/*dbClient.ZAdd(updListKey + ":" + "TempCam2",
			Int64ToString(time), Int64ToString(time) + ":" + Float32ToString(float64(value))) */
	}

	return nil
}

func (req *Request) washerDataHandler() *ServerError {
	//to be continued
	return nil
}

func Float32ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 32)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(int64(n), 10)
}

func httpDevHandler(w http.ResponseWriter, r *http.Request) {

	devKeys, _ := dbClient.SMembers("devDataKeys")

	var devKeysTokens [][]string = make([][]string, len(devKeys))
	for index, key := range devKeys {
		devKeysTokens[index] = strings.Split(key, ":")
	}

	var device Device
	var devices []Device

	for index, key := range devKeysTokens {
		params, _ := dbClient.SMembers(devKeys[index])

		device.Type = key[1]
		device.Name = key[2]
		device.Data = make(map[string][]string)

		values := make([][]string, len(params))
		for i, p := range params {
			values[i], _ = dbClient.ZRangeByScore(devKeys[index]+":"+p, "-inf", "inf")
			device.Data[p] = values[i]
		}

		devices = append(devices, device)
	}
	json.NewEncoder(w).Encode(devices)
}

//-------------------WEB Socket Handler -----------------------
func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connection[conn] = true
	defer delete(connection, conn)

	for {
		err := conn.WriteJSON("23")
		if err != nil {
			log.Println("write:", err)
			break
		}

	}

}


