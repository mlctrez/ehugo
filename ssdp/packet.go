package ssdp

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/textproto"
	"strings"
)

type PacketCallback func(p *Packet)
type Packet struct {
	Client     *net.UDPAddr
	Data       []byte
	Method     string
	RequestURI string
	Proto      string
	MIMEHeader textproto.MIMEHeader
}

type BridgeInfo struct {
	Location     string
	SerialNumber string
	UUID         string
}

func (p *Packet) Parse() (err error) {
	tp := textproto.NewReader(bufio.NewReader(bytes.NewBuffer(p.Data)))
	var headerLine string
	if headerLine, err = tp.ReadLine(); err != nil {
		return err
	}

	method, rest, ok1 := strings.Cut(headerLine, " ")
	requestURI, proto, ok2 := strings.Cut(rest, " ")
	if !ok1 || !ok2 {
		return fmt.Errorf("invalid header: %s", headerLine)
	}
	p.Method = method
	p.RequestURI = requestURI
	p.Proto = proto
	p.MIMEHeader, err = tp.ReadMIMEHeader()
	return err
}

func (p *Packet) Reply(bridgeInfo *BridgeInfo) error {
	conn, err := net.DialUDP("udp", nil, p.Client)
	if err != nil {
		return err
	}
	defer func(conn *net.UDPConn) { _ = conn.Close() }(conn)

	var serviceName = "urn:schemas-upnp-org:device:basic:1"

	if p.MIMEHeader.Get("St") == "upnp:rootdevice" {
		serviceName = fmt.Sprintf("uuid:%s::upnp:rootdevice", bridgeInfo.UUID)
	}

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"CACHE-CONTROL: max-age=60\r\n"+
		"EXT:\r\n"+
		"LOCATION: %s\r\n"+
		"SERVER: FreeRTOS/6.0.5, UPnP/1.0, IpBridge/1.16.0\r\n"+
		"hue-bridgeid: %s\r\n"+
		"ST: %s\r\n"+
		"USN: %s\r\n"+
		"\r\n",
		bridgeInfo.Location,
		bridgeInfo.SerialNumber,
		p.MIMEHeader.Get("St"),
		serviceName,
	)

	_, err = conn.Write([]byte(response))
	return err
}
