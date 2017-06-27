package dao

import (
	"menteslibres.net/gosexy/redis"
	"fmt"
)

type MyRedis struct {
	Client *redis.Client
	count int
}

func (myRedis *MyRedis) FlushAll() error {
	_, err := myRedis.Client.FlushAll()
	return err
}
func (myRedis *MyRedis)Publish(channel string, message interface{}) (int64, error){
	return myRedis.Client.Publish(channel,message)
}

func (myRedis *MyRedis)Connect(host string, port uint) ( error) {
	myRedis.Client = redis.New()
	err := myRedis.Client.Connect(host, port)
	return  err
}
func (myRedis *MyRedis) Subscribe(cn chan []string, channel ...string) error {
	return myRedis.Client.Subscribe(cn, channel...)
}

func (myRedis *MyRedis) Count(){
	myRedis.count += 1
	fmt.Println(myRedis.count)
}
func (myRedis *MyRedis) Close() error{
	return  myRedis.Client.Close()
}





