package push

import (
	"net"
)

type UdpEvent struct {
	RemoteAddr  *net.UDPAddr
	Msg string

	Socket *net.UDPConn
}

func NewUdpEvent(addr *net.UDPAddr, msg string,socket *net.UDPConn) *UdpEvent {
	return &UdpEvent{
		RemoteAddr: addr,
		Msg: msg,
		Socket:socket,
	}
}
