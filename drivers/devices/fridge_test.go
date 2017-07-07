package devices

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/KharkivGophers/center-smart-house/dao"
	"github.com/KharkivGophers/center-smart-house/models"


	"reflect"
	"strconv"
)

// ------------------------With Mock--------------------------------------------------------------
func TestSetDevConfig(t *testing.T){
	dbMock := dao.DBMock{&dao.DbMockClient{}}
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


// Impossible to testing StreamOn and TurnedOn
func TestValidateDevData(t *testing.T){
	fridge := Fridge{}
	var devConfig models.DevConfig
	Convey("Invalid MAC. Should be false", t, func() {
		devConfig = models.DevConfig{true,true,
					      int64(0),int64(0),"Invalid mac"}
		valid, _ :=fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeFalse)
	})
	Convey("Valid DevConfig. Should be false", t, func() {
		devConfig = models.DevConfig{true,true,
					     int64(200),int64(200),"00-00-00-11-11-11"}
		valid, _ :=fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeTrue)
	})
	Convey("Collect Frequency should be more than 150!", t, func() {
		devConfig = models.DevConfig{true,true,
					     int64(100),int64(200),"00-00-00-11-11-11"}
		valid, _ :=fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeFalse)
	})
	Convey("Send Frequency should be more than 150!", t, func() {
		devConfig = models.DevConfig{true,true,
					     int64(200),int64(100),"00-00-00-11-11-11"}
		valid, _ :=fridge.ValidateDevData(devConfig)
		So(valid, ShouldBeFalse)
	})
}

//------------------------------------------------------------

