package tcp

import (
	"strings"
	"encoding/json"
	"time"
	"net"

	log "github.com/Sirupsen/logrus"

	. "github.com/KharkivGophers/center-smart-house/models"
	"github.com/KharkivGophers/center-smart-house/dao"
	"fmt"
	"github.com/KharkivGophers/center-smart-house/drivers"
)

type TCPConfigServer struct {
	DbServer    Server
	LocalServer Server
	Reconnect   *time.Ticker
	Pool        ConnectionPool
	Messages    chan []string
}

func NewTCPConfigServer(local Server, db Server, reconnect *time.Ticker, messages chan []string) *TCPConfigServer {
	return &TCPConfigServer{
		LocalServer: local,
		DbServer:    db,
		Reconnect:   reconnect,
		Messages:    messages,
	}
}

func NewDefaultConfig() *DevConfig {
	return &DevConfig{
		TurnedOn:    true,
		StreamOn:    true,
		CollectFreq: 1000,
		SendFreq:    5000,
	}
}

func (server *TCPConfigServer) RunConfigServer() {

	//TODO For this method could work correct we must add switch case. Which chose type our device. For exemle in TCP Data connection
	//

	server.Pool.Init()

	myRedis, err := dao.MyRedis{Host: server.DbServer.IP, Port: server.DbServer.Port}.RunDBConnection()
	defer myRedis.Close()
	CheckError("TCP Connection: runConfigServer", err)

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

func (server *TCPConfigServer) sendNewConfiguration(config DevConfig, pool *ConnectionPool) {


	connection := pool.GetConn(config.MAC)
	if connection == nil {
		log.Error("Has not connection with mac:config.MAC  in connectionPool")
		return
	}

	// log.Println("mac in pool sendNewConfig", config.MAC)
	err := json.NewEncoder(connection).Encode(&config)

	if err != nil {
		pool.RemoveConn(config.MAC)
	}
	CheckError("sendNewConfig", err)
}

func (server *TCPConfigServer) sendDefaultConfiguration(conn net.Conn, pool *ConnectionPool) {
	var (
		req    Request
		config *DevConfig
		device drivers.ConfigDevDriver

	)
	err := json.NewDecoder(conn).Decode(&req)
	CheckError("sendDefaultConfiguration JSON Decod", err)

	device = *drivers.SelectDevice(req)

	// Send Default Configuration to Device

	myRedis, err := dao.MyRedis{Host: server.DbServer.IP, Port: server.DbServer.Port}.RunDBConnection()
	defer myRedis.Close()
	CheckError("DBConnection Error in ----> sendDefaultConfiguration", err)


	pool.AddConn(conn, req.Meta.MAC)

	configInfo := req.Meta.MAC + ":" + "config" // key

	if ok, _ := myRedis.Client.Exists(configInfo); ok {

		state, err := myRedis.Client.HMGet(configInfo, "TurnedOn")
		CheckError("Get from DB error: TurnedOn ", err)

		if strings.Join(state, " ") != "" {
			config = device.GetDevConfig(configInfo, req.Meta.MAC, myRedis.Client)
			log.Println("Old Device with MAC: ", req.Meta.MAC, "detected.")
		}

	} else {
		log.Warningln("New Device with MAC: ", req.Meta.MAC, "detected.")
		log.Warningln("Default Config will be sent.")
		config = NewDefaultConfig()
		device.SetDevConfig(configInfo, config,myRedis.Client)
	}

	err = json.NewEncoder(conn).Encode(&config)
	CheckError("sendDefaultConfiguration JSON enc", err)
	log.Warningln("Configuration has been successfully sent")
}

func (server *TCPConfigServer) configSubscribe(roomID string, message chan []string, pool *ConnectionPool) {
	myRedis, err := dao.MyRedis{Host: server.DbServer.IP, Port: server.DbServer.Port}.RunDBConnection()
	defer myRedis.Close()
	CheckError("TCP Connection: runConfigServer", err)

	myRedis.Subscribe(message, roomID)
	for {
		var config DevConfig
		select {
		case msg := <-message:
			if msg[0] == "message" {
				err := json.Unmarshal([]byte(msg[2]), &config)
				CheckError("configSubscribe: unmarshal", err)
				go server.sendNewConfiguration(config, pool)
			}
		}
	}
}
