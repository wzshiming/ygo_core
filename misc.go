package ygo_core

import (
	"time"

	"github.com/wzshiming/base"
)

func nap(i int) {
	time.Sleep(time.Second * time.Duration(i) / 10)
}

func Debug(v ...interface{}) {
	base.DEBUG(v...)
}

func DebugStack() {
	if x := recover(); x != nil {
		if _, ok := x.(string); ok {

		} else {
			base.DebugStack()
		}
	}
}

func RandInt(i int) int {
	return int(<-base.LCG) % i
}

type Action func(ca *Card) bool

func (ac Action) IsExits() bool {
	if ac != nil {
		return true
	}
	return false
}

func (ac Action) Call(ca *Card) bool {
	if ac != nil {
		return ac(ca)
	}
	return false
}

type Arg map[string]interface{}

func eventParameter(i []interface{}) (ss []string, e interface{}) {
	l := len(i) - 1
	for k, v := range i {
		if k == l {
			e = v
		} else if s, ok := v.(string); ok {
			ss = append(ss, s)
		} else {
			base.DEBUG("eventParameter", i)
		}
	}
	return
}
