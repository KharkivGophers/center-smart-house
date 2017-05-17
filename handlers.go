package main

import (
	"encoding/json"
	"net"
	"strings"
	"net/http"
	"strconv"
	log "github.com/logrus"
)

var (
	updDevData = "updDevData"
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
	devParamsKey := devKey + ":" + "params"
	//updDevDataKey := updDevDataList + ":" + devKey

	dbClient.SAdd("devParamsKeys", devParamsKey)
	dbClient.HMSet(devKey, "ReqTime", devReqTime)
	dbClient.SAdd(devParamsKey, "TempCam1", "TempCam2")
	//dbClient.LPush(updDevDataList, devKey)

	var devData FridgeData
	json.Unmarshal([]byte(req.Data), &devData)

	for time, value := range devData.TempCam1 {
		dbClient.ZAdd(devParamsKey + ":" + "TempCam1",
			Int64ToString(time), Int64ToString(time) + ":" + Float32ToString(float64(value)))

		/*dbClient.ZAdd(updListKey + ":" + "TempCam1",
			Int64ToString(time), Int64ToString(time) + ":" + Float32ToString(float64(value))) */
	}

	for time, value := range devData.TempCam2 {
		dbClient.ZAdd(devParamsKey + ":" + "TempCam2",
			Int64ToString(time), Int64ToString(time) + ":" + Float32ToString(float64(value)))

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
	devParamsKeys, _ := dbClient.SMembers("devParamsKeys")

	var devParamsKeysTokens [][]string = make([][]string, len(devParamsKeys))
	for i, k := range devParamsKeys {
		devParamsKeysTokens[i] = strings.Split(k, ":")
	}

	var device Device
	var devices []Device

	for index, key := range devParamsKeysTokens {
		params, _ := dbClient.SMembers(devParamsKeys[index])

		device.Type = key[1]
		device.Name = key[2]
		device.Data = make(map[string][]string)

		values := make([][]string, len(params))
		for i, p := range params {
			values[i], _ = dbClient.ZRangeByScore(devParamsKeys[index] + ":" + p, "-inf", "inf")
			device.Data[p] = values[i]
		}

		devices = append(devices, device)
	}

	json.NewEncoder(w).Encode(devices)
}


