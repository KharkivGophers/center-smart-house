package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"menteslibres.net/gosexy/redis"
	"github.com/KharkivGophers/center-smart-house/server/common/models"
	"github.com/KharkivGophers/center-smart-house/dao"
)

//--------------------TCP-------------------------------------------------------------------------------------
func tcpDataHandler(conn net.Conn) {
	var req models.Request
	var res Response
	for {
		err := json.NewDecoder(conn).Decode(&req)
		if err != nil {
			log.Errorln("requestHandler JSON Decod", err)
			return
		}
		//sends resp struct from  devTypeHandler by channel;
		go devTypeHandler(req)

		log.Println("Data has been received")

		res = Response{
			Status: http.StatusOK,
			Descr:  "Data has been delivered successfully",
		}
		err = json.NewEncoder(conn).Encode(&res)
		CheckError("tcpDataHandler JSON enc", err)
	}

}

/*
Checks  type device and call special func for send data to DB.
*/
func devTypeHandler(req models.Request) string {
	switch req.Action {
	case "update":
		switch req.Meta.Type {
		case "fridge":
			if err := fridgeDataHandler(&req); err != nil {
				log.Errorf("%v", err.Error)
			}
		case "washer":
			//if err := req.washerDataHandler(); err != nil {
			//	log.Errorf("%v", err.Error)
			//}

		default:
			log.Println("Device request: unknown device type")
			return string("Device request: unknown device type")
		}
		go dao.PublishWS(req,"devWS","0.0.0.0", 6379)
		//go publishWS(req)

	default:
		log.Println("Device request: unknown action")
		return string("Device request: unknown action")

	}
	return string("Device request correct")
}

/*
Save data about fridge in DB. Return struct ServerError
*/
func  fridgeDataHandler(req *models.Request) *ServerError {
	dbClient, _ := RunDBConnection()

	var devData FridgeData
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	devKey := "device" + ":" + devType + ":" + devName + ":" + mac
	devParamsKey := devKey + ":" + "params"

	_, err := dbClient.SAdd("devParamsKeys", devParamsKey)
	CheckError("DB error11", err)
	_, err = dbClient.HMSet(devKey, "ReqTime", devReqTime)
	CheckError("DB error12", err)
	_, err = dbClient.SAdd(devParamsKey, "TempCam1", "TempCam2")
	CheckError("DB error13", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		log.Error("Error in fridgedatahandler")
		return &ServerError{Error: err}
	}

	for time, value := range devData.TempCam1 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam1",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error14", err) != nil {
			return &ServerError{Error: err}
		}
	}

	for time, value := range devData.TempCam2 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam2",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error15", err) != nil {
			return &ServerError{Error: err}
		}
	}

	return nil
}

func (req *Request) washerDataHandler() *ServerError {
	//to be continued
	return nil
}

func configSubscribe(client *redis.Client, roomID string, message chan []string, pool *ConnectionPool) {
	Subscribe(client, roomID, message)
	for {
		var config DevConfig
		select {
		case msg := <-message:
			// log.Println("message", msg)
			if msg[0] == "message" {
				// log.Println("message[0]", msg[0])
				err := json.Unmarshal([]byte(msg[2]), &config)
				CheckError("configSubscribe: unmarshal", err)
				go sendNewConfiguration(config, pool)
			}
		}
	}
}

//----------------------HTTP Dynamic Connection----------------------------------------------------------------------------------

func getDevicesHandler(w http.ResponseWriter, r *http.Request) {
	client, _ := RunDBConnection()
	devices := getAllDevices(client)
	err := json.NewEncoder(w).Encode(devices)
	CheckError("getDevicesHandler JSON enc", err)
}

func getDevDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	devID := "device:" + vars["id"]

	devParamsKeysTokens := []string{}
	devParamsKeysTokens = strings.Split(devID, ":")
	devParamsKey := devID + ":" + "params"

	device := getDevice(devParamsKey, devParamsKeysTokens)
	err := json.NewEncoder(w).Encode(device)
	CheckError("getDevDataHandler JSON enc", err)
}

func getDevConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"] // type + mac
	mac := strings.Split(id, ":")[2]

	dbClient, _ := RunDBConnection()
	configInfo := mac + ":" + "config" // key

	var config = GetFridgeConfig(dbClient, configInfo, mac)

	json.NewEncoder(w).Encode(config)
}

func patchDevConfigHandler(w http.ResponseWriter, r *http.Request) {

	var config *DevConfig
	vars := mux.Vars(r)
	id := vars["id"] // warning!! type : name : mac
	mac := strings.Split(id, ":")[2]
	dbClient, _ := RunDBConnection()
	configInfo := mac + ":" + "config" // key

	config = GetFridgeConfig(dbClient, configInfo, mac)

	// log.Warnln("Config Before", config)
	err := json.NewDecoder(r.Body).Decode(&config)

	if err != nil {
		http.Error(w, err.Error(), 400)
		log.Errorln("NewDec: ", err)
	}

	log.Info(config)
	CheckError("Encode error", err) // log.Warnln("Config After: ", config)


	if !validateMAC(config.MAC) {
		log.Error("Invalid MAC")
		http.Error(w, "Invalid MAC", 400)
	} else if !validateStreamOn(config.StreamOn) {
		log.Error("Invalid Stream Value")
		http.Error(w, "Invalid Stream Value", 400)
	} else if !validateTurnedOn(config.TurnedOn) {
		log.Error("Invalid Turned Value")
		http.Error(w, "Invalid Turned Value", 400)
	} else if !validateCollectFreq(config.CollectFreq) {
		log.Error("Invalid Collect Frequency Value")
		http.Error(w, "Collect Frequency should be more than 150!", 400)
	} else if !validateSendFreq(config.SendFreq) {
		log.Error("Invalid Send Frequency Value")
		http.Error(w, "Send Frequency should be more than 150!", 400)
	} else {
		// Save New Configuration to DB
		SetFridgeConfig(dbClient,configInfo,config)
		log.Println("New Config was added to DB: ", config)
		JSONСonfig, _ := json.Marshal(config)
		go PublishConfigMessage(JSONСonfig, "configChan")
	}
}

// Validate MAC got from User
func validateMAC(mac interface{}) bool {
	switch v := mac.(type) {
	case string:
		switch len(v) {
		case 17:
			return true
		default:
			log.Error("MAC should contain 17 symbols")
			return false
		}
	default:
		log.Error("MAC should be in string format")
		return false
	}
}

// Validate Send Frequency Value got from User
func validateSendFreq(sendFreq interface{}) bool {
	switch v := sendFreq.(type) {
	case int64:
		switch {
		case v > 150:
			return true
		default:
			log.Error("Send Frequency should be more than 150!")
			return false
		}
	default:
		log.Error("Send Frequency should be in int64 format")
		return false
	}
}

// Validate Collect Frequency got from User
func validateCollectFreq(collectFreq interface{}) bool {
	switch v := collectFreq.(type) {
	case int64:
		switch {
		case v > 150:

			return true
		default:
			log.Error("Collect Frequency should be more than 150!")
			return false
		}
	default:
		log.Error("Collect Frequency should be in int64 format")
		return false
	}
}

// Validate TurnedOn Value got from User
func validateTurnedOn(turnedOn interface{}) bool {
	switch turnedOn.(type) {
	case bool:
		return true
	default:
		log.Error("TurnedOn should be in bool format!")
		return false
	}
}

// Validate StreamOn Value got from User
func validateStreamOn(streamOn interface{}) bool {
	switch streamOn.(type) {
	case bool:
		return true
	default:
		log.Error("StreamOn should be in bool format!")
		return false
	}
}

//-----------------Common functions-------------------------------------------------------------------------------------------

func CheckError(desc string, err error) error {
	if err != nil {
		log.Errorln(desc, err)
		return err
	}
	return nil
}

func Subscribe(client *redis.Client, roomID string, channel chan []string) {
	err := client.Connect(dbHost, dbPort)
	CheckError("Subscribe", err)
	go client.Subscribe(channel, roomID)
}
func PublishMessage(message interface{}, roomID string) {
	dbClient, _ := RunDBConnection()

	_, err := dbClient.Publish(roomID, message)
	CheckError("Publish", err)
}
func PublishConfigMessage(message []byte, roomID string) {
	dbClient, _ := RunDBConnection()
	_, err := dbClient.Publish(roomID, message)
	CheckError("Publish", err)
}

func Float32ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 32)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(int64(n), 10)
}
//-------------------Work with data base-------------------------------------------------------------------------------------------

func getAllDevices(dbClient *redis.Client) []DevData {
	var device DevData
	var devices []DevData

	devParamsKeys, err := dbClient.SMembers("devParamsKeys")
	if CheckError("Cant read members from devParamsKeys", err) != nil {
		return nil
	}

	var devParamsKeysTokens = make([][]string, len(devParamsKeys))
	for i, k := range devParamsKeys {
		devParamsKeysTokens[i] = strings.Split(k, ":")
	}

	for index, key := range devParamsKeysTokens {
		params, err := dbClient.SMembers(devParamsKeys[index])
		CheckError("Cant read members from "+devParamsKeys[index], err)

		device.Meta.Type = key[1]
		device.Meta.Name = key[2]
		device.Meta.MAC = key[3]
		device.Data = make(map[string][]string)

		values := make([][]string, len(params))
		for i, p := range params {
			values[i], _ = dbClient.ZRangeByScore(devParamsKeys[index]+":"+p, "-inf", "inf")
			CheckError("Cant use ZRangeByScore for "+devParamsKeys[index], err)
			device.Data[p] = values[i]
		}

		devices = append(devices, device)
	}
	return devices
}

func getDevice(devParamsKey string, devParamsKeysTokens []string) DevData {
	var device DevData
	dbClient, _ := RunDBConnection()

	params, err := dbClient.SMembers(devParamsKey)
	CheckError("Cant read members from devParamsKeys", err)
	device.Meta.Type = devParamsKeysTokens[1]
	device.Meta.Name = devParamsKeysTokens[2]
	device.Meta.MAC = devParamsKeysTokens[3]
	device.Data = make(map[string][]string)

	values := make([][]string, len(params))
	for i, p := range params {
		values[i], err = dbClient.ZRangeByScore(devParamsKey+":"+p, "-inf", "inf")
		CheckError("Cant use ZRangeByScore", err)
		device.Data[p] = values[i]
	}
	return device
}

func SetFridgeConfig(dbClient *redis.Client, configInfo string, config *DevConfig) {
	// Save default configuration to DB
	_, err := dbClient.HMSet(configInfo, "TurnedOn", config.TurnedOn)
	CheckError("DB error1: TurnedOn", err)
	_, err = dbClient.HMSet(configInfo, "CollectFreq", config.CollectFreq)
	CheckError("DB error2: CollectFreq", err)
	_, err = dbClient.HMSet(configInfo, "SendFreq", config.SendFreq)
	CheckError("DB error3: SendFreq", err)
	_, err = dbClient.HMSet(configInfo, "StreamOn", config.StreamOn)
	CheckError("DB error4: StreamOn", err)
}

func GetFridgeConfig(dbClient *redis.Client, configInfo, mac string) (*DevConfig) {
	var config DevConfig

	state, err := dbClient.HMGet(configInfo, "TurnedOn")
	CheckError("Get from DB error1: TurnedOn ", err)
	sendFreq, err := dbClient.HMGet(configInfo, "SendFreq")
	CheckError("Get from DB error2: SendFreq ", err)
	collectFreq, err := dbClient.HMGet(configInfo, "CollectFreq")
	CheckError("Get from DB error3: CollectFreq ", err)
	streamOn, err := dbClient.HMGet(configInfo, "StreamOn")
	CheckError("Get from DB error4: StreamOn ", err)

	stateBool, _ := strconv.ParseBool(strings.Join(state, " "))
	sendFreqInt, _ := strconv.Atoi(strings.Join(sendFreq, " "))
	collectFreqInt, _ := strconv.Atoi(strings.Join(collectFreq, " "))
	streamOnBool, _ := strconv.ParseBool(strings.Join(streamOn, " "))

	config = DevConfig{
		MAC: mac,
		TurnedOn:    stateBool,
		CollectFreq: int64(collectFreqInt),
		SendFreq:    int64(sendFreqInt),
		StreamOn:    streamOnBool,
	}

	log.Println("Configuration from DB: ", state, sendFreq, collectFreq)

	return &config
}
