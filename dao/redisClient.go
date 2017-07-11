package dao

import (
	"time"
	"errors"
	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"
	"menteslibres.net/gosexy/redis"

	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/sys"
)

const (
	partKeyToConfig = ":config"
)

type RedisClient struct {
	Client   *redis.Client
	DbServer Server
}

func (rc *RedisClient) SetDBServer(server Server){
	rc.DbServer = server
}

func (rc *RedisClient) FlushAll() error {
	_, err := rc.Client.FlushAll()
	return err
}

func (rc *RedisClient) GetClient() DbRedisDriver {
	return rc.Client
}

func (rc *RedisClient) Publish(channel string, message interface{}) (int64, error) {
	return rc.Client.Publish(channel, message)
}

func (rc *RedisClient) Connect() (error) {
	if rc.DbServer.IP == "" || rc.DbServer.Port == 0 {
		return errors.New("Empty value. Check Host or Port.")
	}
	rc.Client = redis.New()
	err := rc.Client.Connect(rc.DbServer.IP, rc.DbServer.Port)
	return err
}

func (rc *RedisClient)Subscribe(cn chan []string, channel ...string) error{
	return rc.Client.Subscribe(cn, channel...)
}

func (rc *RedisClient) Close() error {
	return rc.Client.Close()
}

func (rc RedisClient) NewDBConnection() (DbClient) {
	err := rc.Connect()
	CheckError("NewDBConnection error: ", err)
	for err != nil {
		log.Errorln("Database: connection has failed:", err)
		time.Sleep(time.Second)
		err = rc.Connect()
	}
	var dbClient DbClient = &rc
	return dbClient
}

func PublishWS(req Request, roomID string, worker DbClient) {
	pubReq, err := json.Marshal(req)
	CheckError("Marshal for publish.", err)

	worker.Connect()
	defer worker.Close()

	_, err = worker.Publish(roomID, pubReq)
	CheckError("Publish", err)
}

func (rc *RedisClient) GetAllDevices() []DevData {
	rc.NewDBConnection()

	var device DevData
	var devices []DevData

	devParamsKeys, err := rc.Client.SMembers("devParamsKeys")
	if CheckError("Can't read members from devParamsKeys", err) != nil {
		return nil
	}

	var devParamsKeysTokens = make([][]string, len(devParamsKeys))
	for i, k := range devParamsKeys {
		devParamsKeysTokens[i] = strings.Split(k, ":")
	}

	for index, key := range devParamsKeysTokens {
		params, err := rc.Client.SMembers(devParamsKeys[index])
		CheckError("Can't read members from "+devParamsKeys[index], err)

		device.Meta.Type = key[1]
		device.Meta.Name = key[2]
		device.Meta.MAC = key[3]
		device.Data = make(map[string][]string)

		values := make([][]string, len(params))
		for i, p := range params {
			values[i], _ = rc.Client.ZRangeByScore(devParamsKeys[index]+":"+p, "-inf", "inf")
			CheckError("Can't use ZRangeByScore for "+devParamsKeys[index], err)
			device.Data[p] = values[i]
		}

		devices = append(devices, device)
	}
	return devices
}

func (rc *RedisClient) GetKeyForConfig(mac string) string {
	return mac + partKeyToConfig
}


func (rc *RedisClient) Multi() (string, error)   {
	return rc.Client.Multi()
}
func (rc *RedisClient) Discard()  (string, error)  {
	return rc.Client.Discard()
}
func (rc *RedisClient) Exec()  ([]interface{}, error) {
	return rc.Client.Exec()
}