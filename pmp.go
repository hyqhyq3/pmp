// a pure go implement for NAT-PMP
package pmp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
	"unsafe"
)

type Gateway string

// g is the router's IP
func NewGateway(g string) Gateway {
	return Gateway(g)
}

type ExtIpResponse struct {
	ver       byte
	op        byte
	retCode   [2]byte
	timestamp uint32
	ip        [4]byte
}

func (r *ExtIpResponse) String() (str string) {
	str += fmt.Sprintln("ver", r.ver)
	str += fmt.Sprintln("op", r.op)
	str += fmt.Sprintln("retCode", binary.BigEndian.Uint16(r.retCode[:]))
	str += fmt.Sprintln("timestamp", r.timestamp)
	str += fmt.Sprintln("extIp", net.IP(r.ip[:]))
	return
}

// get the external IP
func (g *Gateway) ExtIP() (net.IP, error) {
	c, _ := net.Dial("udp", string(*g)+":5351")
	c.Write([]byte{0, 0})
	var resp ExtIpResponse
	b := (*(*[1 << 30]byte)(unsafe.Pointer(&resp)))[:unsafe.Sizeof(resp)]
	_, err := io.ReadFull(c, b)
	if err != nil {
		return nil, err
	}
	retCode := binary.BigEndian.Uint16(resp.retCode[:])
	if retCode != 0 {
		return nil, errors.New(fmt.Sprintf("code = %d", retCode))
	}
	return net.IP(resp.ip[:]), nil
}

type Protocol uint8

const (
	UDP Protocol = iota + 1
	TCP
)

type PortMapReq struct {
	ver          byte
	op           byte
	reserve      uint16
	internalPort uint16
	externalPort uint16
	lifetime     uint32
}

type PortMapResp struct {
	ver          byte
	op           byte
	retCode      [2]byte
	timestamp    [4]byte
	internalPort [2]byte
	externalPort [2]byte
	lifetime     [4]byte
}

func htons(v uint16) (s uint16) {
	p := (*[2]byte)(unsafe.Pointer(&s))
	p[0] = byte(v >> 8)
	p[1] = byte(v)
	return
}

func htonl(v uint32) (s uint32) {
	p := (*[4]byte)(unsafe.Pointer(&s))
	p[0] = byte(v >> 24)
	p[1] = byte(v >> 16)
	p[2] = byte(v >> 8)
	p[3] = byte(v)
	return
}

func (g *Gateway) AddPortMapping(proto Protocol, internalPort, externalPort, lifetime int) (int, error) {
	var req PortMapReq
	req.op = byte(proto)
	req.internalPort = htons(uint16(internalPort))
	req.externalPort = htons(uint16(externalPort))
	req.lifetime = htonl(uint32(lifetime))

	b := (*(*[1 << 30]byte)(unsafe.Pointer(&req)))[:unsafe.Sizeof(req)]
	c, err := net.Dial("udp", string(*g)+":5351")
	if err != nil {
		return 0, err
	}
	c.SetDeadline(time.Now().Add(time.Second))
	_, err = c.Write(b)
	if err != nil {
		return 0, err
	}
	var resp PortMapResp
	b = (*(*[1 << 30]byte)(unsafe.Pointer(&resp)))[:unsafe.Sizeof(resp)]
	_, err = io.ReadFull(c, b)
	if err != nil {
		return 0, err
	}
	retCode := binary.BigEndian.Uint16(resp.retCode[:])
	port := binary.BigEndian.Uint16(resp.externalPort[:])
	if retCode != 0 {
		return 0, errors.New(fmt.Sprintf("errorCode %d", retCode))
	}
	return int(port), nil
}
