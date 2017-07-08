package devices

import (
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/models"
)

type Washer struct {
	Data   WasherData
	Config DevConfig
	Meta   DevMeta
}

type WasherData struct {
	Mode   string
	Drying string
	Temp   map[int64]float32
}

type WasherConfig struct {
	MAC            string   `json:"mac"`
	WashTime       int64    `json:"washTime"`
	WashTurnovers  int64    `json:"washTurnovers"`
	RinseTime      int64    `json:"rinseTime"`
	RinseTurnovers int64    `json:"rinseTurnovers"`
	SpinTime       int64    `json:"spinTime"`
	SpinTurnovers  int64    `json:"spinTurnovers"`
}

var (
	LightMode WasherConfig = WasherConfig{
		WashTime:       90,
		WashTurnovers:  240,
		RinseTime:      30,
		RinseTurnovers: 120,
		SpinTime:       30,
		SpinTurnovers:  60,
	}
	FastMode WasherConfig = WasherConfig{
		WashTime:       30,
		WashTurnovers:  300,
		RinseTime:      15,
		RinseTurnovers: 240,
		SpinTime:       15,
		SpinTurnovers:  60,
	}
	StandartMode WasherConfig = WasherConfig{
		WashTime:       120,
		WashTurnovers:  240,
		RinseTime:      60,
		RinseTurnovers: 180,
		SpinTime:       60,
		SpinTurnovers:  60,
	}
)

func (washer *Washer) GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig)  { return &DevConfig{} }
func (washer *Washer) SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver) {}
func (washer *Washer) ValidateDevData(config DevConfig) (bool, string)                         { return true, "" }
func (washer *Washer) GetDefaultConfig() (*DevConfig)                                          { return &DevConfig{} }
func (washer *Washer) CheckDevConfig(arr []byte, configInfo, mac string, client DbDriver)([]byte){return []byte{}}

func (washer *Washer) GetDevData(devParamsKey string, devParamsKeysTokens DevMeta, worker DbRedisDriver) DevData {
	return DevData{}
}
func (washer *Washer) SetDevData(req *Request, worker DbRedisDriver) *ServerError {
	return nil
}
