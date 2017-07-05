package drivers

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/KharkivGophers/center-smart-house/drivers/devices"
)

func IdentifyDev(devType string)(*DevConfigDriver){
	var device DevConfigDriver

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
