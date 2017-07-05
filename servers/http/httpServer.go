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
	"github.com/KharkivGophers/center-smart-house/driver"
	. "github.com/KharkivGophers/center-smart-house/driver/devices"
	. "github.com/KharkivGophers/center-smart-house/dao"
	. "github.com/KharkivGophers/center-smart-house/sys"
	"strconv"
)

type HTTPServer struct{
	LocalServer Server
	DbServer Server
}

func NewHTTPServer (local , db Server) *HTTPServer{
	return &HTTPServer{LocalServer:local, DbServer: db}
}

func (server *HTTPServer)RunHTTPServer(control Control) {
	r := mux.NewRouter()
	r.HandleFunc("/devices", server.getDevicesHandler).Methods(http.MethodGet)
	r.HandleFunc("/devices/{id}/data", server.getDevDataHandler).Methods(http.MethodGet)
	r.HandleFunc("/devices/{id}/config", server.getDevConfigHandler).Methods(http.MethodGet)

	r.HandleFunc("/devices/{id}/config", server.patchDevConfigHandler).Methods(http.MethodPatch)

	//provide static html pages
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../view/")))

	var port string =strconv.FormatUint(uint64(server.LocalServer.Port),10)

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

func (server *HTTPServer)getDevicesHandler(w http.ResponseWriter, r *http.Request) {

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
	devID := "device:" + vars["id"]

	devParamsKeysTokens := []string{}
	devParamsKeysTokens = strings.Split(devID, ":")
	devParamsKey := devID + ":" + "params"

	device := dbClient.GetDevice(devParamsKey, devParamsKeysTokens)
	err := json.NewEncoder(w).Encode(device)
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

	dbClient, err := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	configInfo := devMeta.MAC + ":" + "config" // key

	var device driver.ConfigDevDriver = *IdentifyDev(devMeta.Type)
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

	var device driver.ConfigDevDriver = *IdentifyDev(devMeta.Type)
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

	log.Info(config)
	CheckError("Encode error", err) // log.Warnln("Config After: ", config)

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