package devices

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/KharkivGophers/center-smart-house/driver"
	. "github.com/KharkivGophers/center-smart-house/models"
)

func IdentifyDevRequest(req Request)(*ConfigDevDriver){
	var (
		device ConfigDevDriver
	)
	switch req.Meta.Type {
	case "fridge":
		device = &Fridge{}
	case "washer":
	default:
		log.Println("Device request: unknown device type")
		return nil
	}
	return &device
}

func IdentifyDevString(typeDev string)(*ConfigDevDriver){
	var (
		device ConfigDevDriver
	)
	switch typeDev {
	case "fridge":
		device = &Fridge{}
	case "washer":
	default:
		log.Println("Device request: unknown device type")
}
	return &device
}
