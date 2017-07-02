package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"

)

type ConfigDevDriver interface {
	GetDevConfig(configInfo, mac string) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig)
}

type DataDevDriver interface {
	GetDevData(devParamsKey string, devParamsKeysTokens []string, worker RedisInteractor) DevData
	SetDevData(req *Request, worker RedisInteractor) *ServerError
}