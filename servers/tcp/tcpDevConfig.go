package tcp

import (
	//"strings"
	"encoding/json"
	"time"
	"net"
	"fmt"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/sys"
	. "github.com/KharkivGophers/center-smart-house/drivers"

)

type TCPDevConfigServer struct {
	DbServer            Server
	LocalServer         Server
	Reconnect           *time.Ticker
	Pool                ConnectionPool
	Messages            chan []string
	StopConfigSubscribe chan struct{}
	Controller          RoutinesController
}

func NewTCPDevConfigServer(local Server, db Server, reconnect *time.Ticker, messages chan []string, stopConfigSubscribe chan struct{},
	controller RoutinesController) *TCPDevConfigServer {
	return &TCPDevConfigServer{
		LocalServer:         local,
		DbServer:            db,
		Reconnect:           reconnect,
		Messages:            messages,
		Controller:          controller,
		StopConfigSubscribe: stopConfigSubscribe,
	}
}
func NewTCPDevConfigServerDefault(local Server, db Server, controller RoutinesController) *TCPDevConfigServer {
	messages := make(chan []string)
	stopConfigSubscribe := make(chan struct{})
	reconnect := time.NewTicker(time.Second * 1)
	return NewTCPDevConfigServer(local, db, reconnect, messages, stopConfigSubscribe, controller)
}

func (server *TCPDevConfigServer) Run() {
	defer func() {
		if r := recover(); r != nil {
			server.StopConfigSubscribe <- struct{}{}
			server.Controller.Close()
			log.Error("TCPDevConfigServer Failed")
		}
	}()

	server.Pool.Init()

	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	ln, err := net.Listen("tcp", server.LocalServer.IP+":"+fmt.Sprint(server.LocalServer.Port))

	for err != nil {
		for range server.Reconnect.C {
			ln, _ = net.Listen("tcp", server.LocalServer.IP+":"+fmt.Sprint(server.LocalServer.Port))
		}
		server.Reconnect.Stop()
	}

	go server.configSubscribe("configChan", server.Messages, &server.Pool)

	for {
		conn, err := ln.Accept()
		CheckError("TCP config conn Accept", err)
		go server.sendDefaultConfiguration(conn, &server.Pool)
	}
}

func (server *TCPDevConfigServer) sendNewConfiguration(config DevConfig, pool *ConnectionPool) {
	connection := pool.GetConn(config.MAC)
	if connection == nil {
		log.Error("Has not connection with mac:config.MAC  in connectionPool")
		return
	}

	_, err := connection.Write(config.Data)

	if err != nil {
		pool.RemoveConn(config.MAC)
	}
	CheckError("sendNewConfig", err)
}

func (server *TCPDevConfigServer) sendDefaultConfiguration(conn net.Conn, pool *ConnectionPool) {
	var (
		req    Request
		config *DevConfig
		device DevConfigDriver
	)
	err := json.NewDecoder(conn).Decode(&req)
	CheckError("sendDefaultConfiguration JSON Decod", err)

	device = *IdentifyDevConfig(req.Meta.Type)//device struct
	configInfo := req.Meta.MAC + ":" + "config" // key

	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	pool.AddConn(conn, req.Meta.MAC)

	if ok, _ := dbClient.GetClient().Exists(configInfo); ok {
		config = device.GetDevConfig(configInfo, req.Meta.MAC, dbClient.GetClient())
		log.Println("Old Device with MAC: ", req.Meta.MAC, "detected.")

	} else {

		log.Warningln("New Device with MAC: ", req.Meta.MAC, "detected.")
		log.Warningln("Default Config will be sent.")
		config = device.GetDefaultConfig()

		device.SetDevConfig(configInfo, config, dbClient.GetClient())
	}

	//err = json.NewEncoder(conn).Encode(&config)
	_, err = conn.Write(config.Data)
	CheckError("sendDefaultConfiguration JSON enc", err)

	log.Warningln("Configuration has been successfully sent ", err)

}

func (server *TCPDevConfigServer) configSubscribe(roomID string, message chan []string, pool *ConnectionPool) {
	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	dbClient.Subscribe(message, roomID)
	for {
		var config DevConfig
		select {
		case msg := <-message:
			if msg[0] == "message" {
				err := json.Unmarshal([]byte(msg[2]), &config)
				CheckError("configSubscribe: unmarshal", err)
				go server.sendNewConfiguration(config, pool)
			}
		case <-server.StopConfigSubscribe:
			return
		}
	}
}
