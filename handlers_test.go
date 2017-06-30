package main

import (
	"testing"
	"net"
	"encoding/json"
	"bytes"
	"strings"
	"net/http"
	"io/ioutil"

	"menteslibres.net/gosexy/redis"

	"github.com/KharkivGophers/center-smart-house/dao"
	"github.com/gorilla/websocket"
	. "github.com/smartystreets/goconvey/convey"
	"time"
)

var timeForSleep time.Duration = 1000 * time.Millisecond

func deleteAllInBase(dbClient dao.DbWorker) {
	err := dbClient.FlushAll()
	CheckError("Some error with FlushAll()", err)
}

//func TestDevTypeHandler(t *testing.T) {
//
//	//Create redis client------------------------------------------------------------
//	var myRedis dao.DbWorker = &dao.MyRedis{}
//	myRedis.Connect()
//	defer myRedis.Close()
//	//--------------------------------------------------------------------------------
//
//	Convey("Message handles correct", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(myRedis)
//	})
//	Convey("Empty message", t, func() {
//		req := Request{}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown action")
//		deleteAllInBase(myRedis)
//	})
//	Convey("Type washer message", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "washer", Name: "bosh", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(myRedis)
//	})
//	Convey("Unknown type message", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "nil", Name: "bosh", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
//		deleteAllInBase(myRedis)
//	})
//	// need to change handlers
//	Convey("Empty MAC", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: ""}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(myRedis)
//	})
//	Convey("Empty Type", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
//		deleteAllInBase(myRedis)
//	})
//	Convey("Empty Name", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(myRedis)
//	})
//	Convey("Empty Time", t, func() {
//		req := Request{Action: "update", Time: 0, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(myRedis)
//	})
//
//}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestSendJSONToServer(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()

	//Create redis client------------------------------------------------------------
	var myRedis dao.DbWorker = &dao.MyRedis{}
	myRedis.Connect()
	defer myRedis.Close()
	//--------------------------------------------------------------------------------

	res := "\"status\":200,\"descr\":\"Data has been delivered successfully\""
	req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
	message, _ := json.Marshal(req)
	conn.Write(message)
	time.Sleep(timeForSleep)
	buffer := make([]byte, 1024)

	for i := 0; i == 0; {
		i, _ = conn.Read(buffer)
	}
	response := bytes.NewBuffer(buffer).String()

	if !strings.Contains(response, res) {
		t.Error("Bad JSON", response, res)
	}
	deleteAllInBase(myRedis)
}

func TestCheckJSONToServer(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()

	res := "\"status\":200,\"descr\":\"Data has been delivered successfully\""

	//Create redis client------------------------------------------------------------
	var myRedis dao.DbWorker = &dao.MyRedis{}
	myRedis.Connect()
	defer myRedis.Close()
	//--------------------------------------------------------------------------------


	Convey("Send Correct JSON to server", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("Warning! Uncorrect JSON was sent to server", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("JSON was sent to server. Action of fridge should be update", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("Warning! JSON was sent to server with uncorrect action value", t, func() {
		req := Request{Action: "nil", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("JSON was sent to server. Action of washer should be update", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "washer", Name: "bosh0e31", MAC: "00-15-E9-2B-99-3B"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("Warning! JSON was sent to server with uncorrect type value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "nil", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("Warning! JSON was sent to server without MAC value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: ""}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("Warning! JSON was sent to server without type value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("Warning! JSON was sent to server without name value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})

	Convey("Warning! JSON was sent to server without time value ", t, func() {
		req := Request{Action: "update", Time: 0, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		conn.Write(message)
		time.Sleep(timeForSleep)
		buffer := make([]byte, 1024)
		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(myRedis)
	})
}
func TestRedisConnection(t *testing.T) {
	client := redis.New()
	Convey("Check redis client connection"+dbHost+":"+string(dbPort)+". Should be without error ", t, func() {
		err := client.Connect(dbHost, dbPort)
		defer client.Close()
		So(err, ShouldBeNil)
	})
}
func TestHTTPConnection(t *testing.T) {
	var httpClient = &http.Client{}

	//Create redis client------------------------------------------------------------
	var myRedis dao.DbWorker = &dao.MyRedis{}
	myRedis.Connect()
	defer myRedis.Close()
	//--------------------------------------------------------------------------------

	Convey("Check http://"+connHost+":"+httpConnPort+"/devices/{id}/data. Should be without error ", t, func() {
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "//devices/fridge:hladik0e31:00-15-E9-2B-99-3C/data")
		So(res, ShouldNotBeNil)
		deleteAllInBase(myRedis)
	})
	Convey("Check http://"+connHost+":"+httpConnPort+"/devices. Should be without error ", t, func() {
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices")
		So(res, ShouldNotBeNil)
	})
}

func TestWorkingServerAfterSendingJSON(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()
	var httpClient = &http.Client{}

	connForDAta, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()

	//Create redis client------------------------------------------------------------
	var myRedis dao.DbWorker = &dao.MyRedis{}
	myRedis.Connect()
	defer myRedis.Close()
	//--------------------------------------------------------------------------------

	Convey("Send correct JSON. Should be return all ok ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"10\":10.5}}}"
		reqConfig := "{\"action\":\"config\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"10\":10.5}}}"
		mustHave := "[{\"site\":\"\",\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\",\"mac\":\"00-15-E9-2B-99-3C\"," +
			"\"ip\":\"\"},\"data\":{\"TempCam1\":[\"10:10.5\"],\"TempCam2\":[\"10:10.5\"]}}]"
		conn.Write([]byte(reqConfig))
		connForDAta.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(myRedis)
	})
	Convey("Send JSON where action = wrongValue. Should not be return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"wrongValue\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName2\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"TempCam1\":[\"10:10.5\"]," +
			"\"TempCam2\":[\"1500:15.5\"]}}"

		mustNotHave := "testName2"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(myRedis)
	})
	Convey("Send JSON where type = wrongValue. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"wrongValue\",\"name\":\"testName3\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave := "testName3"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(myRedis)
	})

	Convey("Send JSON without name. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave := "TestMACFridge3"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(myRedis)
	})
	Convey("Send JSON without mac. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"config\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"fridge4\"" +
			",\"mac\":\"\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave := "fridge4"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(myRedis)
	})

	Convey("Send JSON with wrong data. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"fridge5\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"qwe\":qwe},\"tempCam2\":{\"" +
			"qwe\":qwe}}}"

		mustNotHave := "fridge5"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotBeNil)
		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(myRedis)
	})
//	// my part
	Convey("Send correct JSON. Initialize turned on as false ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"turnedOn\":false"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices/fridge:testName1:00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(myRedis)
	})
	Convey("Send correct JSON. Initialize CollectFreq as 0 ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"collectFreq\":0"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices/fridge:testName1:00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(myRedis)
	})
	Convey("Send correct JSON. Initialize SendFreq as 0 ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"sendFreq\":0"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices/fridge:testName1:00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(myRedis)
	})
	Convey("Send correct JSON. Initialize StreamOn as false ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"streamOn\":false"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + connHost + ":" + httpConnPort + "/devices/fridge:testName1:00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(myRedis)

	})
	Convey("Send correct JSON. Patch device data: turned on as true ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"turnedOn\":false"
		conn.Write([]byte(reqMessage))
		url := "http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:00-15-E9-2B-99-3C/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"turnedOn\":true}")))
		httpClient.Do(r)
		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(myRedis)
	})

	Convey("Send correct JSON. Patch device data: stream on as true ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"streamOn\":false"
		conn.Write([]byte(reqMessage))
		url := "http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:00-15-E9-2B-99-3C/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"streamOn\":true}")))
		time.Sleep(timeForSleep)
		httpClient.Do(r)

		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(myRedis)
	})
}

func TestWSConnection(t *testing.T) {

	req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge",
		Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
	 mustBe :="{\"action\":\"update\",\"time\":1496741392463499334,\"meta\":{\"type\":\"fridge\",\"name\":\"hladik0e31\",\"" +
		 "mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":null}"

	//Create redis client------------------------------------------------------------
	var myRedis dao.DbWorker = &dao.MyRedis{}
	myRedis.Connect()
	defer myRedis.Close()
	//--------------------------------------------------------------------------------

	Convey("Checking how to work ws connection. Should be true", t, func() {
		//Create Web Socket connection from the client side--------------------------------
		url := "ws://" + connHost + ":" + wsConnPort + "/devices/00-15-E9-2B-99-3C"
		var dialer *websocket.Dialer
		conn, _, _ := dialer.Dial(url, nil)
		//---------------------------------------------------------------------------------
		publishWS(req)
		_, message, _ := conn.ReadMessage()
		So(bytes.NewBuffer(message).String(), ShouldEqual, mustBe)
	})
}

