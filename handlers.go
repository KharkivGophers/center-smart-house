package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"menteslibres.net/gosexy/redis"
	"github.com/KharkivGophers/center-smart-house/common/models"
	"github.com/KharkivGophers/center-smart-house/dao"
)

//--------------------TCP-------------------------------------------------------------------------------------
func tcpDataHandler(conn net.Conn) {
	var req models.Request
	var res Response
	for {
		err := json.NewDecoder(conn).Decode(&req)
		if err != nil {
			log.Errorln("requestHandler JSON Decod", err)
			return
		}
		//sends resp struct from  devTypeHandler by channel;
		go devTypeHandler(req)

		log.Println("Data has been received")

		res = Response{
			Status: http.StatusOK,
			Descr:  "Data has been delivered successfully",
		}
		err = json.NewEncoder(conn).Encode(&res)
		CheckError("tcpDataHandler JSON enc", err)
	}

}

/*
Checks  type device and call special func for send data to DB.
*/
func devTypeHandler(req models.Request) string {
	switch req.Action {
	case "update":
		switch req.Meta.Type {
		case "fridge":
			if err := fridgeDataHandler(&req); err != nil {
				log.Errorf("%v", err.Error)
			}
		case "washer":
			//if err := req.washerDataHandler(); err != nil {
			//	log.Errorf("%v", err.Error)
			//}

		default:
			log.Println("Device request: unknown device type")
			return string("Device request: unknown device type")
		}
		go dao.PublishWS(req, "devWS",&dao.MyRedis{Host:"0.0.0.0", Port:6379} )
		//go publishWS(req)

	default:
		log.Println("Device request: unknown action")
		return string("Device request: unknown action")

	}
	return string("Device request correct")
}

/*
Save data about fridge in DB. Return struct ServerError
*/
func  fridgeDataHandler(req *models.Request) *ServerError {
	dbClient, _ := RunDBConnection()

	var devData FridgeData
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	devKey := "device" + ":" + devType + ":" + devName + ":" + mac
	devParamsKey := devKey + ":" + "params"

	_, err := dbClient.SAdd("devParamsKeys", devParamsKey)
	CheckError("DB error11", err)
	_, err = dbClient.HMSet(devKey, "ReqTime", devReqTime)
	CheckError("DB error12", err)
	_, err = dbClient.SAdd(devParamsKey, "TempCam1", "TempCam2")
	CheckError("DB error13", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		log.Error("Error in fridgedatahandler")
		return &ServerError{Error: err}
	}

	for time, value := range devData.TempCam1 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam1",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error14", err) != nil {
			return &ServerError{Error: err}
		}
	}

	for time, value := range devData.TempCam2 {
		_, err := dbClient.ZAdd(devParamsKey+":"+"TempCam2",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error15", err) != nil {
			return &ServerError{Error: err}
		}
	}

	return nil
}

func (req *Request) washerDataHandler() *ServerError {
	//to be continued
	return nil
}

func configSubscribe(client *redis.Client, roomID string, message chan []string, pool *ConnectionPool) {
	Subscribe(client, roomID, message)
	for {
		var config DevConfig
		select {
		case msg := <-message:
			// log.Println("message", msg)
			if msg[0] == "message" {
				// log.Println("message[0]", msg[0])
				err := json.Unmarshal([]byte(msg[2]), &config)
				CheckError("configSubscribe: unmarshal", err)
				go sendNewConfiguration(config, pool)
			}
		}
	}
}

//-----------------Common functions-------------------------------------------------------------------------------------------

func CheckError(desc string, err error) error {
	if err != nil {
		log.Errorln(desc, err)
		return err
	}
	return nil
}

func Subscribe(client *redis.Client, roomID string, channel chan []string) {
	err := client.Connect(dbHost, dbPort)
	CheckError("Subscribe", err)
	go client.Subscribe(channel, roomID)
}

func Float32ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 32)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(int64(n), 10)
}
//-------------------Work with data base-------------------------------------------------------------------------------------------

func SetFridgeConfig(dbClient *redis.Client, configInfo string, config *DevConfig) {
	// Save default configuration to DB
	_, err := dbClient.HMSet(configInfo, "TurnedOn", config.TurnedOn)
	CheckError("DB error1: TurnedOn", err)
	_, err = dbClient.HMSet(configInfo, "CollectFreq", config.CollectFreq)
	CheckError("DB error2: CollectFreq", err)
	_, err = dbClient.HMSet(configInfo, "SendFreq", config.SendFreq)
	CheckError("DB error3: SendFreq", err)
	_, err = dbClient.HMSet(configInfo, "StreamOn", config.StreamOn)
	CheckError("DB error4: StreamOn", err)
}

func GetFridgeConfig(dbClient *redis.Client, configInfo, mac string) (*DevConfig) {
	var config DevConfig

	state, err := dbClient.HMGet(configInfo, "TurnedOn")
	CheckError("Get from DB error1: TurnedOn ", err)
	sendFreq, err := dbClient.HMGet(configInfo, "SendFreq")
	CheckError("Get from DB error2: SendFreq ", err)
	collectFreq, err := dbClient.HMGet(configInfo, "CollectFreq")
	CheckError("Get from DB error3: CollectFreq ", err)
	streamOn, err := dbClient.HMGet(configInfo, "StreamOn")
	CheckError("Get from DB error4: StreamOn ", err)

	stateBool, _ := strconv.ParseBool(strings.Join(state, " "))
	sendFreqInt, _ := strconv.Atoi(strings.Join(sendFreq, " "))
	collectFreqInt, _ := strconv.Atoi(strings.Join(collectFreq, " "))
	streamOnBool, _ := strconv.ParseBool(strings.Join(streamOn, " "))

	config = DevConfig{
		MAC: mac,
		TurnedOn:    stateBool,
		CollectFreq: int64(collectFreqInt),
		SendFreq:    int64(sendFreqInt),
		StreamOn:    streamOnBool,
	}

	log.Println("Configuration from DB: ", state, sendFreq, collectFreq)

	return &config
}
