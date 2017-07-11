package tcp

import (
	"encoding/json"
	"time"
	"net"
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/sys"
	. "github.com/KharkivGophers/center-smart-house/drivers"


)

type TCPDevConfigServer struct {
	LocalServer         Server
	Reconnect           *time.Ticker
	Pool                ConnectionPool
	Messages            chan []string
	StopConfigSubscribe chan struct{}
	Controller          RoutinesController
	DbClient 	    DbClient
}

func NewTCPDevConfigServer(local Server, reconnect *time.Ticker, messages chan []string, stopConfigSubscribe chan struct{},
	controller RoutinesController, dbClient DbClient) *TCPDevConfigServer {
	return &TCPDevConfigServer{
		LocalServer:         local,
		Reconnect:           reconnect,
		Messages:            messages,
		Controller:          controller,
		StopConfigSubscribe: stopConfigSubscribe,
		DbClient: dbClient,
	}
}
func NewTCPDevConfigServerDefault(local Server,  controller RoutinesController, dbClient DbClient) *TCPDevConfigServer {
	messages := make(chan []string)
	stopConfigSubscribe := make(chan struct{})
	reconnect := time.NewTicker(time.Second * 1)
	return NewTCPDevConfigServer(local,  reconnect, messages, stopConfigSubscribe, controller,dbClient)
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

	dbClient := server.DbClient.NewDBConnection()
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
		config DevConfig
		device DevServerHandler
	)
	err := json.NewDecoder(conn).Decode(&req)
	CheckError("sendDefaultConfiguration JSON Decod", err)
	pool.AddConn(conn, req.Meta.MAC)
	dbClient := server.DbClient.NewDBConnection()
	defer dbClient.Close()

	device = IdentifyDevice(req.Meta.Type)//device struct
	log.Info(reflect.TypeOf(device))
	config.Data = device.SendDefaultConfigurationTCP(conn, dbClient, &req)

	_, err = conn.Write(config.Data)
	CheckError("sendDefaultConfiguration JSON enc", err)

	log.Warningln("Configuration has been successfully sent ", err)

}

func (server *TCPDevConfigServer) configSubscribe(roomID string, message chan []string, pool *ConnectionPool) {
	dbClient := server.DbClient.NewDBConnection()
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
