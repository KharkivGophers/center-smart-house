package http

import (
	. "github.com/smartystreets/goconvey/convey"
	//log "github.com/Sirupsen/logrus"
	"testing"
)


func NewHTTPServerTest(t *testing.T) {
	Convey("Check HTTPServer constructor", t, func() {
		NewHTTPServer(Server{}, Server{}, nil)
	})
}
