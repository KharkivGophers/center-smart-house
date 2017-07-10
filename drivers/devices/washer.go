package devices

import (
	"time"
	"net/http"
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/sys"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/models"
	"net"
)

/**
Time keep
 */
type Washer struct {
	Data   WasherData
	Config WasherConfig
	Meta   DevMeta

	timeStartWash int64
}

type WasherData struct {
	Turnovers map[int64]int64        `json:"turnovers"`
	WaterTemp map[int64]float32      `json:"waterTemp"`
}

type WasherConfig struct {
	Name           string   `json:"name"`
	MAC            string   `json:"mac"`
	Temperature    float32  `json:"temperature"`
	WashTime       int64    `json:"washTime"`
	WashTurnovers  int64    `json:"washTurnovers"`
	RinseTime      int64    `json:"rinseTime"`
	RinseTurnovers int64    `json:"rinseTurnovers"`
	SpinTime       int64    `json:"spinTime"`
	SpinTurnovers  int64    `json:"spinTurnovers"`
}
type TimerMode struct {
	Name      string   `json:"name"`
	StartTime int64    `json:time`
}

var (
	LightMode WasherConfig = WasherConfig{
		Name:           "LightMode",
		Temperature:    60,
		WashTime:       90,
		WashTurnovers:  240,
		RinseTime:      30,
		RinseTurnovers: 120,
		SpinTime:       30,
		SpinTurnovers:  60,
	}
	FastMode WasherConfig = WasherConfig{
		Name:           "FastMode",
		Temperature:    180,
		WashTime:       30,
		WashTurnovers:  300,
		RinseTime:      15,
		RinseTurnovers: 240,
		SpinTime:       15,
		SpinTurnovers:  60,
	}
	StandartMode WasherConfig = WasherConfig{
		Name:           "StandartMode",
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
		washer.Config = LightMode
		return LightMode
	case "FastMode":
		washer.Config = FastMode
		return FastMode
	case "StandartMode":
		washer.Config = StandartMode
		return StandartMode
	}
	return WasherConfig{}
}


//--------------------------------------DevDataDriver--------------------------------------------------------------
func (washer *Washer) GetDevData(devParamsKey string, devMeta DevMeta, worker DbRedisDriver) DevData {
	var device DevData

	params, err := worker.SMembers(devParamsKey)
	CheckError("Cant read members from devParamsKeys", err)
	device.Meta = devMeta
	device.Data = make(map[string][]string)

	values := make([][]string, len(params))
	for i, p := range params {
		values[i], err = worker.ZRangeByScore(devParamsKey+":"+p, "-inf", "inf")
		CheckError("Cant use ZRangeByScore", err)
		device.Data[p] = values[i]
	}
	return device
}

func (washer *Washer) SetDevData(req *Request, worker DbRedisDriver) *ServerError {

	var devData WasherData

	devKey := "device" + ":" + req.Meta.Type + ":" + req.Meta.Name + ":" + req.Meta.MAC
	devParamsKey := devKey + ":" + "params"

	err := json.Unmarshal([]byte(req.Data), &devData)
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
		_, err := worker.ZAdd(key, Int64ToString(time), Int64ToString(time)+":"+Float64ToString(value))
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


//--------------------------------------DevConfigDriver--------------------------------------------------------------
func (washer *Washer) GetDevConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig) {
	return &DevConfig{}

}

//This method needed for work with http. If washer is have washing we will
// return washer.Config, else we will return empty WasherConfig
//func (washer *Washer) getDevConfig()(bool, WasherConfig) {
//	if washer.timeStartWash <= 0 {
//		return  false, WasherConfig{}
//	}
//
//	t := time.Now().UnixNano() / int64(time.Minute)
//	timeAfterStart := t - washer.timeStartWash
//	timeWasherWorking := washer.Config.SpinTime + washer.Config.RinseTime + washer.Config.WashTime
//
//	if timeWasherWorking - timeAfterStart < 0 {
//		washer.timeStartWash = 0
//		washer.Config = WasherConfig{}
//		return false, washer.Config
//	}
//
//	return true, washer.Config
//}

func (washer *Washer) SetDevConfig(configInfo string, config *DevConfig, worker DbRedisDriver) {
	var timerMode TimerMode
	json.Unmarshal(config.Data, &timerMode)
	_, err := worker.ZAdd(configInfo, timerMode.StartTime, timerMode.Name)
	CheckError("DB error1: TurnedOn", err)
}

func (washer *Washer) setDevToBD(configInfo string, config *DevConfig, worker DbRedisDriver, req *Request) {

	var timerMode TimerMode
	json.Unmarshal(config.Data, &timerMode)

	devKey := "device" + ":" + req.Meta.Type + ":" + req.Meta.Name + ":" + req.Meta.MAC
	devParamsKey := devKey + ":" + "params"

	_, err := worker.SAdd("devParamsKeys", devParamsKey)
	CheckError("SetDevData. Exception with SAdd", err)
	_, err = worker.HMSet(devKey, "ReqTime", req.Time)
	CheckError("SetDevData. Exception with HMSet", err)
	_, err = worker.SAdd(devParamsKey, "Turnovers", "WaterTemp")
	CheckError("SetDevData. Exception with SAdd", err)

	_, err = worker.ZAdd(configInfo, timerMode.StartTime, timerMode.Name)
	CheckError("DB error1: TurnedOn", err)
}

func (washer *Washer) ValidateDevData(config DevConfig) (bool, string) {
	return true, ""
}
func (washer *Washer) GetDefaultConfig() (*DevConfig) {
	b, _ := json.Marshal(WasherConfig{})
	return &DevConfig{Data: b}
}
func (washer *Washer) CheckDevConfigAndMarshal(arr []byte, configInfo, mac string, client DbClient) ([]byte) {
	return []byte{}
}



//--------------------------------------DevServerHandler--------------------------------------------------------------

func (washer *Washer) GetDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta, client DbClient) {
}

func (washer *Washer) SendDefaultConfigurationTCP(conn net.Conn, dbClient DbClient, req *Request) ([]byte) {
	var config *DevConfig
	configInfo := req.Meta.MAC + ":" + "config" // key

	if ok, _ := dbClient.GetClient().Exists(configInfo); ok {
		config = washer.getActualConfig(configInfo, req.Meta.MAC, dbClient.GetClient())
		log.Println("Old Device with MAC: ", req.Meta.MAC, "detected.")

	} else {
		log.Warningln("New Device with MAC: ", req.Meta.MAC, "detected.")
		log.Warningln("Default Config will be sent.")
		config = washer.GetDefaultConfig()
		washer.setDevToBD(configInfo, config, dbClient.GetClient(), req)
	}
	return config.Data
}

func (washer *Washer) PatchDevConfigHandlerHTTP(w http.ResponseWriter, r *http.Request, meta DevMeta, client DbClient) {

}

func (washer *Washer) getActualConfig(configInfo, mac string, worker DbRedisDriver) (*DevConfig) {
	config := washer.GetDefaultConfig()
	config.MAC = mac

	t := time.Now().UnixNano() / int64(time.Minute)

	mode, err := worker.ZRangeByScore(configInfo, t-100, t+100)
	if err != nil {
		CheckError("Washer. GetDevConfig. Cant perform ZRangeByScore", err)
	}
	log.Info("Washer. GetDevConfig. Mode ", mode, "Time ", t)

	if len(mode) == 0 {
		return config
	}

	configWasher := washer.selectMode(mode[0])
	config.Data, err = json.Marshal(configWasher)

	CheckError("Washer. GetDevConfig. Cant perform json.Marshal(configWasher)", err)
	return config
}
