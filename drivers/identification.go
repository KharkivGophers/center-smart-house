package drivers

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/KharkivGophers/center-smart-house/drivers/devices"
)

func IdentifyDevConfig(devType string)(*DevConfigDriver){
	var device DevConfigDriver

	switch devType {
	case "fridge":
		device = &Fridge{}
	case "washer":
		device = &Washer{}
	default:
		log.Println("Device request: unknown device type")
		return nil
}
	return &device
}


func IdentifyDevData(devType string)(DevDataDriver){
	var device DevDataDriver

	switch devType {
	case "fridge":
		device = &Fridge{}
	case "washer":
		device = &Washer{}
	default:
		log.Println("Device request: unknown device type")
		return nil
	}
	return device
}
