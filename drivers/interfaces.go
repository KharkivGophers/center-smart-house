package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"
	"net/http"

)

type DevConfigDriver interface {
	GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver)
	ValidateDevData(config DevConfig) (bool, string)
	GetDefaultConfig() (*DevConfig)
	CheckDevConfigAndMarshal(arr []byte, configInfo, mac string, client DbDriver)([]byte)
}

type DevDataDriver interface {
	GetDevData(devParamsKey string, devMeta DevMeta, worker DbRedisDriver) DevData
	SetDevData(req *Request, worker DbRedisDriver) *ServerError
}
//Idea: Use this interface in the server. Than we give an opportunity to produce realization work logic samself
type DevServerHandler interface{
	GetDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta)
}
