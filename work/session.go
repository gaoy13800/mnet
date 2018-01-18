package work

import (
	"mnet/IBase"
	"mnet/task"
	"mnet/event"
	"mnet/push"
	"net"
	"sync"
	"time"
	"fmt"
	"log"
)

type SocketSession struct {
	WithClose func()

	id int64

	terminalId string

	terminalT event.GTYPE

	isLock	bool

	conn net.Conn

	wait sync.WaitGroup

	base IBase.IBaseNet

	tickerContainer int

	ConnectStatus	string

	tasks 			task.EventQueue
}

func (this *SocketSession) ID() int64 {
	return this.id
}

func (this *SocketSession) TerminalId() string {
	return this.terminalId
}

func (this *SocketSession) SetTerminalId(terminalId string) {
	this.terminalId = terminalId
}

func (this *SocketSession) FromIBase() IBase.IBaseNet {
	return this.base
}

func (this *SocketSession) Close() {
	this.conn.Close() //可关闭多次
	this.WithClose()
}

func (this *SocketSession) RawSend(data string) {
	if data == "" {
		return
	}

	_, err := this.conn.Write([]byte(data))

	if err != nil {
		log.Print("消息发送失败！")
	}else {
		log.Print("消息发送成功！")
	}
}

func (this *SocketSession) Send(data string) {

}

func (this *SocketSession) Work(queue task.EventQueue) {

	this.wait.Add(1)

	go func() {

		this.wait.Wait()

		this.Close()
	}()

	go this.pipelineRevc(queue)
}

func (this *SocketSession) pipelineRevc(queue task.EventQueue) {

	this.tasks = queue

	for {

		data, err := this.Decode()

		if err != nil {
			//log.Println("close event", err)
			goto CLOSELABLE
		}

		queue.PostData(push.NewSessionEvent(event.Msg, this, data))

		//this.tickerContainer = event.TICKER

		continue

	CLOSELABLE:
		queue.PostData(push.NewSessionEvent(event.Close, this, ""))
		break
	}

	this.wait.Done()
}

func (this *SocketSession) StartPingDoor(){
	this.tickerContainer = event.TICKER

	go this.pingSocket()
}

func (this *SocketSession) SetContainerTicker(){

	this.tickerContainer = event.TICKER
}

func (this *SocketSession) pingSocket() {
	ticker := time.NewTicker(20 * time.Second)

	for {
		select {
		case <-ticker.C:
			if this.tickerContainer < 0 {
				goto TICKEREND
			}

			this.tickerContainer = this.tickerContainer - 20
		}
	}
TICKEREND:

	this.tasks.PostData(push.NewSessionEvent(event.Msg, this, "heartUnConnect"))

	//只需要发送一次关闭连接的信号即可，这里的close会使 pipe发送关闭信号
	fmt.Println("门禁心跳超时")

	//this.Close()
}

func (this *SocketSession) Decode() (string, error) {
	byt := make([]byte, 1024)
	index, err := this.conn.Read(byt)

	if err != nil {
		return "", err
	}

	return string(byt[0:index]), err
}


func (this *SocketSession) SetTerminalType(t event.GTYPE) {
	this.terminalT = t
}

func (this *SocketSession) SetClientType(isLock bool){
	this.isLock = isLock
}

func (this *SocketSession) IsLock() bool{
	return this.isLock
}

func (this *SocketSession) CheckConnect() bool {
	if this.id == 0 || this.terminalId == ""{
		return false
	}

	return true
}

func NewSession(conn net.Conn, base IBase.IBaseNet) *SocketSession {

	this := &SocketSession{
		conn: conn,
		base: base,
		ConnectStatus:"connect",
	}

	return this
}
