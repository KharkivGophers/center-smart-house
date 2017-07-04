package sysFunc

import (
	log "github.com/Sirupsen/logrus"
	"strconv"
	. "github.com/KharkivGophers/center-smart-house/models"
	"errors"
)

//-----------------Common functions-------------------------------------------------------------------------------------------

func CheckError(desc string, err error) error {
	if err != nil {
		log.Errorln(desc, err)
		return err
	}
	return nil
}

func Float32ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 32)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(int64(n), 10)
}

// Validate MAC got from User
func ValidateMAC(mac interface{}) bool {
	switch v := mac.(type) {
	case string:
		switch len(v) {
		case 17:
			return true
		default:
			log.Error("MAC should contain 17 symbols")
			return false
		}
	default:
		log.Error("MAC should be in string format")
		return false
	}
}

// Validate Send Frequency Value got from User
func ValidateSendFreq(sendFreq interface{}) bool {
	switch v := sendFreq.(type) {
	case int64:
		switch {
		case v > 150:
			return true
		default:
			log.Error("Send Frequency should be more than 150!")
			return false
		}
	default:
		log.Error("Send Frequency should be in int64 format")
		return false
	}
}

// Validate Collect Frequency got from User
func ValidateCollectFreq(collectFreq interface{}) bool {
	switch v := collectFreq.(type) {
	case int64:
		switch {
		case v > 150:

			return true
		default:
			log.Error("Collect Frequency should be more than 150!")
			return false
		}
	default:
		log.Error("Collect Frequency should be in int64 format")
		return false
	}
}

// Validate TurnedOn Value got from User
func ValidateTurnedOn(turnedOn interface{}) bool {
	switch turnedOn.(type) {
	case bool:
		return true
	default:
		log.Error("TurnedOn should be in bool format!")
		return false
	}
}

// Validate StreamOn Value got from User
func ValidateStreamOn(streamOn interface{}) bool {
	switch streamOn.(type) {
	case bool:
		return true
	default:
		log.Error("StreamOn should be in bool format!")
		return false
	}
}

func ValidateDevMeta(meta DevMeta) (bool, error) {
	var err string
	if !ValidateMAC(meta.MAC) {
		log.Error("Invalid MAC")
		err += "Invalid MAC. "
	}
	if err != "" {
		return false, errors.New(err)
	}
	return true, nil
}
