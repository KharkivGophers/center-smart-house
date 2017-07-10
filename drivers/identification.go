package drivers

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/KharkivGophers/center-smart-house/drivers/devices"
)

func IdentifyDevice(devType string)(DevDriver){
	var device DevDriver

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

