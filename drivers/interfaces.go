package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"
	"net/http"

	"net"
)

type DevConfigDriver interface {
	GetDevConfig(configInfo, mac string, client DbClient) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig, client DbClient)
	GetDefaultConfig() (*DevConfig)
}

type DevDataDriver interface {
	GetDevData(devParamsKey string, devMeta DevMeta, client DbClient) DevData
	SetDevData(req *Request, worker DbClient) *ServerError
}
type DevServerHandler interface{
	SendDefaultConfigurationTCP(conn net.Conn, dbClient DbClient, req *Request)([]byte)
	PatchDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta, client DbClient)
}
type DevDriver interface{
	DevDataDriver
	DevServerHandler
	DevConfigDriver
}