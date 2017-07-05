package dao

import (
	"menteslibres.net/gosexy/redis"
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/sys"
	log "github.com/Sirupsen/logrus"


	"time"
	"errors"
	"encoding/json"
	"strings"
)

type MyRedis struct {
	Client *redis.Client
	Host   string
	Port   uint
}

func (myRedis *MyRedis) FlushAll () error {
	_, err := myRedis.Client.FlushAll()
	return err
}

func (myRedis *MyRedis) GetClient() RedisClient{
	return  myRedis.Client
}

func (myRedis *MyRedis)Publish (channel string, message interface{}) (int64, error){
	return myRedis.Client.Publish(channel,message)
}

func (myRedis *MyRedis)Connect() (error) {
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

func (myRedis *MyRedis) RunDBConnection() (error) {
	err := myRedis.Connect()
	CheckError("RunDBConnection error: ", err)
	for err != nil {
		log.Errorln("Database: connection has failed:", err)
		time.Sleep(time.Second)
		err = myRedis.Connect()
	}

	return err
}

func  PublishWS(req Request, roomID string, worker DbClient) {
	pubReq, err := json.Marshal(req)
	CheckError("Marshal for publish.", err)

	worker.Connect()
	defer worker.Close()

	_, err = worker.Publish(roomID, pubReq)
	CheckError("Publish", err)
}

func (myRedis *MyRedis) GetAllDevices() []DevData {
	myRedis.RunDBConnection()

	var device DevData
	var devices []DevData

	devParamsKeys, err := myRedis.Client.SMembers("devParamsKeys")
	if CheckError("Can't read members from devParamsKeys", err) != nil {
		return nil
	}

	var devParamsKeysTokens = make([][]string, len(devParamsKeys))
	for i, k := range devParamsKeys {
		devParamsKeysTokens[i] = strings.Split(k, ":")
	}

	for index, key := range devParamsKeysTokens {
		params, err := myRedis.Client.SMembers(devParamsKeys[index])
		CheckError("Can't read members from "+devParamsKeys[index], err)

		device.Meta.Type = key[1]
		device.Meta.Name = key[2]
		device.Meta.MAC = key[3]
		device.Data = make(map[string][]string)

		values := make([][]string, len(params))
		for i, p := range params {
			values[i], _ = myRedis.Client.ZRangeByScore(devParamsKeys[index]+":"+p, "-inf", "inf")
			CheckError("Can't use ZRangeByScore for "+devParamsKeys[index], err)
			device.Data[p] = values[i]
		}

		devices = append(devices, device)
	}
	return devices
}

func (myRedis *MyRedis) GetDevice(devParamsKey string, devParamsKeysTokens []string) DevData {
	myRedis.RunDBConnection()

	var device DevData

	params, err := myRedis.Client.SMembers(devParamsKey)
	CheckError("Can't read members from devParamsKeys", err)
	device.Meta.Type = devParamsKeysTokens[1]
	device.Meta.Name = devParamsKeysTokens[2]
	device.Meta.MAC = devParamsKeysTokens[3]
	device.Data = make(map[string][]string)

	values := make([][]string, len(params))
	for i, p := range params {
		values[i], err = myRedis.Client.ZRangeByScore(devParamsKey+":"+p, "-inf", "inf")
		CheckError("Can't use ZRangeByScore", err)
		device.Data[p] = values[i]
	}
	return device
}

func GetDBConnection(db Server) (DbClient) {
	client := &MyRedis{Host: db.IP, Port: db.Port}
	client.RunDBConnection()
	return client
}