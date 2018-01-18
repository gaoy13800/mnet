package socket

import (
	"net"
	"log"
	"mnet/IBase"
	"fmt"
	"mnet/push"
	"strings"
	"mnet/work"
	"mnet/task"
)

type udpPeer struct {
	*wtnetBase
	udpListener *net.UDPConn
	vipTask task.EventQueue
	UdpManage *net.UDPConn
}

func (this *udpPeer) UdpStart(address string) IBase.IBaseUdp {
	// 必须要先声明defer，否则不能捕获到panic异常
	defer func() {
		if err := recover(); err != nil {
			log.Println("recover error", err)
		}
	}()

	listener, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 6002,
	})

	this.UdpManage = listener

	if err != nil {
		panic(err)
	}
	//udpAddr, err := net.ResolveUDPAddr("udp", address)
	//if err != nil {
	//	panic(err)
	//}
	//listener, err := net.ListenUDP("udp", udpAddr)
	//if err != nil {
	//	panic(err)
	//}
	this.udpListener = listener

	go this.UdpAcceptor(this.udpListener)

	return this
}

func (this *udpPeer) UdpAcceptor(socket *net.UDPConn) {

	for {
		data := make([]byte, 50)

		index, remoteAddr, err := socket.ReadFromUDP(data)

		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}

		msg := string(data[0:index])

		log.Println("收到信息：", remoteAddr, msg)

		if !strings.Contains(msg, "wt") && !strings.Contains(msg,"000") && !strings.Contains(msg,"111"){

			log.Println("未识别信息：",msg)

			continue
		}

		//VIP 通道

		if  len(msg) == 23{
			this.vipTask.PostData(msg)
			continue
		}

		this.PostData(push.NewUdpEvent(remoteAddr, msg, socket))
	}
}

func NewUdpPeer(queue task.EventQueue, ip int64, vipQueue task.EventQueue, manage *work.UdpManage) *udpPeer {
	this := &udpPeer{
		wtnetBase: NewWtnetBase(queue, ip),
		vipTask:vipQueue,
	}

	return this
}
