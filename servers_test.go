package main

import (
	"testing"
	"net"
	. "github.com/smartystreets/goconvey/convey"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestTCPConnection(t *testing.T) {

	conn, _ := net.Dial("tcp", connHost+":"+tcpConnPort)

	Convey("TCP connection should be without error", t, func() {
		So(conn, ShouldNotBeNil)
	})
}
