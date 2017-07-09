package drivers

import (
"testing"
. "github.com/smartystreets/goconvey/convey"
"reflect"
)

func TestIdentifyDevConfig(t *testing.T) {

	Convey("Fridge. Should return empty fridge", t, func() {
		thisFridge := IdentifyDevConfig("fridge")
		mustBe := "*devices.Fridge"
		So(mustBe,ShouldEqual, reflect.TypeOf(thisFridge).String())
	})

	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDevConfig("washer")
		mustBe := "*devices.Washer"
		So(mustBe,ShouldEqual, reflect.TypeOf(thisWasher).String())
	})
	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDevConfig("Something")
		So(thisWasher,ShouldBeNil)
	})
}

func TestIdentifyDevData(t *testing.T) {

	Convey("Fridge. Should return empty fridge", t, func() {
		thisFridge := IdentifyDevData("fridge")
		mustBe := "*devices.Fridge"
		So(mustBe,ShouldEqual, reflect.TypeOf(thisFridge).String())
	})

	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDevData("washer")
		mustBe := "*devices.Washer"
		So(mustBe,ShouldEqual, reflect.TypeOf(thisWasher).String())
	})
	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDevData("Something")
		So(thisWasher,ShouldBeNil)
	})
}
