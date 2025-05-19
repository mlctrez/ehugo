package ssdp

import (
	"net"
	"strings"
)

const (
	DefaultNetwork = "udp4"
	DefaultAddress = "239.255.255.250:1900"
	maxBufferSize  = 2048
)

type SSDP struct {
	network       string
	address       string
	addr          *net.UDPAddr
	interfaceName string
	conn          *net.UDPConn
	callback      PacketCallback
}

func New(opts ...Option) *SSDP {
	s := &SSDP{network: DefaultNetwork, address: DefaultAddress}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *SSDP) Listen() error {
	var err error
	s.addr, err = net.ResolveUDPAddr(s.network, s.address)
	if err != nil {
		return err
	}
	var ifi *net.Interface
	if s.interfaceName != "" {
		var interfaces []net.Interface
		interfaces, err = net.Interfaces()
		if err != nil {
			return err
		}
		for _, i := range interfaces {
			if i.Name == s.interfaceName {
				ifi = &i
				break
			}
		}
	}

	s.conn, err = net.ListenMulticastUDP(s.network, ifi, s.addr)
	return err
}

func (s *SSDP) Read() {
	buffer := make([]byte, maxBufferSize)
	for {
		n, client, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				s.conn = nil
				return
			}
			continue
		}
		if s.callback != nil {
			s.callback(&Packet{Client: client, Data: buffer[:n]})
		}
	}
}

func (s *SSDP) Shutdown() {
	if s.conn != nil {
		_ = s.conn.Close()
	}
}

type Option func(*SSDP)

func WithNetwork(network string) Option {
	return func(s *SSDP) {
		s.network = network
	}
}

func WithCallback(callback PacketCallback) Option {
	return func(s *SSDP) {
		s.callback = callback
	}
}

func WithAddress(address string) Option {
	return func(s *SSDP) {
		s.address = address
	}
}

func WithInterface(name string) Option {
	return func(s *SSDP) {
		s.interfaceName = name
	}
}
