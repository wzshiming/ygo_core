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
