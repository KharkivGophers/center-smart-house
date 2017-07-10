package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"
	"net/http"

	"net"
)

type DevConfigDriver interface {
	GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver)
	ValidateDevData(config DevConfig) (bool, string)
	GetDefaultConfig() (*DevConfig)
	CheckDevConfigAndMarshal(arr []byte, configInfo, mac string, client DbClient)([]byte)
}

type DevDataDriver interface {
	GetDevData(devParamsKey string, devMeta DevMeta, worker DbRedisDriver) DevData
	SetDevData(req *Request, worker DbRedisDriver) *ServerError
}
//Idea: Use this interface in the server. Than we give an opportunity to produce realization work logic samself
type DevServerHandler interface{
	GetDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta, client DbClient)
	SendDefaultConfigurationTCP(conn net.Conn, dbClient DbClient, req *Request)([]byte)
	PatchDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta, client DbClient)
}
type DevDriver interface{
	DevDataDriver
	DevServerHandler
	DevConfigDriver
}