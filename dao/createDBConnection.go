package dao

import (
	. "github.com/KharkivGophers/center-smart-house/models"
)

func GetDBConnection(db Server) (DbDriver) {
	client := &RedisClient{DbServer: db}
	client.RunDBConnection()
	return client
}
//func GetDBConnection(db Server) (DbDriver) {
//	client := &DBMock{DbServer: db}
//	client.RunDBConnection()
//	return client
//}
