package http

import (
	"net/http"
	"github.com/gorilla/mux"
	"time"
	"github.com/gorilla/handlers"
	log "github.com/Sirupsen/logrus"
	"encoding/json"


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

	log.Info(dbClient.GetClient())

	devMeta := DevMeta{r.FormValue("type"),r.FormValue("name"),r.FormValue("mac"),""}

	devID := "device:" + devMeta.Type+":"+devMeta.Name+":"+devMeta.MAC
	devParamsKey := devID + ":" + "params"

	device := IdentifyDevData(devMeta.Type)
	deviceData := device.GetDevData(devParamsKey, devMeta, dbClient.GetClient())

	err := json.NewEncoder(w).Encode(deviceData)
	CheckError("getDevDataHandler JSON enc", err)
}

func (server *HTTPServer) getDevConfigHandler(w http.ResponseWriter, r *http.Request) {
	devMeta := DevMeta{r.FormValue("type"),r.FormValue("name"),r.FormValue("mac"),""}
	flag, err :=ValidateDevMeta(devMeta)
	if !flag {
		log.Errorf("getDevConfigHandler. %v. Exit the method",err)
		return
	}

	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	configInfo := dbClient.GetKeyForConfig(devMeta.MAC) // key

	var device DevConfigDriver = *IdentifyDevConfig(devMeta.Type)

	if device==nil{
		http.Error(w, "This type is not found", 400)
		return
	}
	config := device.GetDevConfig(configInfo, devMeta.MAC, dbClient.GetClient())
	w.Write(config.Data)
}

func (server *HTTPServer) patchDevConfigHandler(w http.ResponseWriter, r *http.Request) {



	devMeta := DevMeta{r.FormValue("type"),r.FormValue("name"),r.FormValue("mac"),""}

	_, err :=ValidateDevMeta(devMeta)
	if err != nil {
		log.Error(err)
		return
	}

	dbClient := GetDBConnection(server.DbServer)
	defer dbClient.Close()

	var device DevServerHandler = IdentifyDevHandler(devMeta.Type)
	if device==nil{
		http.Error(w, "This type is not found", 400)
		return
	}

	device.PatchDevConfigHandlerHTTP(w,r,devMeta,dbClient)


}