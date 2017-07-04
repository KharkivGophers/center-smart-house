package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"

)

type ConfigDevDriver interface {
	GetDevConfig(configInfo, mac string, worker RedisClient) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig, worker RedisClient)
}

type DataDevDriver interface {
	GetDevData(devParamsKey string, devParamsKeysTokens []string, worker RedisClient) DevData
	SetDevData(req *Request, worker RedisClient) *ServerError
}
