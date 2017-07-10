package devices

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/KharkivGophers/center-smart-house/dao"
	"github.com/KharkivGophers/center-smart-house/models"

	"reflect"
	"encoding/json"
	"strconv"
	"errors"
)

// ------------------------With Mock--------------------------------------------------------------
func TestSetDevConfig(t *testing.T){
	expectedErr := errors.New("err")
	dbMock := dao.DBMock{func() error {
		return expectedErr
	}}



	fridge := Fridge{}
	devConfig := models.DevConfig{true,true,
		int64(0),int64(0),"mac"}

	Convey("Should be all ok", t, func() {
		fridge.SetDevConfig("test", &devConfig, dbMock.Client)
		mustBe := " TurnedOn true CollectFreq 0 SendFreq 0 StreamOn true"

		var actual string

		for _, val := range dbMock.Client.Hash["test"]{
			arr:=val.([]interface{})
			for _, val := range arr{
				switch reflect.TypeOf(val).String() {
				case "bool":
					actual += " " + strconv.FormatBool(val.(bool))
				case "string":
					actual += " " + val.(string)
				case "int64":
					actual += " " + strconv.FormatInt(val.(int64),10)
				}
			}
		}
		So(actual, ShouldEqual, mustBe)
	})
}

func TestGetDevConfig(t *testing.T){
	dbMock := dao.DBMock{&dao.DbMockClient{}}
	fridge := Fridge{}
	devConfig := models.DevConfig{true,true,
				      int64(0),int64(0),"00-00-00-11-11-11"}

	Convey("Should be all ok", t, func() {
		fridge.SetDevConfig("test", &devConfig, dbMock.Client)
		config := fridge.GetDevConfig("test", "00-00-00-11-11-11", dbMock.Client)
		So(reflect.DeepEqual(*config, devConfig), ShouldBeTrue)
	})
}



//------------------------------------------------------------
func TestSetDevConfigWithRedis(t *testing.T) {
	dbWorker := dao.RedisClient{DbServer: models.Server{IP: "0.0.0.0", Port: uint(6379)}}
	dbWorker.Connect()
	defer dbWorker.Close()
	fridge := Fridge{}

	fridge.Config = FridgeConfig{true, true, int64(200), int64(200)}
	b,_ := json.Marshal(fridge.Config)
	devConfig := models.DevConfig{"00-00-00-11-11-11", b}

	Convey("Should be all ok", t, func() {
		fridge.SetDevConfig("00-00-00-11-11-11:config", &devConfig, dbWorker.Client)
		isConfig := fridge.GetDevConfig("00-00-00-11-11-11:config", "00-00-00-11-11-11", dbWorker.Client)
		dbWorker.FlushAll()
		So(reflect.DeepEqual(*isConfig, devConfig), ShouldBeTrue)
	})
}

// Impossible to testing StreamOn and TurnedOn
func TestValidateDevData(t *testing.T) {
	dbWorker := dao.RedisClient{DbServer: models.Server{IP: "0.0.0.0", Port: uint(6379)}}
	dbWorker.Connect()
	defer dbWorker.Close()
	fridge := Fridge{}
	var devConfig models.DevConfig
	Convey("Invalid MAC. Should be false", t, func() {
		fridge.Config = FridgeConfig{true, true, int64(0), int64(0)}
		b,_ := json.Marshal(fridge.Config)
		devConfig = models.DevConfig{"Invalid mac", b}
		valid, _ := fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeFalse)
	})
	Convey("Valid DevConfig. Should be false", t, func() {

		fridge.Config = FridgeConfig{true, true, int64(200), int64(200)}
		b,_ := json.Marshal(fridge.Config)
		devConfig = models.DevConfig{"00-00-00-11-11-11", b}

		valid, _ := fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeTrue)
	})
	Convey("Collect Frequency should be more than 150!", t, func() {

		fridge.Config = FridgeConfig{true, true, int64(100), int64(200)}
		b,_ := json.Marshal(fridge.Config)
		devConfig = models.DevConfig{"00-00-00-11-11-11", b}
		valid, _ := fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeFalse)
	})
	Convey("Send Frequency should be more than 150!", t, func() {
		fridge.Config = FridgeConfig{true, true, int64(200), int64(100)}
		b,_ := json.Marshal(fridge.Config)
		devConfig = models.DevConfig{"00-00-00-11-11-11", b}
		valid, _ := fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeFalse)
	})
	dbWorker.FlushAll()
}

func TestSmallSetDevData(t *testing.T) {
	dbWorker := dao.RedisClient{DbServer: models.Server{IP: "0.0.0.0", Port: uint(6379)}}
	dbWorker.Connect()
	defer dbWorker.Close()

	tempCam := make(map[int64]float32)
	tempCam[1] = 1.0
	tempCam[2] = 2.0

	key := "test"

	Convey("", t, func() {
		expected := []string{"1:1","2:2"}
		setDevData(tempCam, key, dbWorker.Client)
		actual, _ := dbWorker.Client.ZRangeByScore(key, "-inf", "inf")
		So(actual, ShouldResemble, expected)
	})
	dbWorker.FlushAll()
}

func TestSetDevData(t *testing.T) {
	dbWorker := dao.RedisClient{DbServer: models.Server{IP: "0.0.0.0", Port: uint(6379)}}
	dbWorker.Connect()
	defer dbWorker.Close()

	tempCam := make(map[int64]float32)
	tempCam[1] = 1.0
	tempCam[2] = 2.0

	dataMap := make(map[string][]string)
	dataMap["TempCam1"]= []string{"1:1","2:2"}
	dataMap["TempCam2"]= []string{"1:1","2:2"}

	meta := models.DevMeta{MAC:"00-00-00-11-11-11", Name:"name",Type:"fridge"}
	data :=FridgeData{tempCam,tempCam}
	fridge := Fridge{}

	b,_:=json.Marshal(data)

	req := models.Request{Meta:meta, Data:b}

	Convey("Must bu all ok", t, func() {

		fridge.SetDevData(&req, dbWorker.Client)
		dbWorker.Connect()
		devParamsKey:="device:" +meta.Type +":"+meta.Name+":"+meta.MAC+":params"

		actual := fridge.GetDevData(devParamsKey,meta,dbWorker.Client)
		expected := models.DevData{Meta:meta,Data:dataMap}
		dbWorker.FlushAll()
		So(actual, ShouldResemble, expected)
	})
}

