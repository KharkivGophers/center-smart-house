package devices

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/KharkivGophers/center-smart-house/dao"
	"github.com/KharkivGophers/center-smart-house/models"
	"fmt"
)
func TestSetDevConfig(t *testing.T){
	dbMock := dao.DBMock{&dao.DbMockClient{}}
	fridge := Fridge{}
	devConfig := models.DevConfig{true,true,
		int64(0),int64(0),"mac"}

	Convey("Should be all ok", t, func() {
		fridge.SetDevConfig("test", &devConfig, dbMock.Client)
		fmt.Println(dbMock.Client.Hash)
		So(fridge, ShouldNotBeNil)
	})
}
