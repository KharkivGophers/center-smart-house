package http

import (
	"net/http"
	"github.com/gorilla/mux"
	"time"
	"github.com/gorilla/handlers"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"strings"

	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/drivers"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/sys"
	"fmt"
)

type HTTPServer struct {
	LocalServer Server
	DbServer    Server
	Controller  RoutinesController
}

func NewHTTPServer (local , db Server, controller RoutinesController) *HTTPServer{
	return &HTTPServer{
		LocalServer:local,
		DbServer: db,
		Controller: controller,
	}
}

func (server *HTTPServer) Run() {
	defer func() {
		if r := recover(); r != nil {
			server.Controller.Close()
			log.Error("HTTPServer Failed")
		}
	} ()

	r := mux.NewRouter()
	r.HandleFunc("/devices", server.getDevicesHandler).Methods(http.MethodGet)
	r.HandleFunc("/devices/{id}/data", server.getDevDataHandler).Methods(http.MethodGet)
	r.HandleFunc("/devices/{id}/config", server.getDevConfigHandler).Methods(http.MethodGet)

	r.HandleFunc("/devices/{id}/config", server.patchDevConfigHandler).Methods(http.MethodPatch)

	//provide static html pages
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../view/")))

	port := fmt.Sprint(server.LocalServer.Port)

	srv := &http.Server{
		Handler:      r,
		Addr:         server.LocalServer.IP + ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//CORS provides Cross-Origin Resource Sharing middleware
	http.ListenAndServe( server.LocalServer.IP + ":" + port, handlers.CORS()(r))

	go log.Fatal(srv.ListenAndServe())
}

//----------------------http Dynamic Connection----------------------------------------------------------------------------------

func (server *HTTPServer) getDevicesHandler(w http.ResponseWriter, r *http.Request) {

	dbClient:= GetDBConnection(server.DbServer)
	defer dbClient.Close()

	devices := dbClient.GetAllDevices()
	err := json.NewEncoder(w).Encode(devices)
	CheckError("getDevicesHandler JSON enc", err)
}

func (server *HTTPServer) getDevDataHandler(w http.ResponseWriter, r *http.Request) {

	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	vars := mux.Vars(r)
	devID := "device:" + vars["type"]+":"+vars["name"]+":"+vars["mac"]

	devMeta := DevMeta{vars["type"],vars["name"],vars["mac"],""}
	devParamsKey := devID + ":" + "params"

	device := IdentifyDevData(vars["type"])
	deviceData := device.GetDevData(devParamsKey, devMeta, dbClient.GetClient())

	err := json.NewEncoder(w).Encode(deviceData)
	CheckError("getDevDataHandler JSON enc", err)
}

func (server *HTTPServer) getDevConfigHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	devMeta := DevMeta{MAC:vars["mac"], Type:vars["type"], Name:vars["name"]}
	_, err :=ValidateDevMeta(devMeta)
	if err != nil {
		log.Error(err)
		return
	}

	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	configInfo := devMeta.MAC + ":" + "config" // key

	var device DevConfigDriver = *IdentifyDevConfig(devMeta.Type)
	log.Info(device)
	if device==nil{
		http.Error(w, "This type is not found", 400)
		return
	}
	config := device.GetDevConfig(configInfo, devMeta.MAC, dbClient.GetClient())
	json.NewEncoder(w).Encode(config)
}

func (server *HTTPServer) patchDevConfigHandler(w http.ResponseWriter, r *http.Request) {

	var config *DevConfig

	vars := mux.Vars(r)
	devMeta := DevMeta{MAC:vars["mac"], Type:vars["type"], Name:vars["name"]}
	_, err :=ValidateDevMeta(devMeta)
	if err != nil {
		log.Error(err)
		return
	}

	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	configInfo := devMeta.MAC + ":" + "config" // key

	var device DevConfigDriver = *IdentifyDevConfig(devMeta.Type)
	if device==nil{
		http.Error(w, "This type is not found", 400)
		return
	}
	config = device.GetDevConfig(configInfo, devMeta.MAC, dbClient.GetClient())

	// log.Warnln("Config Before", config)
	err = json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), 400)
		log.Errorln("NewDec: ", err)
	}

	valid, message := device.ValidateDevData(*config)
	if  !valid{
		http.Error(w, message, 400)
	} else {
		// Save New Configuration to DB
		device.SetDevConfig(configInfo,config, dbClient.GetClient())
		log.Println("New Config was added to DB: ", config)
		JSONСonfig, _ := json.Marshal(config)
		dbClient.Publish("configChan",JSONСonfig)
	}
}