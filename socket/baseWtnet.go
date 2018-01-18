package socket

import (
	"mnet/task"
)

type wtnetBase struct {
	ip   int64
	name string
	task.EventQueue
}

func NewWtnetBase(evq task.EventQueue, ip int64) *wtnetBase {

	self := &wtnetBase{
		EventQueue: evq,
		ip:         ip,
	}
	return self
}

func (self *wtnetBase) IP() int64 {
	return self.ip
}
