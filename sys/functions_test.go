package sys
import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"

	"errors"
)

func TestCheckError(t *testing.T) {

	Convey("CheckError err = nil. Should return nil", t, func() {
	actual := CheckError("", nil)
		So(actual,ShouldBeNil)
	})

	Convey("CheckError err = not nil. Should return nil", t, func() {
		err := errors.New("Error")
		actual := CheckError("", err)
		So(actual,ShouldResemble, err)
	})

}

func TestFloat32ToString(t *testing.T) {

	Convey("Float64ToString. Value = 23.3", t, func() {
		expected := "23.2"
		actual := Float64ToString(23.2)
		So(actual,ShouldEqual, expected)
	})
	Convey("Float64ToString. Value = -9 223 372 036 854 775 808 .0 ", t, func() {
		expected := "-9223372036854775808.0 "
		actual := Float64ToString(-9223372036854775808.0 )
		So(actual,ShouldEqual, expected)
	})
	Convey("Float64ToString. Value =  9 223 372 036 854 775 807.0", t, func() {
		expected := "9223372036854775807.0"
		actual := Float64ToString( 9223372036854775807.0)
		So(actual,ShouldEqual, expected)
	})
	Convey("Float64ToString. Value = 0", t, func() {
		expected := "0"
		actual := Float64ToString(0)
		So(actual,ShouldEqual, expected)
	})


}

func TestInt64ToString(t *testing.T) {

	Convey("Float64ToString. Value = 23", t, func() {
		expected := "23"
		actual := Int64ToString(23)
		So(actual,ShouldEqual, expected)
	})
	Convey("Float64ToString. Value = 9223372036854775807", t, func() {
		expected := "9223372036854775807"
		actual := Int64ToString(9223372036854775807)
		So(actual,ShouldEqual, expected)
	})
	Convey("Float64ToString. Value = -9223372036854775808", t, func() {
		expected := "-9223372036854775808"
		actual := Int64ToString(-9223372036854775808)
		So(actual,ShouldEqual, expected)
	})
	Convey("Float64ToString. Value = 0", t, func() {
		expected := "0"
		actual := Int64ToString(0)
		So(actual,ShouldEqual, expected)
	})
}


func TestValidateMAC(t *testing.T) {

	Convey("Float64ToString. MAC = 00-00-00-00-00-00", t, func() {
		actual := ValidateMAC("00-00-00-00-00-00")
		So(actual,ShouldBeTrue)
	})
	Convey("Float64ToString. MAC = 00-00-00-00-00-00", t, func() {
		actual := ValidateMAC("00-00-00-00-00-")
		So(actual,ShouldBeFalse)
	})
	Convey("Float64ToString. MAC = 12345678912345678", t, func() {
		actual := ValidateMAC("12345678912345678")
		So(actual,ShouldBeFalse)
	})

}