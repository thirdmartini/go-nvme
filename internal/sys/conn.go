package sys

import (
	"errors"
	"net"
)

type Conn struct {
	net.Conn
}

func (c *Conn) Writev(_ [][]byte) error {
	return errors.New("not implemented")
}

func (c *Conn) Readv(_ [][]byte) error {
	return errors.New("not implemented")
}

func NewConn(c net.Conn) *Conn {

	sc := &Conn{
		Conn: c,
	}

	return sc
}
