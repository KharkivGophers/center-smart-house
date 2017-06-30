package devices

import (
	"strconv"
	"strings"
	"log"
	"encoding/json"
	"github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/models"
)

func (server *TCPDataServer) SetDevData(req *Request) *ServerError {
	myRedis, err := dao.MyRedis{Host: server.DbServer.IP, Port: server.DbServer.Port}.RunDBConnection()
	defer myRedis.Close()

	var devData FridgeData
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	devKey := "device" + ":" + devType + ":" + devName + ":" + mac
	devParamsKey := devKey + ":" + "params"

	_, err = myRedis.Client.SAdd("devParamsKeys", devParamsKey)
	CheckError("DB error11", err)
	_, err = myRedis.Client.HMSet(devKey, "ReqTime", devReqTime)
	CheckError("DB error12", err)
	_, err = myRedis.Client.SAdd(devParamsKey, "TempCam1", "TempCam2")
	CheckError("DB error13", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		log.Error("Error in fridgedatahandler")
		return &ServerError{Error: err}
	}

	for time, value := range devData.TempCam1 {
		_, err := myRedis.Client.ZAdd(devParamsKey+":"+"TempCam1",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error14", err) != nil {
			return &ServerError{Error: err}
		}
	}

	for time, value := range devData.TempCam2 {
		_, err := myRedis.Client.ZAdd(devParamsKey+":"+"TempCam2",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error15", err) != nil {
			return &ServerError{Error: err}
		}
	}

	return nil
}

func (myRedis *MyRedis) GetDevConfig(configInfo, mac string) (*DevConfig) {
	var config DevConfig

	state, err := myRedis.Client.HMGet(configInfo, "TurnedOn")
	CheckError("Get from DB error1: TurnedOn ", err)
	sendFreq, err := myRedis.Client.HMGet(configInfo, "SendFreq")
	CheckError("Get from DB error2: SendFreq ", err)
	collectFreq, err := myRedis.Client.HMGet(configInfo, "CollectFreq")
	CheckError("Get from DB error3: CollectFreq ", err)
	streamOn, err := myRedis.Client.HMGet(configInfo, "StreamOn")
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

func  (myRedis *MyRedis) SetDevConfig(configInfo string, config *DevConfig) {
	_, err := myRedis.Client.HMSet(configInfo, "TurnedOn", config.TurnedOn)
	CheckError("DB error1: TurnedOn", err)
	_, err = myRedis.Client.HMSet(configInfo, "CollectFreq", config.CollectFreq)
	CheckError("DB error2: CollectFreq", err)
	_, err = myRedis.Client.HMSet(configInfo, "SendFreq", config.SendFreq)
	CheckError("DB error3: SendFreq", err)
	_, err = myRedis.Client.HMSet(configInfo, "StreamOn", config.StreamOn)
	CheckError("DB error4: StreamOn", err)
}
