package devices

import (
	"strconv"
	"strings"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/sys"
)

type Fridge struct {
	Data   FridgeData
	Config DevConfig
	Meta   DevMeta
}
type FridgeData struct {
	TempCam1 map[int64]float32 `json:"tempCam1"`
	TempCam2 map[int64]float32 `json:"tempCam2"`
}

func (fridge *Fridge) GetDevData(devParamsKey string, devParamsKeysTokens []string, worker DbRedisDriver) DevData {
	var device DevData

	params, err := worker.SMembers(devParamsKey)
	CheckError("Cant read members from devParamsKeys", err)
	device.Meta.Type = devParamsKeysTokens[1]
	device.Meta.Name = devParamsKeysTokens[2]
	device.Meta.MAC = devParamsKeysTokens[3]
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
	defer worker.Close()

	var devData FridgeData
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	devKey := "device" + ":" + devType + ":" + devName + ":" + mac
	devParamsKey := devKey + ":" + "params"

	_, err := worker.SAdd("devParamsKeys", devParamsKey)
	CheckError("DB error11", err)
	_, err = worker.HMSet(devKey, "ReqTime", devReqTime)
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

func setDevData(TempCam map[int64]float32, key string, worker DbRedisDriver) error {
	for time, value := range TempCam {
		_, err := worker.ZAdd(key, Int64ToString(time), Int64ToString(time)+":"+Float32ToString(value))
		if CheckError("DB error14", err) != nil {
			return err
		}
	}
	return nil
}

func (fridge *Fridge) GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig) {
	var config DevConfig

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

	config = DevConfig{
		MAC:         mac,
		TurnedOn:    stateBool,
		CollectFreq: int64(collectFreqInt),
		SendFreq:    int64(sendFreqInt),
		StreamOn:    streamOnBool,
	}

	log.Println("Configuration from DB: ", state, sendFreq, collectFreq)

	return &config
}

func (fridge *Fridge) SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver) {
	_, err := worker.HMSet(configInfo, "TurnedOn", config.TurnedOn)
	CheckError("DB error1: TurnedOn", err)
	_, err = worker.HMSet(configInfo, "CollectFreq", config.CollectFreq)
	CheckError("DB error2: CollectFreq", err)
	_, err = worker.HMSet(configInfo, "SendFreq", config.SendFreq)
	CheckError("DB error3: SendFreq", err)
	_, err = worker.HMSet(configInfo, "StreamOn", config.StreamOn)
	CheckError("DB error4: StreamOn", err)
}

func (fridge *Fridge) ValidateDevData(config DevConfig) (bool, string) {
	if !ValidateMAC(config.MAC) {
		log.Error("Invalid MAC")
		return false, "Invalid MAC"
	} else if !ValidateStreamOn(config.StreamOn) {
		log.Error("Invalid Stream Value")
		return false, "Invalid Stream Value"
	} else if !ValidateTurnedOn(config.TurnedOn) {
		log.Error("Invalid Turned Value")
		return false, "Invalid Turned Value"
	} else if !ValidateCollectFreq(config.CollectFreq) {
		log.Error("Invalid Collect Frequency Value")
		return false, "Collect Frequency should be more than 150!"
	} else if !ValidateSendFreq(config.SendFreq) {
		log.Error("Invalid Send Frequency Value")
		return false, "Send Frequency should be more than 150!"
	}
	return true, ""
}
