package devices

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/KharkivGophers/center-smart-house/driver"
)

func IdentifyDev(devType string)(*ConfigDevDriver){
	var device ConfigDevDriver

	switch devType {
	case "fridge":
		device = &Fridge{}
	case "washer":
	default:
		log.Println("Device request: unknown device type")
		return nil
}
	return &device
}
