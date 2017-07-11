package drivers

import (
"testing"
. "github.com/smartystreets/goconvey/convey"
"reflect"
)

func TestIdentifyDevice(t *testing.T) {

	Convey("Fridge. Should return empty fridge", t, func() {
		thisFridge := IdentifyDevice("fridge")
		mustBe := "*devices.Fridge"
		So(mustBe,ShouldEqual, reflect.TypeOf(thisFridge).String())
	})

	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDevice("washer")
		mustBe := "*devices.Washer"
		So(mustBe,ShouldEqual, reflect.TypeOf(thisWasher).String())
	})
	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDevice("Something")
		So(thisWasher,ShouldBeNil)
	})
}


