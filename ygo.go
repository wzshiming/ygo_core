package ygo_core

import (
	"service/proto"
	"time"

	"github.com/wzshiming/base"
	"github.com/wzshiming/dispatcher"
	"github.com/wzshiming/server/agent"
)

type YGO struct {
	dispatcher.Events
	CardVer  *CardVersion
	Room     *agent.Room
	StartAt  time.Time
	Cards    map[uint]*Card
	Players  map[uint]*Player
	Current  *Player
	Survival map[int]int
	Over     bool
	round    []uint

	pending   map[uint]*Card
	cardevent map[string]map[*Card]Action

	both  map[string]bool
	multi map[string]bool
}

func NewYGO(r *agent.Room) *YGO {
	yg := &YGO{
		Events:   dispatcher.NewLineEvent(),
		Room:     r,
		Cards:    map[uint]*Card{},
		Survival: map[int]int{},
		StartAt:  time.Now(),
		Players:  map[uint]*Player{},
		both:     map[string]bool{},
		multi:    map[string]bool{},
	}

	return yg
}

//func (yg *YGO) Dispatch(eventName string, args ...interface{}) {
//	rego.ERR(eventName, args)
//}

func (yg *YGO) RegisterBothEvent(eventName string) {
	yg.both[eventName] = true
}

func (yg *YGO) RegisterMultiEvent(eventName string) {
	yg.multi[eventName] = true
}

func (yg *YGO) Chain(eventName string, ca *Card, pl *Player, args []interface{}) {

	if ca != nil {
		flag := true
		for _, v := range args {
			if _, ok := v.(*Card); ok {
				flag = false
				break
			}
		}
		if flag {
			args = append(args, ca)
		}
	}

	if pl != nil {
		flag := true
		for _, v := range args {
			if _, ok := v.(*Player); ok {
				flag = false
				break
			}
		}
		if flag {
			args = append(args, pl)
		}
	}

	yg.EmptyEvent(Chain)
	yg.Dispatch(eventName, args...)

	cs := NewCards()
	yg.ForEventEach(Chain, func(i interface{}) {
		if v, ok := i.(*Card); ok {
			cs.EndPush(v)
		}
	})
	if cs.Len() > 0 || yg.both[eventName] {
		pl.Chain(eventName, ca, cs, args)
		if ca != nil && ca.IsValid() {
			pl.GetTarget().Chain(eventName, ca, cs, args)
		}
	}
	yg.EmptyEvent(Chain)

}

func (yg *YGO) GetPlayer(sess *agent.Session) *Player {
	return yg.Players[sess.ToUint()]
}

func (yg *YGO) GetCard(u uint) (c *Card) {
	c = yg.Cards[u]
	return
}

func (yg *YGO) RegisterCards(c *Card) {
	yg.Cards[c.ToUint()] = c
}

func (yg *YGO) ForEachPlayer(fun func(*Player)) {
	for _, v := range yg.round {
		fun(yg.Players[v])
	}
}

func (yg *YGO) CallAll(method string, reply interface{}) error {
	yg.Room.Broadcast(Call{
		Method: method,
		Args:   reply,
	})
	return nil
}

func (yg *YGO) GetPlayerForIndex(i int) *Player {
	return yg.Players[yg.round[i]]
}

func (yg *YGO) Loop() {
	defer func() {
		if x := recover(); x != nil {
			base.DebugStack()
		}
	}()
	yg.Room.ForEach(func(sess *agent.Session) {
		p := NewPlayer(yg)
		p.Session = sess
		yg.Players[sess.ToUint()] = p
	})

	nap(1) // 服务端初始化
	for k, _ := range yg.Players {
		yg.round = append(yg.round, k)
	}

	for k, v := range yg.round {
		ca := yg.Players[v].Camp
		yg.Survival[ca] = yg.Survival[ca] + 1
		yg.Players[v].Index = k
		yg.Players[v].game = yg
		yg.Players[v].RoundSize = 0

		var id uint
		yg.Players[v].Session.Data.Get("userid", &id)
		if id == 0 {
			yg.Players[v].Name = "Guest"
		} else {
			var name string
			yg.Players[v].Session.Data.Get("username", &name)
			yg.Players[v].Name = name
		}

	}

	nap(1) // 客户端初始化
	gi := proto.GameInitResponse{}
	for _, v := range yg.round {
		pi := proto.PlayerInit{
			//Hp:   yg.Players[v].Hp,
			Name: yg.Players[v].Name,
		}
		gi.Users = append(gi.Users, pi)
	}
	for _, v := range yg.round {
		gi.Index = yg.Players[v].Index
		yg.Players[v].Call("init", gi)
	}

	nap(10) // 界面初始化
	i := 31
	for _, v := range yg.round {
		i++
		yg.Players[v].InitPlayer(i)
	}

	nap(10) // 牌组初始化
	for _, v := range yg.round {
		var deck proto.Deck
		yg.Players[v].Session.Data.Get("deck", &deck)

		yg.Players[v].initDeck(deck.GetMain())

	}

	nap(20) // 手牌初始化
	for _, v := range yg.round {
		yg.Players[v].init()
		yg.Players[v].ChangeHp(4000)
	}

	//必要连锁初始化
	yg.RegisterBothEvent(Summon)
	yg.RegisterBothEvent(SummonFlip)
	yg.RegisterBothEvent(SummonSpecial)
	yg.RegisterBothEvent(Declaration)
	yg.RegisterBothEvent(UseTrap)
	yg.RegisterBothEvent(UseMagic)
	yg.RegisterMultiEvent(DP)
	yg.RegisterMultiEvent(SP)
	yg.RegisterMultiEvent(MP)
	yg.RegisterMultiEvent(EP)

	nap(10) // 游戏开始

	pl := yg.GetPlayerForIndex(0)
	pl.MsgPub("msg.001", nil)
	if pl.Portrait.Len() == 1 {
		ca := pl.Portrait.Get(0)
		ca.RegisterGlobalListen(BP, func(tar *Player) {
			tar.Mzone.ForEach(func(c *Card) bool {
				c.SetNotCanAttack()
				return true
			})
		})
		ca.RegisterGlobalListen(RoundEnd, func() {
			ca.UnregisterGlobalListen()
		})
	}

	yg.Room.LeaveEvent(func(sess *agent.Session) {
		pl := yg.Players[sess.ToUint()]
		yg.Current.MsgPub("msg.009", Arg{"rival": pl.Name})
		pl.Fail()
	})
loop:
	for {
		for _, v := range yg.round {
			nap(5)
			yg.Current = yg.Players[v]
			yg.Current.round()
			yg.ForEachPlayer(func(pl *Player) {
				if pl.IsFail() {
					yg.Over = true
				}
			})
			if yg.Over {
				break loop
			}
		}
	}

	yg.GameOver()
	return
}

func (yg *YGO) GameOver() {
	yg.CallAll("over", nil)
	yg.Current.MsgPub("msg.000", nil)
}
