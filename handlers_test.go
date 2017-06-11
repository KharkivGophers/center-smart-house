package main

import (
	"testing"
	"net"
	. "github.com/smartystreets/goconvey/convey"
	"encoding/json"
	"bytes"
	"strings"
	"net/http"
	"io/ioutil"
)

func TestDevTypeHandler(t *testing.T) {
	Convey("Message handles correct", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Empty message", t, func() {
		req := Request{}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown action")
	})
	Convey("Type washer message", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "washer", Name: "bosh", MAC: "00-15-E9-2B-99-3C"}}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Unknown type message", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "nil", Name: "bosh", MAC: "00-15-E9-2B-99-3C"}}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
	})
	// need to change handlers
	Convey("Empty MAC", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: ""}}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Empty Type", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request: unknown device type")
	})
	Convey("Empty Name", t, func() {
		req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "", MAC: "00-15-E9-2B-99-3C"}}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})
	Convey("Empty Time", t, func() {
		req := Request{Action: "update", Time: 0, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
		So(devTypeHandler(req), ShouldContainSubstring, "Device request correct")
	})

}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestSendJSONToServer(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()

	res := "\"status\":200,\"descr\":\"Data have been delivered successfully\""
	req := Request{Action: "update", Time: 1496741392463499334, Meta: DevMeta{Type: "fridge", Name: "hladik0e31", MAC: "00-15-E9-2B-99-3C"}}
	message, _ := json.Marshal(req)
	conn.Write(message)

	buffer := make([]byte, 1024)

	for i := 0; i == 0; {
		i, _ = conn.Read(buffer)
	}
	response := bytes.NewBuffer(buffer).String()

	if !strings.Contains(response, res) {
		t.Error("Bad JSON", response, res)
	}
}

func TestCheckJSONToServer(t *testing.T) {
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()

	res := "\"status\":200,\"descr\":\"Data have been delivered successfully\""

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

func TestWorkingServerAfterSendingJSON(t *testing.T){
	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)
	defer conn.Close()
	var httpClient = &http.Client{}

	Convey("Check http://"+connHost+":"+httpConnPort+"/devices. Should be without error ", t, func() {
		res, err := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices")
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	})


	Convey("Send correct JSON. Should be return all ok ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="{\"site\":\"\",\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\",\"mac\":\"Test1\"," +
			"\"ip\":\"\"},\"data\":{\"TempCam1\":[\"10:10.5\"],\"TempCam2\":[\"1500:15.5\"]}}"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldContainSubstring,mustHave)
	})
	Convey("Send JSON where action = wrongValue. Should not be return data about our fridge", t, func() {
		reqMessage :="{\"action\":\"wrongValue\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName2	\"" +
			",\"mac\":\"Test2\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave :="testName2"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotBeNil)
		So(bodyString, ShouldNotContainSubstring,mustNotHave)
	})
	Convey("Send JSON where type = wrongValue. Should not to return data about our fridge", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"wrongValue\",\"name\":\"testName3\"" +
			",\"mac\":\"Test3\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave :="testName3"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotBeNil)
		So(bodyString, ShouldNotContainSubstring,mustNotHave)
	})

	Convey("Send JSON without name. Should not to return data about our fridge", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"\"" +
			",\"mac\":\"TestMACFridge3\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave :="TestMACFridge3"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotBeNil)
		So(bodyString, ShouldNotContainSubstring,mustNotHave)
	})
	Convey("Send JSON without mac. Should not to return data about our fridge", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"fridge4\"" +
			",\"mac\":\"\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustNotHave :="fridge4"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotBeNil)
		So(bodyString, ShouldNotContainSubstring,mustNotHave)
	})

	Convey("Send JSON with wrong data. Should not to return data about our fridge", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"fridge5\"" +
			",\"mac\":\"test5\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"qwe\":qwe},\"tempCam2\":{\"" +
			"qwe\":qwe}}}"

		mustNotHave :="fridge5"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices")
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)

		So(bodyString, ShouldNotBeNil)
		So(bodyString, ShouldNotContainSubstring,mustNotHave)
	})
	// my part
	Convey("Send correct JSON. Initialize turned on as false ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"turnedOn\":false"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Initialize CollectFreq as 0 ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"collectFreq\":0"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Initialize SendFreq as 0 ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"sendFreq\":0"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Initialize StreamOn as false ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"streamOn\":false"
		conn.Write([]byte(reqMessage))

		res, _ := httpClient.Get("http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config")

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Patch device data: turned on as true ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"turnedOn\":true"
		conn.Write([]byte(reqMessage))
		url := "http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"turnedOn\":true}")))
		httpClient.Do(r)
		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		r, _ = http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"turnedOn\":false}")))
		httpClient.Do(r)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Patch device data: CollectFreq as 5 ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"collectFreq\":5"
		conn.Write([]byte(reqMessage))
		url := "http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"collectFreq\":5}")))
		httpClient.Do(r)
		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		r, _ = http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"collectFreq\":0}")))
		httpClient.Do(r)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Patch device data: CollectFreq as 5 ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"collectFreq\":5"
		conn.Write([]byte(reqMessage))
		url := "http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"collectFreq\":5}")))
		httpClient.Do(r)
		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		r, _ = http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"collectFreq\":0}")))
		httpClient.Do(r)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Patch device data: SendFreq as 15 ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"sendFreq\":15"
		conn.Write([]byte(reqMessage))
		url := "http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"sendFreq\":15}")))
		httpClient.Do(r)
		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		r, _ = http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"sendFreq\":0}")))
		httpClient.Do(r)
		So(bodyString, ShouldContainSubstring, mustHave)
	})
	Convey("Send correct JSON. Patch device data: stream on as true ", t, func() {
		reqMessage :="{\"action\":\"update\",\"time\":20,\"meta\":{\"type\":\"fridge\",\"name\":\"testName1\"" +
			",\"mac\":\"Test1\",\"ip\":\"\"},\"data\":{\"tempCam1\":{\"10\":10.5},\"tempCam2\":{\"" +
			"1500\":15.5}}}"

		mustHave :="\"streamOn\":true"
		conn.Write([]byte(reqMessage))
		url := "http://"+connHost+":"+httpConnPort+"/devices/fridge:testName1:Test1/config"
		r, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"streamOn\":true}")))
		httpClient.Do(r)
		res, _ := httpClient.Get(url)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		r, _ = http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("{\"streamOn\":false}")))
		httpClient.Do(r)
		So(bodyString, ShouldContainSubstring, mustHave)
	})

}