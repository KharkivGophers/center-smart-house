package tcpData

import (
	"encoding/json"
	"time"
	"net"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/common"
	. "github.com/KharkivGophers/center-smart-house/common/models"
	"github.com/KharkivGophers/center-smart-house/dao"
)

type TCPDataServer struct {
	DBURL

	Host      string
	Port      string
	reconnect *time.Ticker
}

func NewTCPDataServer(host, port string, dburl DBURL, reconnect *time.Ticker) *TCPDataServer {
	return &TCPDataServer{
		Host: host,
		Port: port,
		DBURL: dburl,
		reconnect: reconnect,
	}
}

func (server *TCPDataServer) RunTCPServer() {

	ln, err := net.Listen("tcp", server.Host+":"+server.Port)

	for err != nil {

		for range server.reconnect.C {
			ln, _ = net.Listen("tcp", server.Host+":"+server.Port)
		}
		server.reconnect.Stop()
	}
	for {
		conn, err := ln.Accept()
		if CheckError("TCP conn Accept", err) == nil {
			go server.tcpDataHandler(conn)
		}
	}
}

//--------------------TCP-------------------------------------------------------------------------------------
func (server *TCPDataServer) tcpDataHandler(conn net.Conn) {
	var req Request
	var res Response
	for {
		err := json.NewDecoder(conn).Decode(&req)
		if err != nil {
			log.Errorln("requestHandler JSON Decod", err)
			return
		}
		//sends resp struct from  devTypeHandler by channel;
		go server.devTypeHandler(req)

		log.Println("Data has been received")

		res = Response{
			Status: 200,
			Descr:  "Data has been delivered successfully",
		}
		err = json.NewEncoder(conn).Encode(&res)
		CheckError("tcpDataHandler JSON enc", err)
	}

}

/*
Checks  type device and call special func for send data to DB.
*/
func (server *TCPDataServer) devTypeHandler(req Request) string {
	switch req.Action {
	case "update":
		switch req.Meta.Type {
		case "fridge":
			if err := server.fridgeDataHandler(&req); err != nil {
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
		go dao.PublishWS(req, "devWS", &dao.MyRedis{Host: "0.0.0.0", Port: 6379})
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
func (server *TCPDataServer) fridgeDataHandler(req *Request) *ServerError {
	myRedis, err := dao.MyRedis{Host: server.DbHost, Port: server.DbPort}.RunDBConnection()
	defer myRedis.Close()

	var devData FridgeData
	mac := req.Meta.MAC
	devReqTime := req.Time
	devType := req.Meta.Type
	devName := req.Meta.Name

	devKey := "device" + ":" + devType + ":" + devName + ":" + mac
	devParamsKey := devKey + ":" + "params"

	_, err = myRedis.Client.SAdd("devParamsKeys", devParamsKey)
	CheckError("DB error11", err)
	_, err = myRedis.Client.HMSet(devKey, "ReqTime", devReqTime)
	CheckError("DB error12", err)
	_, err = myRedis.Client.SAdd(devParamsKey, "TempCam1", "TempCam2")
	CheckError("DB error13", err)

	err = json.Unmarshal([]byte(req.Data), &devData)
	if err != nil {
		log.Error("Error in fridgedatahandler")
		return &ServerError{Error: err}
	}

	for time, value := range devData.TempCam1 {
		_, err := myRedis.Client.ZAdd(devParamsKey+":"+"TempCam1",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error14", err) != nil {
			return &ServerError{Error: err}
		}
	}

	for time, value := range devData.TempCam2 {
		_, err := myRedis.Client.ZAdd(devParamsKey+":"+"TempCam2",
			Int64ToString(time), Int64ToString(time)+":"+Float32ToString(float64(value)))
		if CheckError("DB error15", err) != nil {
			return &ServerError{Error: err}
		}
	}

	return nil
}
