package tcp

import (
	"sync"
	"net"
)

type ConnectionPool struct {
	sync.Mutex
	conn map[string]net.Conn
}

func (pool *ConnectionPool) AddConn(conn net.Conn, key string) {
	pool.Lock()
	pool.conn[key] = conn
	defer pool.Unlock()
}

func (pool *ConnectionPool) GetConn(key string) net.Conn {
	pool.Lock()
	defer pool.Unlock()
	return pool.conn[key]
}

func (pool *ConnectionPool) RemoveConn(key string)  {
	pool.Lock()
	defer pool.Unlock()
	delete(pool.conn, key)
}

func (pool *ConnectionPool) Init() {
	pool.Lock()
	defer pool.Unlock()

	pool.conn = make(map[string]net.Conn)
}
