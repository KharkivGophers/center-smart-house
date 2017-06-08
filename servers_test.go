package main

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"net"
	"encoding/json"
	"bytes"
	"strings"
)

func TestSendJSONToServer1(t *testing.T) {
	host:= "192.168.104.60"
	port:= "3030"
	resToCheck := "{\"status\":200,\"descr\":\"Data have been delivered successfully\"}"
	conn, _ := net.Dial("tcp", host+":"+port)
	defer conn.Close()

	req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"fridge",
		Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
	message, _ := json.Marshal(req)

	//simpleupdate
	conn.Write(message)
	//simple Read
	buffer := make([]byte, 1024)

	for i :=0; i==0;{
		i,_ =conn.Read(buffer)
	}
	res := bytes.NewBuffer(buffer).String()

	if !strings.Contains(res, resToCheck) {
		t.Error("asdfasdfasdfasdfasdfasdfasdf ")
	}


}
func TestSendJSONToServer(t *testing.T) {
	host:= "192.168.104.60"
	port:= "3030"
	resToCheck := "{\"status\":200,\"descr\":\"Data have been delivered successfully\"}"
	conn, err := net.Dial("tcp", host+":"+port)
	defer conn.Close()

	Convey("TCP connections should be without error", t, func() {
		So(err, ShouldBeNil)
	})

	Convey("Send correct JSON to server. Should be return 200", t, func() {

		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"fridge",
			Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)

		//simpleupdate
		conn.Write(message)
		//simple Read
		buffer := make([]byte, 1024)

		for i :=0; i==0;{
			i,_ =conn.Read(buffer)
		}
		res := bytes.NewBuffer(buffer).String()

		So(res, ShouldContainSubstring, resToCheck)
	})

	Convey("Send to server correct JSON where action it is wrong value. Should be return 200. Check default case", t, func() {

		req := Request{Action:"wrongValue",Time:1496741392463499334, Meta:DevMeta{Type:"fridge",
			Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)

		//simple
		conn.Write(message)
		//simple Read
		buffer := make([]byte, 1024)

		for i :=0; i==0;{
			i,_ =conn.Read(buffer)
		}
		res := bytes.NewBuffer(buffer).String()

		So(res, ShouldContainSubstring, resToCheck)
	})

	Convey("Send to server correct JSON where type it is wrong value. Should be return 200. Check default case", t, func() {

		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"wrongValue",
			Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)

		//simple
		conn.Write(message)
		//simple Read
		buffer := make([]byte, 1024)

		for i :=0; i==0;{
			i,_ =conn.Read(buffer)
		}
		res := bytes.NewBuffer(buffer).String()

		So(res, ShouldContainSubstring, resToCheck)
	})

}

