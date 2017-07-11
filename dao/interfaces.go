package dao
import (
	. "github.com/KharkivGophers/center-smart-house/models"

)

// Abstract database interface.
type DbClient interface {
	FlushAll() (error)
	Publish(channel string, message interface{}) (int64, error)
	Connect()(error)
	Subscribe(cn chan []string, channel ...string) error
	Close() (error)
	NewDBConnection()(DbClient)
	GetAllDevices() ([]DevData)
	GetClient() DbRedisDriver

	GetKeyForConfig(mac string)string
	SetDBServer(server Server)

	// Not yet implemented.
	//GetDevConfig(configInfo, devType string, mac string) (*DevConfig)
	//SetDevConfig(configInfo string, devType string, config *DevConfig)

	// Only to Fridge. Must be refactored
	//GetFridgeConfig(configInfo, mac string) (*DevConfig)
	//SetFridgeConfig(configInfo string, config *DevConfig)
}

// Concrete redis database interface.
type DbRedisDriver interface {
	SAdd(key string, member ...interface{}) (int64, error)
	ZAdd(key string, arguments ...interface{}) (int64, error)

	HMSet(key string, values ...interface{}) (string, error)
	HMGet(key string, fields ...string) ([]string, error)

	ZRangeByScore(key string, values ...interface{}) ([]string, error)
	SMembers(key string) ([]string, error)

	Close() error
	Connect(host string, port uint) (err error)

	Exists(key string) (bool, error)

	Multi() (string, error)
	Discard()  (string, error)
	Exec()  ([]interface{}, error)
}
