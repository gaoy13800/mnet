package work

import "net"

type UdpManage struct {
	udpSocket *net.UDPConn
}

func (this *UdpManage) SOCKET_GET () *net.UDPConn{
	return this.udpSocket
}

func (this *UdpManage) SOCKET_SET(conn *net.UDPConn){
	this.udpSocket = conn
}

func NewUdpManage() *UdpManage{

	return &UdpManage{
	}
}