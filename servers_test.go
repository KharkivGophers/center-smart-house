package main

import (
	//"net/http"
	//"net/http/httptest"
	//"testing"
	//
	//"github.com/gorilla/mux"
	//"github.com/stretchr/testify/assert"
	"net"

	"testing"
	"fmt"
	"bytes"
	"log"
	"os"
)

//func Router() *mux.Router {
//
//	r := mux.NewRouter()
//	r.HandleFunc("/devices/{id}", webSocketHandler)
//	r.HandleFunc("/devWS", webSocketHandler)
//	return r
//}
//
//func TestCreateEndpoint(t *testing.T) {
//	request, _ := http.NewRequest("GET", "/devices", nil)
//	response := httptest.NewRecorder()
//	Router().ServeHTTP(response, request)
//	assert.Equal(t, 200, response.Code, "OK response is expected")
//}

func TestSendDefaultConfiguration(t *testing.T) {

	var pool ConectionPool
	pool.init()

	//ln, err := net.Listen(connType, "localhost"+":"+"3000")
	//fmt.Println(ln,',',err)
	//for {
	//	conn, err := ln.Accept()
	//	fmt.Println(conn,',',err)
	//}
	//conn, err := ln.Accept()
	var conn net.Conn
	fmt.Println(conn)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	sendDefaultConfiguration(&conn, &pool)
	t.Log(buf.String())


}