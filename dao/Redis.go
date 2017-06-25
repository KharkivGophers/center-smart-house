package dao

import (
	"menteslibres.net/gosexy/redis"
)

type MyRedis struct {
	Client *redis.Client
}
func (myRedis MyRedis) FlushAll() error {
	_, err := myRedis.Client.FlushAll()
	return err
}
func (myRedis MyRedis)Publish(channel string, message interface{}) (int64, error){
	return myRedis.Client.Publish(channel,message)
}
func (myRedis MyRedis) NewClient() *MyRedis {
	myRedis.Client = redis.New()
	return  &myRedis
}
func (myRedis MyRedis)Connect(host string, port uint) (*MyRedis, error) {
	myRedis.Client = redis.New()
	err := myRedis.Client.Connect(host, port)
	return &myRedis, err
}
func (myRedis MyRedis) Subscribe(cn chan []string, channel ...string) error {
	return myRedis.Client.Subscribe(cn, channel...)
}