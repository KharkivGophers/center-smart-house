package common

import (
	"menteslibres.net/gosexy/redis"
	"strconv"
	"log"
	"time"
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
