package IBase

import (
	"net"
	"mnet/event"
	"mnet/task"
)

type ISession interface {

	ID() int64

	TerminalId ()string

	SetTerminalId(string)

	FromIBase() IBaseNet

	Close()

	Send(string)

	RawSend(string)

	Work(queue task.EventQueue)

	Decode() (string, error)

	//GetTerminalType() event.GTYPE

	SetTerminalType(t event.GTYPE)

	SetClientType(isLock bool)

	IsLock() bool

	CheckConnect() bool

	StartPingDoor()

	SetContainerTicker()
	//Ping() error
}

type IStorage interface {
	Add(ISession)

	Remove(ISession)

	GetSessionById(int64) ISession

	Count() int
}

type ITerminals interface {
	Add(string, ISession)

	Remove(string)

	GetSessionByTerminalId(string) (ISession, error)

	IsExists(string) bool

	GetDeviceIds() []string
}

type IBaseNet interface {
	Start(string) IBaseNet

	Stop()

	IP() int64

	Acceptor()

	Handler(net.Conn)

	IStorage
}

type IBaseUdp interface {
	UdpStart(string) IBaseUdp

	UdpAcceptor(*net.UDPConn)
}