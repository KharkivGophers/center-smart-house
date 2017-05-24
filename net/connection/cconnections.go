package connection

import (
	"net"
)

type Connection struct {
	conn *net.Conn
	MAC  string
	idle bool
}

type ConnsPool struct {
	connsChan   map[string]*net.Conn
	connsNumber int
}
