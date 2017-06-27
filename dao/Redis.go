package dao

import (
	"menteslibres.net/gosexy/redis"
	"github.com/KharkivGophers/center-smart-house/server/common"
	log "github.com/Sirupsen/logrus"


	"time"
	"errors"
)

type MyRedis struct {
	Client *redis.Client
	host string
	port uint
}

func (myRedis *MyRedis) FlushAll() error {
	_, err := myRedis.Client.FlushAll()
	return err
}
func (myRedis *MyRedis)Publish(channel string, message interface{}) (int64, error){
	return myRedis.Client.Publish(channel,message)
}

func (myRedis *MyRedis)Connect()(error) {
	if myRedis.host == "" || myRedis.port == 0{
		return errors.New("Empty value. Check host or port.")
	}
	myRedis.Client = redis.New()
	err := myRedis.Client.Connect(myRedis.host, myRedis.port)
	return  err
}

func (myRedis *MyRedis) Subscribe(cn chan []string, channel ...string) error {
	return myRedis.Client.Subscribe(cn, channel...)
}

func (myRedis *MyRedis) Close() error{
	return  myRedis.Client.Close()
}


func (myRedis *MyRedis)RunDBConnection() (*MyRedis, error) {
	err := myRedis.Connect()
	common.CheckError("RunDBConnection error: ", err)
	for err != nil {
		log.Errorln("1 Database: connection has failed:", err)
		time.Sleep(time.Second)
		err = myRedis.Connect()
		if err == nil {
			continue
		}
		log.Errorln("err not nil")
	}

	return  myRedis, err
}



