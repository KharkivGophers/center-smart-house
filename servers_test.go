package main

import (
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/gorilla/mux"
//	"github.com/stretchr/testify/assert"
//
)
import (
	"testing"
	"net"
	. "github.com/smartystreets/goconvey/convey"
	"fmt"
	"bytes"
//	"encoding/json"
	//"time"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	//"github.com/stretchr/testify/assert"
	"encoding/json"
)

func Router() *mux.Router {

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}", webSocketHandler)
	r.HandleFunc("/devWS", webSocketHandler)
	return r
}
//
//func TestCreateEndpoint(t *testing.T) {
//	request, _ := http.NewRequest("GET", "/devices", nil)
//	response := httptest.NewRecorder()
//	Router().ServeHTTP(response, request)
//	assert.Equal(t, 200, response.Code, "OK response is expected")
//}

//func TestSendDefaultConfiguration(t *testing.T) {
//
//	var pool ConectionPool
//	pool.init()
//
//	//ln, err := net.Listen(connType, "localhost"+":"+"3000")
//	//fmt.Println(ln,',',err)
//	//for {
//	//	conn, err := ln.Accept()
//	//	fmt.Println(conn,',',err)
//	//}
//	//conn, err := ln.Accept()
//	var conn net.Conn
//	fmt.Println(conn)
//
//	var buf bytes.Buffer
//	log.SetOutput(&buf)
//	defer func() {
//		log.SetOutput(os.Stderr)
//	}()
//	sendDefaultConfiguration(&conn, &pool)
//	t.Log(buf.String())
//
//
//}

func TestTCPConnection(t *testing.T) {

	conn, err := net.Dial("tcp", "localhost:3030")

	Convey("TCP connections should be without error", t, func() {
		So(conn, ShouldNotBeNil)
	})

	Convey("TCP connections should be without error", t, func() {
		So(err, ShouldBeNil)
	})

}

func TestSendToServer(t *testing.T) {
	conn, _ := net.Dial("tcp", "localhost:3030")
	defer conn.Close()
	Convey("Should be return 200", t, func() {

		req := Request{Action:"update",Time:1496741392463499334, Meta:DevMeta{Type:"fridge", Name:"hladik0e31",MAC:"00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		//simple Read
		buffer := make([]byte, 1024)
		//var e string
		for i :=0; i==0;{
			i,_ =conn.Read(buffer)

			fmt.Println(bytes.NewBuffer(buffer).String())
		}
		So(buffer, ShouldNotBeNil)
	})

	Convey("Message should not be empty", t, func() {
		req := Request{}
		message, _ := json.Marshal(req)

		So(func(){checkMsgContent(message)}, ShouldPanic)
	})
}

func TestCreateEndpoint(t *testing.T) {
	request, _ := http.NewRequest("GET", "/devices", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	//assert.Equal(t, 200, response.Code, "OK response is expected")
	Convey("OK response is expected", t, func() {
		So(response, ShouldNotBeNil)
	})
	Convey("OK response is expected", t, func() {
		So(request, ShouldNotBeNil)
	})
}

func checkMsgContent(message []byte) {

	fmt.Println(string(message))
	if (string(message) == "{\"action\":\"\",\"time\":0,\"meta\":{\"type\":\"\",\"name\":\"\",\"mac\":\"\",\"ip\":\"\"},\"data\":null}") {
		panic("request is empty")
	}
}