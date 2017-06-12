package dao

import "menteslibres.net/gosexy/redis"

type MyRedis struct {
	Client *redis.Client
}
func (redis MyRedis) FlushAll() error {
	_, err := redis.Client.FlushAll()
	return err
}
