package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"

)

type DevConfigDriver interface {
	GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver)
	ValidateDevData(config DevConfig) (bool, string)
	GetDefaultConfig() (*DevConfig)
	CheckDevConfig(arr []byte, configInfo, mac string, client DbDriver)([]byte)
}

type DevDataDriver interface {
	GetDevData(devParamsKey string, devParamsKeysTokens DevMeta, worker DbRedisDriver) DevData
	SetDevData(req *Request, worker DbRedisDriver) *ServerError
}
