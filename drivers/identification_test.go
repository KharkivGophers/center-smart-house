package drivers

import (
"testing"
. "github.com/smartystreets/goconvey/convey"
"reflect"
)

func TestIdentifyDev(t *testing.T) {

	Convey("Fridge. Should return empty fridge", t, func() {
		thisFridge := IdentifyDev("fridge")
		mustBe := "*devices.Fridge"
		So(mustBe,ShouldEqual, reflect.TypeOf(*thisFridge).String())
	})

	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDev("washer")
		mustBe := "*devices.Washer"
		So(mustBe,ShouldEqual, reflect.TypeOf(*thisWasher).String())
	})
	Convey("Washer. Should return empty washer", t, func() {
		thisWasher := IdentifyDev("Something")
		So(thisWasher,ShouldBeNil)
	})
}
