package push

import "mnet/IBase"

type SessionEvent struct {
	Type int
	Sess IBase.ISession
	Data string
}

func NewSessionEvent(t int, sess IBase.ISession, data string) *SessionEvent {

	return &SessionEvent{
		Type: t,
		Sess: sess,
		Data: data,
	}
}
