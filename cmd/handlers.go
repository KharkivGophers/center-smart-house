package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"

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



//-----------------Common functions-------------------------------------------------------------------------------------------

func CheckError(desc string, err error) error {
	if err != nil {
		log.Errorln(desc, err)
		return err
	}
	return nil
}



func Float32ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 32)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(int64(n), 10)
}
