package transport

import (
	"bufio"
	"net"
	"strconv"
	"sync"

	"paficent/GoFox2X/protocol"
)

type Conn struct {
	netConn net.Conn
	reader  *bufio.Reader

	writeMu sync.Mutex

	Host string
	Port int

	Session any
}

func newConn(nc net.Conn) *Conn {
	host, portStr, err := net.SplitHostPort(nc.RemoteAddr().String())
	port := 0
	if err == nil {
		port, _ = strconv.Atoi(portStr)
	}
	return &Conn{
		netConn: nc,
		reader:  bufio.NewReader(nc),
		Host:    host,
		Port:    port,
	}
}

func (c *Conn) Receive() (*protocol.Message, error) {
	body, err := protocol.ReadFrame(c.reader)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeMessage(body)
}

func (c *Conn) Send(m *protocol.Message) error {
	body, err := m.MarshalBinary()
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return protocol.WriteFrame(c.netConn, body)
}

func (c *Conn) Close() error {
	return c.netConn.Close()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.netConn.RemoteAddr()
}
