package drivers

import (
	. "github.com/KharkivGophers/center-smart-house/models"
)

type ConfigDevDriver interface {
	GetDevConfig(configInfo, mac string) (*DevConfig)
	SetDevConfig(configInfo string, config *DevConfig)
}

type DataDevDriver interface {
	GetDevData(configInfo, mac string) (*DevConfig)
	SetDevData(configInfo string, config *DevConfig)
}