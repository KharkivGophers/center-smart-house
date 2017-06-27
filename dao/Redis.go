package dao

import (
	"menteslibres.net/gosexy/redis"
	. "github.com/KharkivGophers/center-smart-house/server/common"
	"github.com/KharkivGophers/center-smart-house/server/common/models"
	log "github.com/Sirupsen/logrus"


	"time"
	"errors"
	"encoding/json"
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


func  PublishWS(req models.Request, roomID, host string, port uint ) {
	pubReq, err := json.Marshal(req)
	CheckError("Marshal for publish.", err)

	dbClient, _ := MyRedis{Host:host, Port:port}.RunDBConnection()
	defer dbClient.Close()

	_, err = dbClient.Publish(roomID, pubReq)
	CheckError("Publish", err)


}


