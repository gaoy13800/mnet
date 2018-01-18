package work

import (
	"errors"
	"mnet/IBase"
	"sync"
)

type terminalManager struct {
	terminalList map[string]IBase.ISession

	syncTex sync.RWMutex
}

func (this *terminalManager) Add(terminalId string, sess IBase.ISession) {

	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	this.terminalList[terminalId] = sess
}

func (this *terminalManager) Remove(terminalId string) {
	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	delete(this.terminalList, terminalId)
}

func (this *terminalManager) GetSessionByTerminalId(terminalId string) (IBase.ISession, error) {
	if v, ok := this.terminalList[terminalId]; ok {
		return v, nil
	}

	return nil, errors.New("not exits")
}

func (this *terminalManager) IsExists(terminalId string) bool {

	if _, ok := this.terminalList[terminalId]; !ok {
		return false
	}

	return true
}

func (this *terminalManager) GetDeviceIds()[]string{
	list := make([]string, 0, 1000)

	for key, _ := range this.terminalList{
		list = append(list, key)
	}

	return list
}

func NewTerminalManage() *terminalManager {

	return &terminalManager{
		terminalList: make(map[string]IBase.ISession),
	}

}
