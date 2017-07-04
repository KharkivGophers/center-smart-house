package drivers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/KharkivGophers/center-smart-house/drivers/devices"
	. "github.com/KharkivGophers/center-smart-house/models"
)

func IdentDevRequest(req Request)(*ConfigDevDriver){
	var (
		device ConfigDevDriver
	)
	switch req.Meta.Type {
	case "fridge":
		device = &devices.Fridge{}
	case "washer":
	default:
		log.Println("Device request: unknown device type")
		return nil
	}
	return &device
}

func IdentDevString(typeDev string)(*ConfigDevDriver){
	var (
		device ConfigDevDriver
	)
	switch typeDev {
	case "fridge":
		device = &devices.Fridge{}
	case "washer":
	default:
		log.Println("Device request: unknown device type")
}
	return &device
}
