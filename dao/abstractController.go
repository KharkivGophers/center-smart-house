package dao
import (
	. "github.com/KharkivGophers/center-smart-house/models"

)
type DbWorker interface {
	FlushAll() (error)
	Publish(channel string, message interface{}) (int64, error)
	Connect()(error)
	Subscribe(cn chan []string, channel ...string) error
	Close() (error)
	RunDBConnection() (*MyRedis, error)
	GetAllDevices() ([]DevData)
	GetDevice(devParamsKey string, devParamsKeysTokens []string) (DevData)

	// Not yet implemented
	GetDevConfig(configInfo, devType string, mac string) (*DevConfig)
	SetDevConfig(configInfo string, devType string, config *DevConfig)

	// Only to Fridge. Must be refactored
	GetFridgeConfig(configInfo, mac string) (*DevConfig)
	SetFridgeConfig(configInfo string, config *DevConfig)
}
