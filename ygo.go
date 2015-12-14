package ygo_core

import (
	"time"

	"github.com/wzshiming/base"
	"github.com/wzshiming/dispatcher"
	"github.com/wzshiming/server/agent"
)

type playerInit struct {
	//Hp   int    `json:"hp"`
	Name string `json:"name"`
}

type gameInitResponse struct {
	// user 表示当前游戏中 卡片的id
	Index int          `json:"index"` //自身的索引
	Users []playerInit `json:"users"`
}

type YGO struct {
	dispatcher.Events
	Room      *agent.Room
	cardVer   *CardVersion
	StartAt   time.Time
	cards     map[uint]*Card
	players   map[uint]*Player
	sesstions map[uint]*agent.Session
	current   *Player
	survival  map[int]int
	Over      bool
	round     []uint

	pending   map[uint]*Card
	cardevent map[string]map[*Card]Action

	both  map[string]bool
	multi map[string]bool
	quick map[*Card]bool
}

func NewYGO(r *agent.Room) *YGO {
	yg := &YGO{
		Events:    dispatcher.NewLineEvent(),
		Room:      r,
		cards:     map[uint]*Card{},
		survival:  map[int]int{},
		StartAt:   time.Now(),
		players:   map[uint]*Player{},
		both:      map[string]bool{},
		multi:     map[string]bool{},
		quick:     map[*Card]bool{},
		sesstions: map[uint]*agent.Session{},
	}
	yg.Room.ForEach(func(sess *agent.Session) {
		p := newPlayer(yg)
		p.sessUniq = sess.ToUint()
		yg.sesstions[p.sessUniq] = sess
		yg.players[p.sessUniq] = p
	})

	return yg
}

//func (yg *YGO) Dispatch(eventName string, args ...interface{}) {
//	rego.ERR(eventName, args)
//}

func (yg *YGO) SetCardVer(v *CardVersion) {
	if yg.cardVer == nil {
		yg.cardVer = v
	}
}

func (yg *YGO) registerBothEvent(eventName string) {
	yg.both[eventName] = true
}

func (yg *YGO) registerMultiEvent(eventName string) {
	yg.multi[eventName] = true
}

func (yg *YGO) chain(eventName string, ca *Card, pl *Player, args []interface{}) {
	// 全局连锁中转站

	// 给补上缺省值
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

	// 广播全局事件
	yg.EmptyEvent(Chain)
	yg.Dispatch(eventName, args...)

	cs := NewCards()
	yg.ForEventEach(Chain, func(i interface{}) {
		if v, ok := i.(*Card); ok {
			cs.EndPush(v)
		}
	})

	// 速攻
	if yg.both[eventName] || yg.multi[eventName] {
		for k, _ := range yg.quick {
			if !(k.GetSummoner().IsCurrent() || yg.multi[eventName]) {
				cs.EndPush(k)
			}
		}
	}

	// 等待用户回应
	if cs.Len() > 0 || yg.both[eventName] {
		pl.chain(eventName, ca, cs, args)
		if ca != nil && ca.IsValid() {
			pl.GetTarget().chain(eventName, ca, cs, args)
		}
	}
	yg.EmptyEvent(Chain)

}

func (yg *YGO) registerQuickPlay(c *Card) {
	yg.quick[c] = true
}

func (yg *YGO) unregisterQuickPlay(c *Card) {
	delete(yg.quick, c)
}

func (yg *YGO) getPlayer(sess *agent.Session) *Player {
	return yg.players[sess.ToUint()]
}

func (yg *YGO) InitForPlayer(sess *agent.Session, id uint, name string, d *deck) {
	p := yg.getPlayer(sess)
	p.name = name
	p.id = id
	p.decks = d

}

func (yg *YGO) AddCodeForPlayer(sess *agent.Session, uniq, method uint) {
	yg.getPlayer(sess).AddCode(uniq, method)
}

func (yg *YGO) getCard(u uint) (c *Card) {
	c = yg.cards[u]
	return
}

func (yg *YGO) registerCards(c *Card) {
	yg.cards[c.ToUint()] = c
}

func (yg *YGO) forEachPlayer(fun func(*Player)) {
	for _, v := range yg.round {
		fun(yg.players[v])
	}
}

func (yg *YGO) call(method string, reply interface{}, pl *Player) error {
	return yg.Room.Push(Call{
		Method: method,
		Args:   reply,
	}, yg.sesstions[pl.sessUniq])
}

func (yg *YGO) callAll(method string, reply interface{}) error {
	yg.Room.Broadcast(Call{
		Method: method,
		Args:   reply,
	})
	return nil
}

func (yg *YGO) getPlayerForIndex(i int) *Player {
	return yg.players[yg.round[i]]
}

func (yg *YGO) Loop() {
	defer func() {
		if x := recover(); x != nil {
			base.DebugStack()
		}
	}()

	// 服务端初始化
	for k, _ := range yg.players {
		yg.round = append(yg.round, k)
	}

	for k, v := range yg.round {
		ca := yg.players[v].camp
		yg.survival[ca] = yg.survival[ca] + 1
		yg.players[v].index = k
		yg.players[v].game = yg
		yg.players[v].roundSize = 0

		if yg.players[v].id == 0 || yg.players[v].name == "" {
			yg.players[v].name = "Guest"
		}
	}

	// 客户端初始化
	gi := gameInitResponse{}
	for _, v := range yg.round {
		pi := playerInit{
			//Hp:   yg.Players[v].Hp,
			Name: yg.players[v].name,
		}
		gi.Users = append(gi.Users, pi)
	}
	for _, v := range yg.round {
		gi.Index = yg.players[v].index
		yg.players[v].call("init", gi)
	}

	//nap(10) // 界面初始化
	i := 31
	for _, v := range yg.round {
		i++
		yg.players[v].initPlayer(i)
	}

	//nap(10) // 牌组初始化
	for _, v := range yg.round {
		yg.players[v].initDeck()
	}

	nap(20) // 手牌初始化
	for _, v := range yg.round {
		yg.players[v].init()
		yg.players[v].ChangeLp(8000)
	}

	//必要连锁初始化
	yg.registerBothEvent(Summon)
	yg.registerBothEvent(SummonFlip)
	yg.registerBothEvent(SummonSpecial)
	yg.registerBothEvent(Declaration)
	yg.registerBothEvent(UseTrap)
	yg.registerBothEvent(UseSpell)
	//yg.registerMultiEvent(DP)
	yg.registerMultiEvent(SP)
	yg.registerMultiEvent(MP)
	//yg.registerMultiEvent(EP)

	nap(10) // 游戏开始

	pl := yg.getPlayerForIndex(0)
	pl.MsgPub("msg.001", nil)
	if pl.Portrait().Len() == 1 {
		ca := pl.Portrait().Get(0)
		ca.RegisterGlobalListen(BP, func(tar *Player) {
			tar.Mzone().ForEach(func(c *Card) bool {
				c.SetNotCanAttack()
				return true
			})
		})
		ca.RegisterGlobalListen(RoundEnd, func() {
			ca.UnregisterAllGlobalListen()
		})
	}

	yg.Room.LeaveEvent(func(sess *agent.Session) {
		pl := yg.players[sess.ToUint()]
		yg.current.MsgPub("msg.009", Arg{"rival": pl.name})
		pl.Fail()
	})
loop:
	for {
		for _, v := range yg.round {
			nap(5)
			yg.current = yg.players[v]
			yg.current.round()
			yg.forEachPlayer(func(pl *Player) {
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
	defer DebugStack()
	yg.callAll(over(yg))
	yg.current.MsgPub("msg.000", nil)
}
