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

	"errors"

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
	checkError("DB error11", err)
	_, err = dbClient.HMSet(devKey, "ReqTime", devReqTime)
	checkError("DB error12", err)
	_, err = dbClient.SAdd(devParamsKey, "TempCam1", "TempCam2")
	checkError("DB error13", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		return &ServerError{Error: err}
	}

	for time, value := range devData.TempCam1 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam1",
			int64ToString(time), int64ToString(time)+":"+float32ToString(float64(value)))
		if checkError("DB error14", err) != nil {
			return &ServerError{Error: err}
		}
	}

	for time, value := range devData.TempCam2 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam2",
			int64ToString(time), int64ToString(time)+":"+float32ToString(float64(value)))
		if checkError("DB error15", err) != nil {
			return &ServerError{Error: err}
		}
	}

	return nil
}

func (req *Request) washerDataHandler() *ServerError {
	//to be continued
	return nil
}

func configSubscribe(client *redis.Client, roomID string, messages chan []string, pool *ConectionPool) {
	subscribe(client, roomID, messages)

	var temp interface{}
	var devConfigTurnedOn DevConfigTurnedOn
	var devConfigFreqs DevConfigFreqs

	for msg := range messages {
		if msg[0] == "message" {
			log.Warningln("for range from publish chann", msg)
			if strings.Contains(msg[2], "turned") {
				err := json.Unmarshal([]byte(msg[2]), &devConfigTurnedOn)
				if checkError("configSubscribe", err) != nil {
					return
				}
				go sendNewConfiguration(devConfigTurnedOn, pool)
			}
			err := json.Unmarshal([]byte(msg[2]), &devConfigFreqs)
			if checkError("configSubscribe", err) != nil {
				return
			}
			go sendNewConfiguration(devConfigFreqs, pool)
			log.Println("subscribe, temp interface: ", temp)
		}
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

func getDevConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"] // type + mac
	mac := strings.Split(id, ":")[2]

	configInfo := mac + ":" + "config" // key

	state, _ := dbClient.ZRange(configInfo+":"+"TurnedOn", 0, 0)
	sendFreq, _ := dbClient.ZRange(configInfo+":"+"CollectFreq", 0, 0)
	collectFreq, _ := dbClient.ZRange(configInfo+":"+"SendFreq", 0, 0)

	// log.Println(state)
	// log.Println(sendFreq)
	// log.Println(collectFreq)

	newState, _ := strconv.ParseBool(state[0])
	newSendFreq, _ := strconv.ParseInt(sendFreq[0], 10, 64)
	newCollectFreq, _ := strconv.ParseInt(collectFreq[0], 10, 64)

	// log.Println(reflect.TypeOf(newState))
	// log.Println(reflect.TypeOf(newSendFreq))
	// log.Println(reflect.TypeOf(newCollectFreq))

	var config = DevConfig{
		TurnedOn: newState,
		// StreamOn:    true,
		CollectFreq: newCollectFreq,
		SendFreq:    newSendFreq,
	}

	json.NewEncoder(w).Encode(config)

}

func patchDevConfigHandler(w http.ResponseWriter, r *http.Request) {
	var configTurnedOn DevConfigTurnedOn
	var configFreqs DevConfigFreqs
	var configInterface map[string]interface{}
	var chanTurnedOn = make(chan []byte)
	var chanFreqs = make(chan []byte)
	// var configFreqsconfig DevConfig

	vars := mux.Vars(r)
	id := vars["id"] // type + mac
	mac := strings.Split(id, ":")[2]
	go validateMAC(mac)

	configInfo := mac + ":" + "config" // key
	Time := time.Now().UnixNano() / int64(time.Millisecond)

	err := json.NewDecoder(r.Body).Decode(&configInterface)
	if err != nil {
		// http.Error(w, err.Error(), 400)
		log.Errorln("NewDec: ", err)
	}
	go publishConfig(chanTurnedOn, chanFreqs)

	for k, v := range configInterface {
		if k == "turnedOn" {
			configTurnedOn.TurnedOn = v.(bool)
			configTurnedOn.MAC = mac

			j, err := json.Marshal(configTurnedOn)
			checkError("JSONconfig error", err)
			chanTurnedOn <- j

			_, err = dbClient.ZAdd(configInfo+":"+"TurnedOn", Time, configTurnedOn.TurnedOn)
			checkError("DB error01", err)
			return
		}

		switch k {
		case "sendFreq":
			configFreqs.SendFreq = int64(v.(float64))
			log.Println("into switch", int64(v.(float64)))
		case "collectFreq":
			configFreqs.CollectFreq = int64(v.(float64))
			log.Println("into switch", int64(v.(float64)))
		}
	}

	configFreqs.MAC = mac

	j, err := json.Marshal(configFreqs)
	checkError("JSONconfig error", err)
	chanFreqs <- j

	_, err = dbClient.ZAdd(configInfo+":"+"CollectFreq", Time, configFreqs.CollectFreq)
	checkError("DB error02", err)
	_, err = dbClient.ZAdd(configInfo+":"+"SendFreq", Time, configFreqs.SendFreq)
	checkError("DB error03", err)
	// add MAC to map

	// Save New Configuration to DB
	// _, err = dbClient.ZAdd(configInfo+":"+"TurnedOn", Time, config.TurnedOn)
	// checkError("DB error", err)
	// _, err = dbClient.ZAdd(configInfo+":"+"CollectFreq", Time, config.CollectFreq)
	// checkError("DB error", err)
	// _, err = dbClient.ZAdd(configInfo+":"+"SendFreq", Time, config.SendFreq)
	// checkError("DB error", err)
	// w.WriteHeader(http.StatusOK)

}

func publishConfig(turnedOn chan []byte, freqs chan []byte) {
	for {
		select {
		case confTurnedOn := <-turnedOn:
			log.Warningln("confTurnedOn", confTurnedOn)
			go publishConfigMessage(confTurnedOn, "configChan")
		case confFreqs := <-freqs:
			log.Warningln("confFreqs", confFreqs)
			go publishConfigMessage(confFreqs, "configChan")
		}
	}
}
func validateSendFreq(c DevConfig) error {
	log.Println(reflect.TypeOf(c.SendFreq).String())
	switch reflect.TypeOf(c.SendFreq).String() {
	case "int64":
		switch {
		case c.SendFreq < 0 && c.SendFreq == 0:
			return errors.New("Failed SendFreq validation")
		default:
			log.Println("validateSendFreq complited sucessfully")
			return nil
		}
	default:
		return errors.New("Failed SendFreq validation")
	}
}

func validateCollectFreq(c DevConfig) error {
	log.Println(reflect.TypeOf(c.CollectFreq).String())
	switch reflect.TypeOf(c.CollectFreq).String() {
	case "int64":
		switch {
		case c.CollectFreq < 0 && c.CollectFreq == 0:
			return errors.New("Failed CollectFreq validation")
		default:
			log.Println("validateCollectFreq complited sucessfully")
			return nil
		}
	default:
		return errors.New("Failed CollectFreq validation")
	}
}

func validateState(c DevConfig) error {
	switch reflect.TypeOf(c.TurnedOn).String() {
	case "bool":
		log.Println("validateState complited sucessfully")
		return nil
	default:
		return errors.New("Failed validateState")
	}
}

func validateMAC(mac string) {
	switch reflect.TypeOf(mac).String() {
	case "string":
		fmt.Println("This is string")
		switch len(mac) {
		case 17:
			return
		default:
			log.Println("MAC should contain 17 symbols")
			break
		}
	default:
		log.Println("MAC should be in string format")
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

	if _, ok := mapConn[uri[2]]; !ok {
		mapConn[uri[2]] = new(listConnection)
	}
	mapConn[uri[2]].Add(conn)

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
			if msg[0] == "message" {
				go checkAndSendInfoToWSClient(msg)
			}
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
		err := val.WriteMessage(1, []byte(message))
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

func publishWS(req Request) {
	pubReq, err := json.Marshal(req)
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
func publishConfigMessage(message []byte, roomID string) {
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

func getAllDevices() []DevData {
	var device DevData
	var devices []DevData
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

func getDevice(devParamsKey string, devParamsKeysTokens []string) DevData {
	var device DevData

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
