package tcp

import (
	"encoding/json"
	"time"
	"net"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/models"
	"github.com/KharkivGophers/center-smart-house/dao"
	"fmt"
)

type TCPDataServer struct {
	DbServer	Server
	LocalServer	Server
	Reconnect	*time.Ticker
}

func NewTCPDataServer(local Server, db Server, reconnect *time.Ticker) *TCPDataServer {
	return &TCPDataServer{
		LocalServer: local,
		DbServer : db,
		Reconnect: reconnect,
	}
}

func (server *TCPDataServer) RunTCPServer() {

	ln, err := net.Listen("tcp", server.LocalServer.IP + ":" + fmt.Sprint(server.LocalServer.Port))

	for err != nil {

		for range server.Reconnect.C {
			ln, _ = net.Listen("tcp", server.LocalServer.IP + ":" + fmt.Sprint(server.LocalServer.Port))
		}
		server.Reconnect.Stop()
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
		go dao.PublishWS(req, "devWS", &dao.MyRedis{Host: server.DbServer.IP, Port: server.DbServer.Port})
		//go publishWS(req)

	default:
		log.Println("Device request: unknown action")
		return string("Device request: unknown action")

	}
	return string("Device request correct")
}
