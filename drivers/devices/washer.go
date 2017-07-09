package devices

import (
	"time"
	"net/http"
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/sys"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/models"
)

/**
Time keep
 */
type Washer struct {
	Data   WasherData
	Config DevConfig
	Meta   DevMeta
}

type WasherData struct {
	Turnovers map[int64]int64        `json:"turnovers"`
	WaterTemp map[int64]float32      `json:"waterTemp"`
}

type WasherConfig struct {
	MAC            string   `json:"mac"`
	Temperature    float32  `json:"temperature"`
	WashTime       int64    `json:"washTime"`
	WashTurnovers  int64    `json:"washTurnovers"`
	RinseTime      int64    `json:"rinseTime"`
	RinseTurnovers int64    `json:"rinseTurnovers"`
	SpinTime       int64    `json:"spinTime"`
	SpinTurnovers  int64    `json:"spinTurnovers"`
}

var (
	LightMode WasherConfig = WasherConfig{
		Temperature:    60,
		WashTime:       90,
		WashTurnovers:  240,
		RinseTime:      30,
		RinseTurnovers: 120,
		SpinTime:       30,
		SpinTurnovers:  60,
	}
	FastMode WasherConfig = WasherConfig{
		Temperature:    180,
		WashTime:       30,
		WashTurnovers:  300,
		RinseTime:      15,
		RinseTurnovers: 240,
		SpinTime:       15,
		SpinTurnovers:  60,
	}
	StandartMode WasherConfig = WasherConfig{
		Temperature:    240,
		WashTime:       120,
		WashTurnovers:  240,
		RinseTime:      60,
		RinseTurnovers: 180,
		SpinTime:       60,
		SpinTurnovers:  60,
	}
)

func (washer *Washer) selectMode(mode string) (WasherConfig) {
	switch mode {
	case "LightMode":
		return LightMode
	case "FastMode":
		return FastMode
	case "StandartMode":
		return StandartMode
	}
	return WasherConfig{}
}

func (washer *Washer) GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig) {
	config := DevConfig{}

	t := time.Now().UnixNano() / int64(time.Minute)

	mode, err := worker.ZRangeByScore(configInfo, t-1, t+1)
	if err != nil {
		CheckError("Washer. GetDevConfig. Cant perform ZRangeByScore", err)
	}
	log.Info("Washer. GetDevConfig. Mode ", mode, "Time ", t)
	if len(mode) == 0 {
		return &config
	}
	configWasher := washer.selectMode(mode[0])
	config.Data, err = json.Marshal(configWasher)
	config.MAC = mac

	CheckError("Washer. GetDevConfig. Cant perform json.Marshal(configWasher)", err)

	return &config
}
func (washer *Washer) SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver) {

}
func (washer *Washer) ValidateDevData(config DevConfig) (bool, string) { return true, "" }
func (washer *Washer) GetDefaultConfig() (*DevConfig) {
	return &DevConfig{}
}
func (washer *Washer) CheckDevConfigAndMarshal(arr []byte, configInfo, mac string, client DbDriver) ([]byte) { return []byte{} }

func (washer *Washer) GetDevData(devParamsKey string, devParamsKeysTokens DevMeta, worker DbRedisDriver) DevData {
	return DevData{}
}
func (washer *Washer) SetDevData(req *Request, worker DbRedisDriver) *ServerError {

	var devData WasherData

	devKey := "device" + ":" + req.Meta.Type + ":" + req.Meta.Name + ":" + req.Meta.MAC
	devParamsKey := devKey + ":" + "params"

	_, err := worker.SAdd("devParamsKeys", devParamsKey)
	CheckError("SetDevData. Exception with SAdd", err)
	_, err = worker.HMSet(devKey, "ReqTime", req.Time)
	CheckError("SetDevData. Exception with HMSet", err)
	_, err = worker.SAdd(devParamsKey, "Turnovers", "WaterTemp")
	CheckError("SetDevData. Exception with SAdd", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		log.Errorf("Error in SetDevData. %v", err)
		return &ServerError{Error: err}
	}

	err = setDevDataInt64(devData.Turnovers, devParamsKey+":"+"Turnovers", worker)
	if err != nil {
		return &ServerError{Error: err}
	}

	err = setDevDataFloat32(devData.WaterTemp, devParamsKey+":"+"WaterTemp", worker)
	if err != nil {
		return &ServerError{Error: err}
	}

	return nil
}

func setDevDataFloat32(TempCam map[int64]float32, key string, worker DbRedisDriver) error {
	for time, value := range TempCam {
		_, err := worker.ZAdd(key, Int64ToString(time), Int64ToString(time)+":"+Float32ToString(value))
		if CheckError("setDevDataFloat32. Exception with ZAdd", err) != nil {
			return err
		}
	}
	return nil
}
func setDevDataInt64(TempCam map[int64]int64, key string, worker DbRedisDriver) error {
	for time, value := range TempCam {
		_, err := worker.ZAdd(key, Int64ToString(time), Int64ToString(time)+":"+Int64ToString(value))
		if CheckError("setDevDataInt64. Exception with ZAdd", err) != nil {
			return err
		}
	}
	return nil
}

func (washer *Washer) GetDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta) {

}
