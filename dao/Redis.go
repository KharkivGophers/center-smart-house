package dao

import (
	"menteslibres.net/gosexy/redis"
	. "github.com/KharkivGophers/center-smart-house/common"
	. "github.com/KharkivGophers/center-smart-house/common/models"
	log "github.com/Sirupsen/logrus"


	"time"
	"errors"
	"encoding/json"
	"strings"
	"strconv"
)

type MyRedis struct {
	Client *redis.Client
	Host   string
	Port   uint
}

func (myRedis *MyRedis) FlushAll() error {
	_, err := myRedis.Client.FlushAll()
	return err
}
func (myRedis *MyRedis)Publish(channel string, message interface{}) (int64, error){
	return myRedis.Client.Publish(channel,message)
}

func (myRedis *MyRedis)Connect()(error) {
	if myRedis.Host == "" || myRedis.Port == 0{
		return errors.New("Empty value. Check Host or Port.")
	}
	myRedis.Client = redis.New()
	err := myRedis.Client.Connect(myRedis.Host, myRedis.Port)
	return  err
}

func (myRedis *MyRedis) Subscribe(cn chan []string, channel ...string) error {
	return myRedis.Client.Subscribe(cn, channel...)
}

func (myRedis *MyRedis) Close() error{
	return  myRedis.Client.Close()
}


func (myRedis MyRedis)RunDBConnection() (*MyRedis, error) {
	err := myRedis.Connect()
	CheckError("RunDBConnection error: ", err)
	for err != nil {
		log.Errorln("1 Database: connection has failed:", err)
		time.Sleep(time.Second)
		err = myRedis.Connect()
		if err == nil {
			continue
		}
		log.Errorln("err not nil")
	}

	return  &myRedis, err
}


func  PublishWS(req Request, roomID string, worker DbWorker ) {
	pubReq, err := json.Marshal(req)
	CheckError("Marshal for publish.", err)

	dbClient, _ := worker.RunDBConnection()
	defer dbClient.Close()

	_, err = dbClient.Publish(roomID, pubReq)
	CheckError("Publish", err)


}

func (myRedis *MyRedis)GetAllDevices() []DevData {
	var device DevData
	var devices []DevData

	devParamsKeys, err := myRedis.Client.SMembers("devParamsKeys")
	if CheckError("Cant read members from devParamsKeys", err) != nil {
		return nil
	}

	var devParamsKeysTokens = make([][]string, len(devParamsKeys))
	for i, k := range devParamsKeys {
		devParamsKeysTokens[i] = strings.Split(k, ":")
	}

	for index, key := range devParamsKeysTokens {
		params, err := myRedis.Client.SMembers(devParamsKeys[index])
		CheckError("Cant read members from "+devParamsKeys[index], err)

		device.Meta.Type = key[1]
		device.Meta.Name = key[2]
		device.Meta.MAC = key[3]
		device.Data = make(map[string][]string)

		values := make([][]string, len(params))
		for i, p := range params {
			values[i], _ = myRedis.Client.ZRangeByScore(devParamsKeys[index]+":"+p, "-inf", "inf")
			CheckError("Cant use ZRangeByScore for "+devParamsKeys[index], err)
			device.Data[p] = values[i]
		}

		devices = append(devices, device)
	}
	return devices
}

func (myRedis *MyRedis) GetDevice(devParamsKey string, devParamsKeysTokens []string) DevData {
	var device DevData

	params, err := myRedis.Client.SMembers(devParamsKey)
	CheckError("Cant read members from devParamsKeys", err)
	device.Meta.Type = devParamsKeysTokens[1]
	device.Meta.Name = devParamsKeysTokens[2]
	device.Meta.MAC = devParamsKeysTokens[3]
	device.Data = make(map[string][]string)

	values := make([][]string, len(params))
	for i, p := range params {
		values[i], err = myRedis.Client.ZRangeByScore(devParamsKey+":"+p, "-inf", "inf")
		CheckError("Cant use ZRangeByScore", err)
		device.Data[p] = values[i]
	}
	return device
}










// Only to Fridge. Must be refactored-----------------------------------------------------------------------

func (myRedis *MyRedis) GetFridgeConfig(configInfo, mac string) (*DevConfig) {
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

func  (myRedis *MyRedis) SetFridgeConfig(configInfo string, config *DevConfig) {
	_, err := myRedis.Client.HMSet(configInfo, "TurnedOn", config.TurnedOn)
	CheckError("DB error1: TurnedOn", err)
	_, err = myRedis.Client.HMSet(configInfo, "CollectFreq", config.CollectFreq)
	CheckError("DB error2: CollectFreq", err)
	_, err = myRedis.Client.HMSet(configInfo, "SendFreq", config.SendFreq)
	CheckError("DB error3: SendFreq", err)
	_, err = myRedis.Client.HMSet(configInfo, "StreamOn", config.StreamOn)
	CheckError("DB error4: StreamOn", err)
}