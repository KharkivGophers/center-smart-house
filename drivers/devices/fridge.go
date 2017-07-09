package devices

import (
	"strconv"
	"strings"
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/sys"
	"net"
)

type Fridge struct {
	Data   FridgeData `json:"data"`
	Config FridgeConfig
	Meta   DevMeta    `json:"meta"`
}
type FridgeData struct {
	TempCam1 map[int64]float32 `json:"tempCam1"`
	TempCam2 map[int64]float32 `json:"tempCam2"`
}

type FridgeConfig struct {
	TurnedOn    bool   `json:"turnedOn"`
	StreamOn    bool   `json:"streamOn"`
	CollectFreq int64  `json:"collectFreq"`
	SendFreq    int64  `json:"sendFreq"`
}

//--------------------------------------DevDataDriver--------------------------------------------------------------
func (fridge *Fridge) GetDevData(devParamsKey string, devMeta DevMeta, worker DbRedisDriver) DevData {
	var device DevData

	params, err := worker.SMembers(devParamsKey)
	CheckError("Cant read members from devParamsKeys", err)
	device.Meta = devMeta
	device.Data = make(map[string][]string)

	values := make([][]string, len(params))
	for i, p := range params {
		values[i], err = worker.ZRangeByScore(devParamsKey+":"+p, "-inf", "inf")
		CheckError("Cant use ZRangeByScore", err)
		device.Data[p] = values[i]
	}
	return device
}

func (fridge *Fridge) SetDevData(req *Request, worker DbRedisDriver) *ServerError {

	var devData FridgeData

	devKey := "device" + ":" + req.Meta.Type + ":" + req.Meta.Name + ":" + req.Meta.MAC
	devParamsKey := devKey + ":" + "params"

	_, err := worker.SAdd("devParamsKeys", devParamsKey)
	CheckError("DB error11", err)
	_, err = worker.HMSet(devKey, "ReqTime", req.Time)
	CheckError("DB error12", err)
	_, err = worker.SAdd(devParamsKey, "TempCam1", "TempCam2")
	CheckError("DB error13", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		log.Error("Error in fridgedatahandler")
		return &ServerError{Error: err}
	}

	err = setDevData(devData.TempCam1, devParamsKey+":"+"TempCam1", worker)
	if CheckError("DB error14", err) != nil {
		return &ServerError{Error: err}
	}

	err = setDevData(devData.TempCam2, devParamsKey+":"+"TempCam2", worker)
	if CheckError("DB error14", err) != nil {
		return &ServerError{Error: err}
	}

	return nil
}

//--------------------------------------DevConfigDriver--------------------------------------------------------------

func setDevData(TempCam map[int64]float32, key string, worker DbRedisDriver) error {
	for time, value := range TempCam {
		_, err := worker.ZAdd(key, Int64ToString(time), Int64ToString(time)+":"+Float64ToString(value))
		if CheckError("DB error14", err) != nil {
			return err
		}
	}
	return nil
}

func (fridge *Fridge) GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig) {
	var fridgeConfig FridgeConfig = *fridge.getFridgeConfig(configInfo, mac, worker)
	var devConfig DevConfig

	arrByte, err := json.Marshal(&fridgeConfig)
	CheckError("GetDevConfig. Exception in the marshal() ", err)

	devConfig = DevConfig{
		MAC:  mac,
		Data: arrByte,
	}

	log.Println("Configuration from DB: ", fridgeConfig.TurnedOn, fridgeConfig.SendFreq, fridgeConfig.CollectFreq)
	return &devConfig
}

func (fridge *Fridge) getFridgeConfig(configInfo, mac string, worker DbRedisDriver) (*FridgeConfig) {
	state, err := worker.HMGet(configInfo, "TurnedOn")
	CheckError("Get from DB error1: TurnedOn ", err)
	sendFreq, err := worker.HMGet(configInfo, "SendFreq")
	CheckError("Get from DB error2: SendFreq ", err)
	collectFreq, err := worker.HMGet(configInfo, "CollectFreq")
	CheckError("Get from DB error3: CollectFreq ", err)
	streamOn, err := worker.HMGet(configInfo, "StreamOn")
	CheckError("Get from DB error4: StreamOn ", err)

	stateBool, _ := strconv.ParseBool(strings.Join(state, " "))
	sendFreqInt, _ := strconv.Atoi(strings.Join(sendFreq, " "))
	collectFreqInt, _ := strconv.Atoi(strings.Join(collectFreq, " "))
	streamOnBool, _ := strconv.ParseBool(strings.Join(streamOn, " "))

	return &FridgeConfig{
		TurnedOn:    stateBool,
		CollectFreq: int64(collectFreqInt),
		SendFreq:    int64(sendFreqInt),
		StreamOn:    streamOnBool,
	}
}

func (fridge *Fridge) SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver) {
	var fridgeConfig FridgeConfig
	json.Unmarshal(config.Data, &fridgeConfig)

	_, err := worker.HMSet(configInfo, "TurnedOn", fridgeConfig.TurnedOn)
	CheckError("DB error1: TurnedOn", err)
	_, err = worker.HMSet(configInfo, "CollectFreq", fridgeConfig.CollectFreq)
	CheckError("DB error2: CollectFreq", err)
	_, err = worker.HMSet(configInfo, "SendFreq", fridgeConfig.SendFreq)
	CheckError("DB error3: SendFreq", err)
	_, err = worker.HMSet(configInfo, "StreamOn", fridgeConfig.StreamOn)
	CheckError("DB error4: StreamOn", err)
}

func (fridge *Fridge) ValidateDevData(config DevConfig) (bool, string) {
	var fridgeConfig FridgeConfig
	json.Unmarshal(config.Data, &fridgeConfig)

	if !ValidateMAC(config.MAC) {
		log.Error("Invalid MAC")
		return false, "Invalid MAC"
	} else if !ValidateStreamOn(fridgeConfig.StreamOn) {
		log.Error("Invalid Stream Value")
		return false, "Invalid Stream Value"
	} else if !ValidateTurnedOn(fridgeConfig.TurnedOn) {
		log.Error("Invalid Turned Value")
		return false, "Invalid Turned Value"
	} else if !ValidateCollectFreq(fridgeConfig.CollectFreq) {
		log.Error("Invalid Collect Frequency Value")
		return false, "Collect Frequency should be more than 150!"
	} else if !ValidateSendFreq(fridgeConfig.SendFreq) {
		log.Error("Invalid Send Frequency Value")
		return false, "Send Frequency should be more than 150!"
	}
	return true, ""
}

func (fridge *Fridge) GetDefaultConfig() (*DevConfig) {
	config := FridgeConfig{
		TurnedOn:    true,
		StreamOn:    true,
		CollectFreq: 1000,
		SendFreq:    5000,
	}

	arrByte, err := json.Marshal(config)
	CheckError("GetDevConfig. Exception in the marshal() ", err)

	return &DevConfig{
		MAC:  fridge.Meta.MAC,
		Data: arrByte,
	}
}

func (fridge *Fridge) CheckDevConfigAndMarshal(arr []byte, configInfo, mac string, client DbDriver) ([]byte) {
	fridgeConfig := fridge.getFridgeConfig(configInfo, mac, client.GetClient())
	json.Unmarshal(arr, &fridgeConfig)
	arr, _ = json.Marshal(fridgeConfig)
	return arr
}

//--------------------------------------DevServerHandler--------------------------------------------------------------
func (fridge *Fridge) GetDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta, client DbDriver) {

}

func (fridge *Fridge) SendDefaultConfigurationTCP(conn net.Conn, dbClient DbDriver, req *Request) ([]byte) {
	var config *DevConfig
	configInfo := req.Meta.MAC + ":" + "config" // key
	if ok, _ := dbClient.GetClient().Exists(configInfo); ok {

		config = fridge.GetDevConfig(configInfo, req.Meta.MAC, dbClient.GetClient())
		log.Println("Old Device with MAC: ", req.Meta.MAC, "detected.")

	} else {
		log.Warningln("New Device with MAC: ", req.Meta.MAC, "detected.")
		log.Warningln("Default Config will be sent.")
		config = fridge.GetDefaultConfig()

		fridge.SetDevConfig(configInfo, config, dbClient.GetClient())
	}

	return config.Data
}

func (fridge *Fridge) PatchDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, devMeta DevMeta, dbClient DbDriver) {
	var config *DevConfig
	configInfo := devMeta.MAC + ":" + "config" // key

	err := json.NewDecoder(r.Body).Decode(&config)

	config.Data = fridge.CheckDevConfigAndMarshal(config.Data, configInfo, devMeta.MAC, dbClient)

	if err != nil {
		http.Error(w, err.Error(), 400)
		log.Errorln("NewDec: ", err)
	}

	valid, message := fridge.ValidateDevData(*config)
	if !valid {
		http.Error(w, message, 400)
	} else {
		// Save New Configuration to DB
		fridge.SetDevConfig(configInfo, config, dbClient.GetClient())
		log.Println("New Config was added to DB: ", config.MAC)
		JSONСonfig, _ := json.Marshal(config)
		dbClient.Publish("configChan", JSONСonfig)
	}
}
