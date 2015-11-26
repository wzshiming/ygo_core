package ygo_core

type AskCode struct {
	Uniq   uint
	Method uint
}

type CallCode struct {
	reply  interface{}
	method string
}

type SessionInterface interface {
	Call(method string, reply interface{}) error
	Ask(method int, uniq int)
}

type session struct {
	ask  chan AskCode
	call chan CallCode
	cb   func(CallCode)
}

func NewSession(cb func(CallCode)) SessionInterface {
	s := &session{
		ask:  make(chan AskCode, 16),
		call: make(chan CallCode, 16),
		cb:   cb,
	}
	return s
}

func (s *session) Ask(method int, uniq int) {

}

func (s *session) Call(method string, reply interface{}) error {
	return nil
}
