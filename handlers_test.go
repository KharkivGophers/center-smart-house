package main

import (
	"testing"
	"net"
	. "github.com/smartystreets/goconvey/convey"
	"encoding/json"
	"bytes"
	"strings"
)

func TestDevTypeHandler(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()
	Convey("Message handles correct", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"fridge", Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Empty message", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown action")
	})
	Convey("Type washer message", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"washer", Name:"bosh",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Unknown type message", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"nil", Name:"bosh",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
	})
	// need to change handlers
	Convey("Empty MAC", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"fridge", Name:"hladik0e31",MAC:""}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Empty Type", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"", Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
	})
	Convey("Empty Name", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"fridge", Name:"",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Empty Time", t, func() {
		conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
		req := Request{Action:"update",Time:0, Meta:DevMeta{Type:"fridge", Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})

}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestSendJSONToServer(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()

	res := "{\"status\":200,\"descr\":\"Data have been delivered successfully\"}"
	req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
	message, _ := json.Marshal(req)
	conn.Write(message)

	buffer := make([]byte, 1024)

	for i := 0; i == 0; {
		i, _ = conn.Read(buffer)
	}
	response := bytes.NewBuffer(buffer).String()

	if !strings.Contains(response, res) {
		t.Error("Bad JSON")
	}
}

func TestCheckJSONToServer(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()

	res := "{\"status\":200,\"descr\":\"Data have been delivered successfully\"}"

	Convey("Send Correct JSON to server", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("Warning! Uncorrect JSON was sent to server", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("JSON was sent to server. Action of fridge should be update", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("Warning! JSON was sent to server with uncorrect action value", t, func() {
		req := Request{Action: "nil", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("JSON was sent to server. Action of washer should be update", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "washer", Name: "bosh0e31", MAC: "00-15-E9-2B-99-3B"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("Warning! JSON was sent to server with uncorrect type value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "nil", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("Warning! JSON was sent to server without MAC value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: ""}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("Warning! JSON was sent to server without type value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("Warning! JSON was sent to server without name value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})

	Convey("Warning! JSON was sent to server without time value ", t, func() {
		req := Request{Action: "update", Time: 0, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)

		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
	})
}
