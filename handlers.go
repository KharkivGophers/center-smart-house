package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"menteslibres.net/gosexy/redis"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//--------------------TCP-------------------------------------------------------------------------------------
func tcpDataHandler(conn *net.Conn) {
	var req Request
	var res Response
	for {
		err := json.NewDecoder(*conn).Decode(&req)
		if err != nil {
			log.Errorln("requestHandler JSON Decod", err)
			return
		}
		//sends resp struct from  devTypeHandler by channel;
		go devTypeHandler(req)

		log.Println("Data has been received")

		res = Response{
			Status: http.StatusOK,
			Descr:  "Data have been delivered successfully",
		}
		err = json.NewEncoder(*conn).Encode(&res)
		checkError("tcpDataHandler JSON enc", err)
	}

}

/*
Checks  type device and call speciall func for send data to DB.
*/
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
			return
		}
		go publishWS(req)

	default:
		log.Println("Device request: unknown action")
	}
}

/*
Save data about fridge in DB. Return struct ServerError
*/
func (req *Request) fridgeDataHandler() *ServerError {
	var devData FridgeData
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	devKey := "device" + ":" + devType + ":" + devName + ":" + mac
	devParamsKey := devKey + ":" + "params"

	_, err := dbClient.SAdd("devParamsKeys", devParamsKey)
	checkError("DB error", err)
	_, err = dbClient.HMSet(devKey, "ReqTime", devReqTime)
	checkError("DB error", err)
	_, err = dbClient.SAdd(devParamsKey, "TempCam1", "TempCam2")
	checkError("DB error", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		return &ServerError{Error: err}
	}

	for time, value := range devData.TempCam1 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam1",
			int64ToString(time), int64ToString(time)+":"+float32ToString(float64(value)))
		checkError("DB error", err)
		return &ServerError{Error: err}
	}

	for time, value := range devData.TempCam2 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam2",
			int64ToString(time), int64ToString(time)+":"+float32ToString(float64(value)))
		checkError("DB error", err)
		return &ServerError{Error: err}
	}

	return nil
}

func (req *Request) washerDataHandler() *ServerError {
	//to be continued
	return nil
}

func configSubscribe(client *redis.Client, roomID string, messages chan []string, pool *ConectionPool) {
	subscribe(client, roomID, messages)
	var cnfg Config
	for msg := range messages {
		err := json.Unmarshal([]byte(msg[2]), &cnfg)
		if checkError("configSubscribe", err) != nil {
			return
		}
		sendNewConfiguration(cnfg, pool)
	}
}

//----------------------HTTP Dynamic Connection----------------------------------------------------------------------------------

func getDevicesHandler(w http.ResponseWriter, r *http.Request) {
	devices := getAllDevices()
	err := json.NewEncoder(w).Encode(devices)
	checkError("getDevicesHandler JSON enc", err)
}

func getDevDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	devID := "device:" + vars["id"]

	devParamsKeysTokens := []string{}
	devParamsKeysTokens = strings.Split(devID, ":")
	devParamsKey := devID + ":" + "params"

	device := getDevice(devParamsKey, devParamsKeysTokens)
	err := json.NewEncoder(w).Encode(device)
	checkError("getDevDataHandler JSON enc", err)
}

func postDevConfigHandler(w http.ResponseWriter, r *http.Request) {
	var config Config
	vars := mux.Vars(r)

	id := "device:" + vars["id"]
	mac := vars["id"]
	configInfo := vars["if"] + ":" + "data"

	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	Time := time.Now().UnixNano() / int64(time.Millisecond)
	log.Println("Received configuration: ", config, "id: ", id)
	w.WriteHeader(http.StatusOK)

	// Validate MAC
	validateMAC(mac, w)
	// Validate State
	validateState(config, w)
	// Validate Collect Frequency
	validateCollectFreq(config, w)
	// Validate Send Frequency
	validateSendFreq(config, w)

	// Save to DB
	_, err = dbClient.SAdd("Config", configInfo)
	checkError("DB error", err)
	_, err = dbClient.HMSet(vars["id"], "ConfigTime", Time)
	checkError("DB error", err)
	_, err = dbClient.SAdd(configInfo, "TurnedOn", "CollectFreq", "SendFreq")
	checkError("DB error", err)
	_, err = dbClient.ZAdd(configInfo+":"+"TurnedOn", config.TurnedOn)
	checkError("DB error", err)
	_, err = dbClient.ZAdd(configInfo+":"+"CollectFreq", config.CollectFreq)
	checkError("DB error", err)
	_, err = dbClient.ZAdd(configInfo+":"+"SendFreq", config.SendFreq)
	checkError("DB error", err)

	log.Println("Received configuration: ", config, "id: ", id)
	w.WriteHeader(http.StatusOK)
}

func validateSendFreq(c Config, w http.ResponseWriter) {
	switch reflect.TypeOf(c.SendFreq).String() {
	case "int":
		switch {
		case c.SendFreq > 0 && c.SendFreq < 100:
			return
		default:
			http.Error(w, "0 < Send Frequency < 100", 400)
			break
		}
	default:
		http.Error(w, "Send Frequency should be in integer format", 415)
		break
	}
}

func validateCollectFreq(c Config, w http.ResponseWriter) {
	switch reflect.TypeOf(c.CollectFreq).String() {
	case "int":
		switch {
		case c.CollectFreq > 0 && c.CollectFreq < 100:
			return
		default:
			http.Error(w, "0 < Collect Frequency < 100", 400)
			break
		}
	default:
		http.Error(w, "Collect Frequency should be in integer format", 415)
		break
	}
}

func validateState(c Config, w http.ResponseWriter) {
	switch reflect.TypeOf(c.TurnedOn).String() {
	case "bool":
		return
	default:
		http.Error(w, "State should be in byte format", 415)
		break
	}
}

func validateMAC(mac string, w http.ResponseWriter) {
	switch reflect.TypeOf(mac).String() {
	case "string":
		fmt.Println("This is string")
		switch len(mac) {
		case 17:
			return
		default:
			http.Error(w, "MAC should contain 17 symbols", 400)
			break
		}
	default:
		http.Error(w, "MAC should be in string format", 415)
		break
	}
}

//-------------------WEB Socket--------------------------------------------------------------------------------------------
func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	//http://..../device/type/name/mac
	uri := strings.Split(r.URL.String(), "/")

	if _, ok := mapConn[uri[4]]; !ok {
		mapConn[uri[4]] = new(listConnection)
	}
	mapConn[uri[4]].Add(conn)
}

/**
Delete connections from mapConn
*/
func CloseWebsocket() {
	for {
		select {
		case connAddres := <-connChanal:
			for _, val := range mapConn {
				if ok := val.Remove(connAddres); ok {
					break
				}
			}
		case <-stopCloseWS:
			log.Info("CloseWebsocket closed")
			return
		}
	}
}

/*
Listens changes in database. If they have, then sent to all websocket which working with them.
*/
func WSSubscribe(client *redis.Client, roomID string, channel chan []string) {
	subscribe(client, roomID, channel)
	for {
		select {
		case msg := <-channel:
			go checkAndSendInfoToWSClient(msg)
		case <-stopSub:
			log.Info("WSSubscribe closed")
			return
		}
	}
}

//We are check mac in our mapConnections.
// If we have mac in the map we will send message to all connections.
// Else we do nothing
func checkAndSendInfoToWSClient(msg []string) {
	r := new(Request)
	log.Warningln(msg[2])
	err := json.Unmarshal([]byte(msg[2]), &r)
	if checkError("checkAndSendInfoToWSClient", err) != nil {
		return
	}
	if _, ok := mapConn[r.Meta.MAC]; ok {
		sendInfoToWSClient(r.Meta.MAC, msg[2])
	}
}

//Send message to all connections which we have in map, and which pertain to mac
func sendInfoToWSClient(mac, message string) {
	mapConn[mac].Lock()
	for _, val := range mapConn[mac].connections {
		fmt.Println(message)
		err := val.WriteJSON(message)
		if err != nil {
			log.Errorf("Connection %v closed", val.RemoteAddr())
			go getToChanal(val)
		}
	}
	mapConn[mac].Unlock()
}

func getToChanal(conn *websocket.Conn) {
	connChanal <- conn
}

func publishWS(req Request){
	pubReq ,err := json.Marshal(req)
	checkError("Marshal for publish.", err)
	go publishMessage(pubReq, roomIDForDevWSPublish)
}

//-----------------Common funcs-------------------------------------------------------------------------------------------

func checkError(desc string, err error) error {
	if err != nil {
		log.Errorln(desc, err)
		return err
	}
	return nil
}

func subscribe(client *redis.Client, roomID string, channel chan []string) {
	client = redis.New()
	err := client.ConnectNonBlock(dbHost, dbPort)
	checkError("Subscribe", err)

	go client.Subscribe(channel, roomID)
}

func publishMessage(message interface{}, roomID string) {
	_, err := dbClient.Publish(roomID, message)
	checkError("Publish", err)
}

func float32ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 32)
}

func int64ToString(n int64) string {
	return strconv.FormatInt(int64(n), 10)
}

//-------------------Work with data base-------------------------------------------------------------------------------------------

func getAllDevices() []DeviceView {
	var device DeviceView
	var devices []DeviceView
	devParamsKeys, err := dbClient.SMembers("devParamsKeys")
	if checkError("Cant read members from devParamsKeys", err) != nil {
		return nil
	}

	var devParamsKeysTokens = make([][]string, len(devParamsKeys))
	for i, k := range devParamsKeys {
		devParamsKeysTokens[i] = strings.Split(k, ":")
	}

	for index, key := range devParamsKeysTokens {
		params, err := dbClient.SMembers(devParamsKeys[index])
		checkError("Cant read members from "+devParamsKeys[index], err)

		device.Meta.Type = key[1]
		device.Meta.Name = key[2]
		device.Meta.MAC = key[3]
		device.Data = make(map[string][]string)

		values := make([][]string, len(params))
		for i, p := range params {
			values[i], _ = dbClient.ZRangeByScore(devParamsKeys[index]+":"+p, "-inf", "inf")
			checkError("Cant use ZRangeByScore for "+devParamsKeys[index], err)
			device.Data[p] = values[i]
		}

		devices = append(devices, device)
	}
	return devices
}

func getDevice(devParamsKey string, devParamsKeysTokens []string) DetailedDevData {
	var device DetailedDevData

	params, err := dbClient.SMembers(devParamsKey)
	checkError("Cant read members from devParamsKeys", err)
	device.Meta.Type = devParamsKeysTokens[1]
	device.Meta.Name = devParamsKeysTokens[2]
	device.Meta.MAC = devParamsKeysTokens[3]
	device.Data = make(map[string][]string)

	values := make([][]string, len(params))
	for i, p := range params {
		values[i], err = dbClient.ZRangeByScore(devParamsKey+":"+p, "-inf", "inf")
		checkError("Cant use ZRangeByScore", err)
		device.Data[p] = values[i]
	}
	return device
}
