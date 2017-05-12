package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"fmt"

)

func requestHandler(conn net.Conn) {
	var req Request
	var res Response

	err := json.NewDecoder(conn).Decode(&req)
	if err != nil {
		log.Errorln(err)
	}

	defer conn.Close()

	//sends resp struct from  devTypeHandler by channel;
	go devTypeHandler(req)

	res = Response{
		Status: http.StatusOK,
		Descr:  "Data have been delivered successfully",
	}

	err = json.NewEncoder(conn).Encode(&res)
	if err != nil {
		log.Errorln(err)
	}
}

func devTypeHandler(req Request) {
	switch req.Action {
	case "update":
		switch req.Meta.Type {
		case "fridge":
			if err := req.fridgeDataHandler(); err != nil {
				log.Errorf("%v", err.Error)
			}
		case "washer":
			if err := req.washerDataHandler(); err != nil {
				log.Errorf("%v", err.Error)
			}

		default:
			log.Println("Device request: unknown device type")
		}

	default:
		log.Println("Device request: unknown action")
	}
}

func (req *Request) fridgeDataHandler() *ServerError {
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	var key = devType + ":" + devName + ":" + mac

	dbClient.HMSet(key,
		"ReqTime", devReqTime,
	)

	var devData FridgeData
	json.Unmarshal([]byte(req.Data), &devData)

	for time, value := range devData.TempCam1 {
		dbClient.ZAdd(key+":"+"TempCam1",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
	}

	for time, value := range devData.TempCam2 {
		dbClient.ZAdd(key+":"+"TempCam2",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
	}

	return nil
}

func (req *Request) washerDataHandler() *ServerError {
	//to be continued
	return nil
}

func Float32ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 32)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(int64(n), 10)
}

//------------------------------------------DB
func GetKeyFromHTable(key, nameTable string) (string, error) {
	value, err := dbClient.HGet(nameTable, key)

	if err != nil {
		fmt.Println("Error when use hget")
		return "", err
	}
	return value, nil
}

func GetAllKeysFromHSet(nametable string) []string {
	keys, err := dbClient.HKeys(nametable)
	if err!=nil{
		return nil
	}
	fmt.Println(keys)
	return keys
}

//-------------------View handler
func staticHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(buildingRespone())
}



func buildingRespone() []ViewDevice{
	var nameGeneralTable string = "devices"
	var nameTableWithFunction string = "function:"
	var nameType struct{Name,Type string}
	collection :=[]ViewDevice{}

	macs := GetAllKeysFromHSet(nameGeneralTable)

	for _, value:=range macs{
		var device ViewDevice = ViewDevice{}
		data,_ :=GetKeyFromHTable(value, nameGeneralTable)
		function:=GetAllKeysFromHSet(nameTableWithFunction+value)

		json.Unmarshal([]byte(data), &nameType)

		device.Name = nameType.Name
		device.Type = nameType.Type
		device.Location ="Home"
		device.Functions = function

		collection = append(collection,device)

	}
	return collection
}
