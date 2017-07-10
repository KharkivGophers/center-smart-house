package tcp

import (
	"encoding/json"
	"time"
	"net"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/drivers"
	"fmt"
	. "github.com/KharkivGophers/center-smart-house/sys"
)

type TCPDevDataServer struct {
	LocalServer Server
	Reconnect   *time.Ticker
	Controller  RoutinesController
	DbClient  DbClient
}

func NewTCPDevDataServer(local Server, reconnect *time.Ticker, controller  RoutinesController,dbClient DbClient) *TCPDevDataServer {
	return &TCPDevDataServer{
		LocalServer: local,
		Reconnect:   reconnect,
		Controller:  controller,
		DbClient: dbClient,
	}
}

func NewTCPDevDataServerDefault(local Server, controller  RoutinesController,dbClient DbClient) *TCPDevDataServer {
	reconnect := time.NewTicker(time.Second * 1)
	return NewTCPDevDataServer(local, reconnect, controller,dbClient)
}

func (server *TCPDevDataServer) Run() {
	defer func() {
		if r := recover(); r != nil {
			server.Controller.Close()
			log.Error("TCPDevDataServer Failed")
		}
	} ()

	ln, err := net.Listen("tcp", server.LocalServer.IP+":"+fmt.Sprint(server.LocalServer.Port))

	for err != nil {

		for range server.Reconnect.C {
			ln, _ = net.Listen("tcp", server.LocalServer.IP+":"+fmt.Sprint(server.LocalServer.Port))
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
func (server *TCPDevDataServer) tcpDataHandler(conn net.Conn) {
	var req Request
	var res Response
	for {
		err := json.NewDecoder(conn).Decode(&req)
		if err != nil {
			log.Errorln("tcpDataHandler. requestHandler JSON Decod", err)
			return
		}
		//sends resp struct from  devTypeHandler by channel;
		go server.devTypeHandler(req)

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
func (server TCPDevDataServer) devTypeHandler(req Request) string {
	dbClient := server.DbClient.NewDBConnection()
	defer dbClient.Close()

	switch req.Action {
	case "update":
		data := IdentifyDevice(req.Meta.Type)
		if data == nil{
			return string("Device request: unknown device type")
		}
		log.Println("Data has been received")

		data.SetDevData(&req, dbClient)
		go PublishWS(req, "devWS", server.DbClient)
		//go publishWS(req)

	default:
		log.Println("Device request: unknown action")
		return string("Device request: unknown action")

	}
	return string("Device request correct")
}


