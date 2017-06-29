package myHTTP

import (
	"net/http"
	"github.com/gorilla/mux"
	"time"
	"github.com/gorilla/handlers"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"strings"

	. "github.com/KharkivGophers/center-smart-house/common/models"
	. "github.com/KharkivGophers/center-smart-house/common"
	"github.com/KharkivGophers/center-smart-house/dao"
)

type HTTPServer struct{
	Host     string
	Port     string

	DBURL	 DBURL
}

func NewHTTPServer (host, port string, dburl DBURL) *HTTPServer{
	return &HTTPServer{Host:host, Port:port, DBURL: dburl}
}

func (server *HTTPServer)RunDynamicServer() {
	r := mux.NewRouter()
	r.HandleFunc("/devices", server.getDevicesHandler).Methods("GET")
	r.HandleFunc("/devices/{id}/data", server.getDevDataHandler).Methods("GET")
	r.HandleFunc("/devices/{id}/config", server.getDevConfigHandler).Methods("GET")

	r.HandleFunc("/devices/{id}/config", server.patchDevConfigHandler).Methods("PATCH")

	//provide static html pages
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../view/")))

	srv := &http.Server{
		Handler:      r,
		Addr:         server.Host + ":" + server.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//CORS provides Cross-Origin Resource Sharing middleware
	http.ListenAndServe(server.Host + ":" + server.Port, handlers.CORS()(r))

	go log.Fatal(srv.ListenAndServe())
}

//----------------------myHTTP Dynamic Connection----------------------------------------------------------------------------------

func (server *HTTPServer)getDevicesHandler(w http.ResponseWriter, r *http.Request) {

	myRedis, err := dao.MyRedis{Host: server.DBURL.DbHost, Port: server.DBURL.DbPort}.RunDBConnection()
	defer myRedis.Close()
	CheckError("HTTP Dynamic Connection: getDevicesHandler", err)

	devices := myRedis.GetAllDevices()
	err = json.NewEncoder(w).Encode(devices)
	CheckError("getDevicesHandler JSON enc", err)
}

func (server *HTTPServer) getDevDataHandler(w http.ResponseWriter, r *http.Request) {

	myRedis, err := dao.MyRedis{Host: server.DBURL.DbHost, Port: server.DBURL.DbPort}.RunDBConnection()
	defer myRedis.Close()
	CheckError("HTTP Dynamic Connection: getDevDataHandler", err)

	vars := mux.Vars(r)
	devID := "device:" + vars["id"]

	devParamsKeysTokens := []string{}
	devParamsKeysTokens = strings.Split(devID, ":")
	devParamsKey := devID + ":" + "params"

	device := myRedis.GetDevice(devParamsKey, devParamsKeysTokens)
	err = json.NewEncoder(w).Encode(device)
	CheckError("getDevDataHandler JSON enc", err)
}

func (server *HTTPServer) getDevConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"] // type + mac
	mac := strings.Split(id, ":")[2]

	myRedis, err := dao.MyRedis{Host: server.DBURL.DbHost, Port: server.DBURL.DbPort}.RunDBConnection()
	defer myRedis.Close()
	CheckError("HTTP Dynamic Connection: getDevConfigHandler", err)

	configInfo := mac + ":" + "config" // key

	var config = myRedis.GetFridgeConfig(configInfo, mac)

	json.NewEncoder(w).Encode(config)
}

func (server *HTTPServer) patchDevConfigHandler(w http.ResponseWriter, r *http.Request) {

	var config *DevConfig
	vars := mux.Vars(r)
	id := vars["id"] // warning!! type : name : mac
	mac := strings.Split(id, ":")[2]

	myRedis, err := dao.MyRedis{Host: server.DBURL.DbHost, Port: server.DBURL.DbPort}.RunDBConnection()
	defer myRedis.Close()
	CheckError("HTTP Dynamic Connection: patchDevConfigHandler", err)

	configInfo := mac + ":" + "config" // key

	config =  myRedis.GetFridgeConfig(configInfo, mac)

	// log.Warnln("Config Before", config)
	err = json.NewDecoder(r.Body).Decode(&config)

	if err != nil {
		http.Error(w, err.Error(), 400)
		log.Errorln("NewDec: ", err)
	}

	log.Info(config)
	CheckError("Encode error", err) // log.Warnln("Config After: ", config)


	if !ValidateMAC(config.MAC) {
		log.Error("Invalid MAC")
		http.Error(w, "Invalid MAC", 400)
	} else if !ValidateStreamOn(config.StreamOn) {
		log.Error("Invalid Stream Value")
		http.Error(w, "Invalid Stream Value", 400)
	} else if !ValidateTurnedOn(config.TurnedOn) {
		log.Error("Invalid Turned Value")
		http.Error(w, "Invalid Turned Value", 400)
	} else if !ValidateCollectFreq(config.CollectFreq) {
		log.Error("Invalid Collect Frequency Value")
		http.Error(w, "Collect Frequency should be more than 150!", 400)
	} else if !ValidateSendFreq(config.SendFreq) {
		log.Error("Invalid Send Frequency Value")
		http.Error(w, "Send Frequency should be more than 150!", 400)
	} else {
		// Save New Configuration to DB
		myRedis.SetFridgeConfig(configInfo,config)
		log.Println("New Config was added to DB: ", config)
		JSONСonfig, _ := json.Marshal(config)
		myRedis.Publish("configChan",JSONСonfig)
	}
}