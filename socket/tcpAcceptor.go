package socket

import (
	"log"
	"mnet/IBase"
	"mnet/event"
	"mnet/push"
	"mnet/work"
	"net"
	"sync"
	"time"
	"mnet/task"
)

type socketPeer struct {
	*wtnetBase

	*work.SessionManager

	listener net.Listener

	running bool

	syncTex sync.RWMutex
}

func (this *socketPeer) Start(address string) IBase.IBaseNet {

	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Println("recover error", err)
		}
	}()

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)

	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		panic(err)
	}

	this.listener = listener

	go this.Acceptor()

	return this
}

func (this *socketPeer) Stop() {
	this.listener.Close()

	this.setRunning(false)
}

func (this *socketPeer) Acceptor() {

	this.setRunning(true)

	for {

		if !this.isRunning() {
			break
		}

		conn, err := this.listener.Accept()

		if err != nil {
			panic(err)
		}

		go this.Handler(conn)
	}

	this.setRunning(false)

	//sigal close all
}

func (this *socketPeer) Handler(conn net.Conn) {

	sess := work.NewSession(conn, this)

	this.Add(sess)

	sess.WithClose = func() {
		this.Remove(sess)
	}

	sess.Work(this)

	data := push.NewSessionEvent(event.Connect, sess, "")

	this.PostData(data)
}

func (this *socketPeer) setRunning(isRun bool) {
	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	this.running = isRun
}

func (this *socketPeer) isRunning() bool {
	return this.running
}

func NewPeer(queue task.EventQueue, ip int64) *socketPeer {
	this := &socketPeer{
		SessionManager: work.NewSessionManager(),
		wtnetBase:      NewWtnetBase(queue, ip),
	}

	go func() {
		tick := time.Tick(time.Second * 15)

		for {
			select {
			case <-tick:
				log.Println("current session number:", this.SessionManager.Count())
			}
		}
	}()

	return this
}
