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
	. "github.com/KharkivGophers/center-smart-house/models"
	. "github.com/KharkivGophers/center-smart-house/sys"
	"time"
	"fmt"
	log "github.com/Sirupsen/logrus"
)

var timeForSleep time.Duration = 1000 * time.Millisecond

func deleteAllInBase(dbClient dao.DbDriver) {
	defer treatmentPanic("Recovered in TestCheckJSONToServer")
	err := dbClient.FlushAll()
	CheckError("Some error with FlushAll()", err)
}

func treatmentPanic(message string) {
	if r := recover(); r != nil {
		fmt.Println(message, r)
	}
}

//func TestDevTypeHandler(t *testing.T) {
//
//	//Create redis client------------------------------------------------------------
//	var dbCli dao.DbClient = &dao.MyRedis{}
//	dbCli.Connect()
//	defer dbCli.Close()
//	//--------------------------------------------------------------------------------
//
//	Convey("Message handles correct", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(dbCli)
//	})
//	Convey("Empty message", t, func() {
//		req := Request{}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown action")
//		deleteAllInBase(dbCli)
//	})
//	Convey("Type washer message", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "washer", Name: "bosh", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(dbCli)
//	})
//	Convey("Unknown type message", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "nil", Name: "bosh", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
//		deleteAllInBase(dbCli)
//	})
//	// need to change handlers
//	Convey("Empty MAC", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: ""}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(dbCli)
//	})
//	Convey("Empty Type", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
//		deleteAllInBase(dbCli)
//	})
//	Convey("Empty Name", t, func() {
//		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(dbCli)
//	})
//	Convey("Empty Time", t, func() {
//		req := Request{Action: "update", Time: 0, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
//		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
//		deleteAllInBase(dbCli)
//	})
//
//}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestSendJSONToServer(t *testing.T) {
	defer treatmentPanic("Recovered in TestSendJSONToServer")

	var coonNotNil bool = false
	buffer := make([]byte, 1024)

	conn, _ := net.Dial("tcp", centerIP+":"+fmt.Sprint(tcpDevDataPort))
	if conn != nil {
		coonNotNil = true
		defer conn.Close()
	} else {
		log.Error("Conn is nil")
	}
	//Create redis client------------------------------------------------------------
	defer treatmentPanic("Recovered in TestCheckJSONToServer")
	var dbCli dao.DbDriver = &dao.RedisClient{DbServer:dbServer}
	dbCli.Connect()
	defer dbCli.Close()
	//--------------------------------------------------------------------------------

	res := "\"status\":200,\"descr\":\"Data has been delivered successfully\""
	req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge",
		Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
	message, _ := json.Marshal(req)

	if coonNotNil {
		conn.Write(message)
		time.Sleep(timeForSleep)

		for i := 0; i == 0; {
			i, _ = conn.Read(buffer)
		}
	}
	response := bytes.NewBuffer(buffer).String()

	if !strings.Contains(response, res) {
		t.Error("Bad JSON", response, res)
	}
	deleteAllInBase(dbCli)
}

func TestCheckJSONToServer(t *testing.T) {
	defer treatmentPanic("Recovered in TestCheckJSONToServer")
	var coonNotNil bool = false

	conn, _ := net.Dial("tcp", centerIP+":"+fmt.Sprint(tcpDevDataPort))
	if conn != nil {
		coonNotNil = true
		defer conn.Close()
	} else {
		log.Error("Conn is nil")
	}

	res := "\"status\":200,\"descr\":\"Data has been delivered successfully\""

	//Create redis client------------------------------------------------------------
	defer treatmentPanic("Recovered in TestCheckJSONToServer")
	var dbCli dao.DbDriver = &dao.RedisClient{DbServer:dbServer}
	dbCli.Connect()
	defer dbCli.Close()
	//--------------------------------------------------------------------------------

	Convey("Send Correct JSON to server", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		//Check on error
		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}

		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("Warning! Uncorrect JSON was sent to server", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("JSON was sent to server. Action of fridge should be update", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("Warning! JSON was sent to server with uncorrect action value", t, func() {
		req := Request{Action: "nil", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("JSON was sent to server. Action of washer should be update", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "washer", Name: "bosh0e31", MAC: "00-15-E9-2B-99-3B"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("Warning! JSON was sent to server with uncorrect type value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "nil", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("Warning! JSON was sent to server without MAC value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: ""}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("Warning! JSON was sent to server without type value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("Warning! JSON was sent to server without name value", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})

	Convey("Warning! JSON was sent to server without time value ", t, func() {
		req := Request{Action: "update", Time: 0, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		message, _ := json.Marshal(req)
		buffer := make([]byte, 1024)

		if coonNotNil {
			conn.Write(message)
			time.Sleep(timeForSleep)
			for i := 0; i == 0; {
				i, _ = conn.Read(buffer)
			}
		}
		response := bytes.NewBuffer(buffer).String()

		So(response, ShouldContainSubstring, res)
		deleteAllInBase(dbCli)
	})
}
func TestRedisConnection(t *testing.T) {
	client := redis.New()
	Convey("Check redis client connection"+dbServer.IP+":"+string(dbServer.Port)+". Should be without error ", t, func() {
		err := client.Connect(dbServer.IP, dbServer.Port)
		defer client.Close()
		So(err, ShouldBeNil)
	})
}
func TestHTTPConnection(t *testing.T) {
	var httpClient = &http.Client{}

	//Create redis client------------------------------------------------------------
	defer treatmentPanic("Recovered in TestCheckJSONToServer")
	var dbCli dao.DbDriver = &dao.RedisClient{DbServer:dbServer}
	dbCli.Connect()
	defer dbCli.Close()
	//--------------------------------------------------------------------------------

	Convey("Check http://"+centerIP+":"+fmt.Sprint(httpConnPort)+"/devices/{id}/data. Should be without error ", t, func() {
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "//devices/fridge:hladik0e31:00-15-E9-2B-99-3C/data")
		So(res, ShouldNotBeNil)
		deleteAllInBase(dbCli)
	})
	Convey("Check http://"+centerIP+":"+fmt.Sprint(httpConnPort)+"/devices. Should be without error ", t, func() {
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices")
		So(res, ShouldNotBeNil)
	})
}

func TestWorkingServerAfterSendingJSON(t *testing.T) {

	defer treatmentPanic("Recovered in TestWorkingServerAfterSendingJSON")
	conn, _ := net.Dial("tcp", centerIP+":"+fmt.Sprint(tcpDevDataPort))
	defer conn.Close()
	var httpClient = &http.Client{}

	defer treatmentPanic("Recovered in TestWorkingServerAfterSendingJSON")
	connForDAta, _ := net.Dial("tcp", centerIP+":"+fmt.Sprint(tcpDevDataPort))
	defer conn.Close()

	//Create redis client------------------------------------------------------------
	defer treatmentPanic("Recovered in TestWorkingServerAfterSendingJSON")
	var dbCli dao.DbDriver = &dao.RedisClient{DbServer:dbServer}
	dbCli.Connect()
	defer dbCli.Close()
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
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(dbCli)
	})
	Convey("Send JSON where action = wrongValue. Should not be return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"wrongValue\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName2\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"TempCam1\":[\"10:10.5\"]," +
			"\"TempCam2\":[\"1500:15.5\"]}}"

		mustNotHave := "testName2"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(dbCli)
	})
	Convey("Send JSON where type = wrongValue. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"wrongValue\",\"name\":\"testName3\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave := "testName3"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(dbCli)
	})

	Convey("Send JSON without name. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave := "TestMACFridge3"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(dbCli)
	})
	Convey("Send JSON without mac. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"config\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"fridge4\"" +
			",\"mac\":\"\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave := "fridge4"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(dbCli)
	})

	Convey("Send JSON with wrong data. Should not to return data about our fridge", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"fridge5\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"qwe\":qwe},\"tempCam2\":{\"" +
			"qwe\":qwe}}}"

		mustNotHave := "fridge5"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotBeNil)
		So(bodyString, ShouldNotContainSubstring, mustNotHave)
		deleteAllInBase(dbCli)
	})
	//	// my part
	Convey("Send correct JSON. Initialize turned on as false ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"turnedOn\":false"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices/type=fridge&name=testName1&mac=00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(dbCli)
	})
	Convey("Send correct JSON. Initialize CollectFreq as 0 ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"collectFreq\":0"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices/type=fridge&name=testName1&mac=00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(dbCli)
	})
	Convey("Send correct JSON. Initialize SendFreq as 0 ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"sendFreq\":0"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices/type=fridge&name=testName1&mac=00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(dbCli)
	})
	Convey("Send correct JSON. Initialize StreamOn as false ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"streamOn\":false"
		conn.Write([]byte(reqMessage))
		time.Sleep(timeForSleep)
		res, _ := httpClient.Get("http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices/type=fridge&name=testName1&mac=00-15-E9-2B-99-3C/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(dbCli)

	})
	Convey("Send correct JSON. Patch device data: turned on as true ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"turnedOn\":false"
		conn.Write([]byte(reqMessage))
		url := "http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices/type=fridge&name=testName1&mac=00-15-E9-2B-99-3C/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"turnedOn\":true}")))
		httpClient.Do(r)
		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(dbCli)
	})

	Convey("Send correct JSON. Patch device data: stream on as true ", t, func() {
		reqMessage := "{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave := "\"streamOn\":false"
		conn.Write([]byte(reqMessage))
		url := "http://" + centerIP + ":" + fmt.Sprint(httpConnPort) + "/devices/type=fridge&name=testName1&mac=00-15-E9-2B-99-3C/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"streamOn\":true}")))
		time.Sleep(timeForSleep)
		httpClient.Do(r)

		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldContainSubstring, mustHave)
		deleteAllInBase(dbCli)
	})
}

func TestWSConnection(t *testing.T) {
	defer treatmentPanic("Recovered in TestWSConnection")

	req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge",
		Name:                                                                       "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
	mustBe := "{\"action\":\"update\",\"time\":1496741392463499334,\"meta\":{\"type\":\"fridge\",\"name\":\"hladik0e31\",\"" +
		"mac\":\"00-15-E9-2B-99-3C\",\"ip\":\"\"},\"data\":null}"

	//Create redis client------------------------------------------------------------
	defer treatmentPanic("Recovered in TestWSConnection")
	var dbCli dao.DbDriver = &dao.RedisClient{DbServer:dbServer}
	dbCli.Connect()
	defer dbCli.Close()
	//--------------------------------------------------------------------------------

	Convey("Checking how to work ws connection. Should be true", t, func() {
		//Create Web Socket connection from the client side--------------------------------
		url := "ws://" + centerIP + ":" + fmt.Sprint(wsPort) + "/devices/00-15-E9-2B-99-3C"
		var dialer *websocket.Dialer
		conn, _, err := dialer.Dial(url, nil)
		if err!=nil{
			log.Error(err)
		}
		//---------------------------------------------------------------------------------

		defer treatmentPanic("Recovered in TestWSConnection")
		dao.PublishWS(req, "devWS", dbCli)

		_, message, _ := conn.ReadMessage()
		So(bytes.NewBuffer(message).String(), ShouldEqual, mustBe)
	})
}