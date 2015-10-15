package pmp

import (
	"net"
	"testing"
)

func TestExtIp(t *testing.T) {
	g := NewGateway("192.168.1.1")
	_, err := g.ExtIP()
	if err != nil {
		t.Fail()
	}
}

func TestPortMapping(t *testing.T) {

	lsn, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fail()
	}

	g := NewGateway("192.168.1.1")
	port, err := g.AddPortMapping(TCP, lsn.Addr().(*net.TCPAddr).Port, 0, 100)
	if err != nil {
		t.Fail()
	}

	go func() {
		c, _ := lsn.Accept()
		c.Write([]byte("hello"))
		c.Close()
	}()

	ip, err := g.ExtIP()
	if err != nil {
		t.Fail()
	}

	c, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		t.Fail()
	}
	b := make([]byte, 100)
	n, err := c.Read(b)
	if err != nil {
		t.Fail()
	}
	if string(b[:n]) != "hello" {
		t.Fail()
	}
}
