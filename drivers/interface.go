package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"

)

type ConfigDevDriver interface {
	GetDevConfig(configInfo, mac string, worker RedisInteractor) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig, worker RedisInteractor)
}

type DataDevDriver interface {
	GetDevData(devParamsKey string, devParamsKeysTokens []string, worker RedisInteractor) DevData
	SetDevData(req *Request, worker RedisInteractor) *ServerError
}