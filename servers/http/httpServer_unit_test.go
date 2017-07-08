package http

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/KharkivGophers/center-smart-house/models"
	//log "github.com/Sirupsen/logrus"
	"testing"
	"reflect"
	//"net/http"
	//"net/http/httptest"
)


func TestNewHTTPServer(t *testing.T) {
	Convey("Check HTTPServer constructor", t, func() {
		server := NewHTTPServer(Server{}, Server{}, RoutinesController{})
		server2 := NewHTTPServer(Server{}, Server{}, RoutinesController{})
		So(reflect.DeepEqual(server, server2), ShouldBeTrue)
	})
}

//func TestRun(t *testing.T) {
//	Convey("Should be panic", t, func() {
//		server := NewHTTPServer(Server{}, Server{}, RoutinesController{})
//		server.Run()
//		So(server, ShouldBeError)
//	})
//}

//func TestHandler(t *testing.T) {
//	server := NewHTTPServer(Server{"0.0.0.0", uint(8100)}, Server{"127.0.0.1", uint(6379)},
//		RoutinesController{make(chan struct{})})
//
//	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
//	// pass 'nil' as the third parameter.
//	req, err := http.NewRequest(http.MethodGet, "/devices", nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
//	rr := httptest.NewRecorder()
//	handler := http.HandlerFunc(server.getDevicesHandler)
//
//	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
//	// directly and pass in our Request and ResponseRecorder.
//	handler.ServeHTTP(rr, req)
//
//	// Check the status code is what we expect.
//	if status := rr.Code; status != http.StatusOK {
//		t.Errorf("handler returned wrong status code: got %v want %v",
//			status, http.StatusOK)
//	}
//
//	// Check the response body is what we expect.
//	expected := `{"alive": true}`
//	if rr.Body.String() != expected {
//		t.Errorf("handler returned unexpected body: got %v want %v",
//			rr.Body.String(), expected)
//	}
//}
