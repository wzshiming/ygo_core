package ygo_core

import (
	"fmt"
	"time"

	"github.com/wzshiming/dispatcher"
)

type Player struct {
	//
	dispatcher.Events
	MsgChan
	name     string // 用户名
	id       uint   //
	decks    *deck  //卡组数据
	sessUniq uint   // 会话
	phases   lp_type
	tophases lp_type

	// 规则属性
	index     int           // 玩家索引
	game      *YGO          // 属于游戏
	overTime  time.Duration // 允许超出的时间
	waitTime  time.Duration // 每次动作等待的时间
	replyTime time.Duration // 回应等待的时间
	passTime  time.Duration // 经过的时间
	// 基础属性
	lp        int  // 生命值
	camp      int  // 阵营
	roundSize int  // 回合数
	drawSize  uint // 抽卡数
	maxHp     uint // 最大生命值
	maxSdi    int  // 最大手牌

	// 卡牌区

	area map[ll_type]*Group

	// 其他的
	lastSummonRound int // 最后召唤回合

	// 回合中
	rounding bool
	noskip   bool

	// 是否失败
	fail bool
}

func newPlayer(yg *YGO) *Player {
	pl := &Player{
		Events:    dispatcher.NewForkEvent(yg.GetFork()),
		camp:      1,
		lp:        0,
		drawSize:  1,
		maxHp:     ^uint(0),
		maxSdi:    6,
		overTime:  time.Second * 120,
		waitTime:  time.Second * 60,
		replyTime: time.Second * 20,
	}
	var pr uint
	pl.MsgChan = NewMsgChan(func(m AskCode) bool {

		if m.Uniq != 0 {
			if ca := pl.Game().getCard(m.Uniq); ca != nil {
				//fmt.Println(ca)
				if m.Method == uint(LI_Over) {
					if pr != 0 {
						pl.Game().callAll(offTouch(pr))
						pr = 0
					}
					pl.GetTarget().call(onTouch(m.Uniq))
				} else if m.Method == uint(LI_Out) {
					if pr == m.Uniq {
						pr = 0
					}
					pl.Game().callAll(offTouch(m.Uniq))
				} else {
					return true
				}
			}
		} else if m.Method == LI_Defeat {
			pl.Fail()
		} else {
			return true
		}
		return false
	})
	pl.area = map[ll_type]*Group{}

	pl.AddEvent(RoundBegin, func() {
		pl.MsgPub("msg.002", Arg{"round": fmt.Sprint(pl.GetRound())})
		pl.SetCanSummon()
		pl.rounding = true
	})
	pl.AddEvent(RoundEnd, func() {
		pl.MsgPub("msg.003", Arg{"round": fmt.Sprint(pl.GetRound())})
		pl.SetCanSummon()
		pl.rounding = false
	})

	pl.AddEvent(ChangeLp, pl.changeLp)
	return pl
}

func (pl *Player) getArea(ll ll_type) *Group {
	if pl.area[ll] == nil {
		pl.area[ll] = NewGroup(pl, ll)
	}
	return pl.area[ll]
}

func (pl *Player) Deck() *Group {
	return pl.getArea(LL_Deck)
}

func (pl *Player) DeckTop(i int) *Cards {
	return pl.Deck().EndPeek(i)
}

func (pl *Player) DeckBot(i int) *Cards {
	return pl.Deck().BeginPeek(i)
}

func (pl *Player) Mzone() *Group {
	return pl.getArea(LL_Mzone)
}

func (pl *Player) Szone() *Group {
	return pl.getArea(LL_Szone)
}

func (pl *Player) SzoneAndField() *Cards {
	return NewCards(pl.Szone(), pl.Field())
}

func (pl *Player) Grave() *Group {
	return pl.getArea(LL_Grave)
}

func (pl *Player) Removed() *Group {
	return pl.getArea(LL_Removed)
}

func (pl *Player) Hand() *Group {
	return pl.getArea(LL_Hand)
}

func (pl *Player) Field() *Group {
	return pl.getArea(LL_Field)
}

func (pl *Player) Extra() *Group {
	return pl.getArea(LL_Extra)
}

func (pl *Player) Portrait() *Group {
	return pl.getArea(LL_Portrait)
}

func (pl *Player) DispatchLocal(eventName string, args ...interface{}) {
	pl.Events.Dispatch(eventName, args...)
}

func (pl *Player) DispatchGlobal(eventName string, args ...interface{}) {
	if pl.Events.IsOpen(eventName) {
		yg := pl.Game()
		yg.chain(eventName, nil, pl, args)
	}
	pl.Events.Dispatch(eventName, args...)
}

func (pl *Player) Dispatch(eventName string, args ...interface{}) {

	if !pl.Events.IsOpen(eventName) {
		return
	}
	pl.DispatchGlobal(Pre+eventName, eventName)
	pl.DispatchGlobal(eventName, args...)
	pl.DispatchGlobal(Suf+eventName, eventName)

}

func (pl *Player) Game() *YGO {
	return pl.game
}

func (pl *Player) Msg(fmts string, a Arg) {
	if a == nil {
		a = map[string]interface{}{}
	}
	if a["self"] == nil {
		a["self"] = pl.name
	}
	if a["rival"] == nil {
		a["rival"] = pl.GetTarget().name
	}
	pl.call(message(fmts, a))
}

func (pl *Player) MsgPub(fmts string, a Arg) {
	if a == nil {
		a = map[string]interface{}{}
	}
	if a["self"] == nil {
		a["self"] = pl.name
	}
	if a["rival"] == nil {
		a["rival"] = pl.GetTarget().name
	}
	pl.callAll(message(fmts, a))
}

func (pl *Player) Fail() {
	pl.fail = true
	pl.Skip(LP_End)
	pl.outTime()
	pl.AddCode(0, uint(LP_End))
	pl.AddCode(0, uint(LP_End))
}

func (pl *Player) Skip(lp lp_type) {
	pl.noskip = false
	pl.tophases = lp
}

func (pl *Player) IsFail() bool {
	return pl.fail
}

func (pl *Player) ForEachPlayer(fun func(p *Player)) {
	pl.Game().forEachPlayer(fun)
}

func (pl *Player) chain(eventName string, ca *Card, cs *Cards, a []interface{}) bool {
	pl.noskip = true
	t := pl.phases
	r := pl.passTime

	//defer DebugStack()
	// 储存连锁前的时间和阶段
	// 连锁结束后恢复
	defer func() {
		pl.phases = t
		pl.passTime = r
		if pl.rounding {
			pl.callAll(flashStep(pl))
		}
	}()
	yg := pl.Game()
	pl.phases = LP_Chain
	if !yg.multi[eventName] {
		pl.resetReplyTime()
	} else {
		pl.resetWaitTime()
	}
	pl.callAll(flashStep(pl))

	cs0 := cs.Find(func(c *Card) bool {
		return c != ca && c.GetSummoner() == pl
	})

	if cs0.Len() == 0 {
		cs1 := pl.Szone().Find(func(c *Card) bool {
			return c.IsFaceDown()
		})
		if cs1.Len() == 0 {
			return false
		}
	}

	if ca != nil {
		pl.MsgPub("msg.004", Arg{"rival": ca.ToUint(), "event": eventName})
	} else {
		pl.MsgPub("msg.005", Arg{"event": eventName})
	}

	c, u := pl.selectForSelf(cs0)
	if c == nil {
		if u == LI_No {
			// 除非输入不连锁 否则直到时间结束
			return false
		} else if pl.IsOutTime() {
			// 如果连锁事件超时
			return false
		}
		return true
	}
	uu := Chain + fmt.Sprint(u)
	//优先级 比较
	if ca == nil || yg.multi[eventName] {
		pl.MsgPub("msg.006", Arg{"self": c.ToUint(), "event": eventName})
		c.Dispatch(uu, a...)
	} else if ca.Priority() <= c.Priority() {
		pl.MsgPub("msg.008", Arg{"self": c.ToUint(), "rival": ca.ToUint(), "event": eventName})
		c.Dispatch(uu, a...)
	} else {
		// 连锁的卡牌优先级大于先发动的卡牌 所以推迟发动先发的卡牌
		ca.OnlyOnce(Suf+eventName, func() {
			pl.MsgPub("msg.007", Arg{"self": c.ToUint(), "rival": ca.ToUint(), "event": eventName})
			c.Dispatch(uu, a...)
		}, c)
	}

	return pl.noskip

}

func (pl *Player) GetRound() int {
	return pl.roundSize
}

func (pl *Player) IsInEP() bool {
	return pl.Phases() == LP_End
}

func (pl *Player) IsInDP() bool {
	return pl.Phases() == LP_Draw
}

func (pl *Player) IsInSP() bool {
	return pl.Phases() == LP_Standby
}

func (pl *Player) IsInBP() bool {
	return pl.Phases() == LP_Battle
}

func (pl *Player) IsInMP1() bool {
	return pl.Phases() == LP_Main1
}

func (pl *Player) IsInMP2() bool {
	return pl.Phases() == LP_Main2
}

func (pl *Player) IsInMP() bool {
	return pl.IsInMP1() || pl.IsInMP2()
}

func (pl *Player) Phases() lp_type {
	return pl.phases
}

func (pl *Player) round() (err error) {
	defer DebugStack()

	// 一个回合阶段流程
	pl.roundSize++
	pl.Dispatch(RoundBegin)

	pl.tophases = LP_Chain

	pl.phases = LP_Draw
	pl.callAll(flashStep(pl))
	pl.Dispatch(DP, LP_Draw)

	pl.phases = LP_Standby
	pl.callAll(flashStep(pl))
	pl.Dispatch(SP, LP_Standby)

	pl.phases = LP_Main1
	pl.callAll(flashStep(pl))
	pl.Dispatch(MP, LP_Main1)

	if pl.tophases != LP_End {
		pl.phases = LP_Battle
		pl.callAll(flashStep(pl))
		pl.Dispatch(BP, LP_Battle)
		if pl.tophases != LP_End {
			pl.phases = LP_Main2
			pl.callAll(flashStep(pl))
			pl.Dispatch(MP, LP_Main2)
		}
	}

	pl.phases = LP_End
	pl.callAll(flashStep(pl))
	pl.Dispatch(EP, LP_End)

	pl.Dispatch(RoundEnd)
	pl.phases = LP_Chain
	return
}

func (pl *Player) initPlayer(u int) {
	if pl.Portrait().Len() > 0 {
		return
	}
	c := NewPortraitCardOriginal().Make(pl)
	pl.Portrait().EndPush(c)
	pl.callAll(setPortrait(c, u))
}

func (pl *Player) initDeck() {
	if pl.Deck().Len() > 0 {
		return
	}

	pl.Game().cardVer.Deck(pl)
	pl.Shuffle()
}

func (pl *Player) GetLp() int {
	return pl.lp
}

func (pl *Player) ChangeLp(i int) {
	pl.Dispatch(ChangeLp, i)
}

func (pl *Player) changeLp(i int) {
	if i < 0 {
		pl.MsgPub("msg.201", Arg{"num": fmt.Sprint(-i)})
	} else if i > 0 {
		pl.MsgPub("msg.202", Arg{"num": fmt.Sprint(i)})
	} else {
		return
	}
	pl.lp += i
	if pl.lp <= 0 {
		pl.Fail()
	}
	pl.callAll(changeHp(pl))
}
func (pl *Player) Coins(i int) int {
	return RandInt(i)
}

func (pl *Player) GetTarget() *Player {
	// 选择目标 如果是多人模式 可能要加一个 选择对方用户的选项
	if pl.index == 0 {
		return pl.Game().getPlayerForIndex(1)
	}
	return pl.Game().getPlayerForIndex(0)
}

func (pl *Player) Shuffle() {
	// 洗牌
	pl.Deck().Shuffle()
}

func (pl *Player) DrawCard(s int) {
	// 抽牌
	if s <= 0 {
		return
	}
	for i := 0; i != s; i++ {
		if pl.Deck().Len() == 0 {
			pl.Fail()
			return
		}
		pl.Dispatch(DrawNum, pl)
		t := pl.Deck().EndPop()
		pl.Hand().EndPush(t)
	}
	pl.Dispatch(Draw, pl)
}

// 是当前的回合者
func (pl *Player) IsCurrent() bool {
	return pl == pl.Game().current
}

// 获得当前回合者
func (pl *Player) GetCurrent() *Player {
	return pl.Game().current
}

func (pl *Player) call(method string, reply interface{}) error {
	return pl.Game().call(method, reply, pl)
}

func (pl *Player) callAll(method string, reply interface{}) error {
	if pl.roundSize != 0 {
		nap(1)
	}
	return pl.Game().callAll(method, reply)
}
func (pl *Player) IsOutTime() bool {
	return pl.passTime == 0
}
func (pl *Player) outTime() {
	pl.passTime = 0
}

func (pl *Player) resetReplyTime() {
	pl.passTime = pl.replyTime
}

func (pl *Player) resetWaitTime() {
	pl.passTime = pl.waitTime
}

func (pl *Player) IsCanSummon() bool {
	return pl.lastSummonRound != 0
}

func (pl *Player) SetCanSummon() {
	pl.lastSummonRound = 1
}

func (pl *Player) SetNotCanSummon() {
	pl.lastSummonRound = 0
}

func (pl *Player) isOutTime() bool {
	return pl.passTime <= 0
}

func (pl *Player) selectWill() (p AskCode) {
	pl.callAll(flashStep(pl))
	for {
		select {
		case <-time.After(time.Second):
			pl.passTime -= time.Second
			if pl.IsOutTime() {
				return
			}
		case p = <-pl.GetCode():
			return
		}
	}
	return
}

func (pl *Player) selectFor() (*Card, uint) {
	p := pl.selectWill()
	if p.Uniq != 0 {
		return pl.Game().getCard(p.Uniq), p.Method
	}
	return nil, p.Method
}

func (pl *Player) selectForSelf(ci ...interface{}) (c *Card, u uint) {
	css := NewCards(ci...)
	pl.call(setPickRe(css, pl))
	if css.Len() == 0 {
		// 如果连锁时没有可以发动的卡 延迟一段时间 直接返回 照顾手残党
		nap(RandInt(20) + 5)
		return nil, LI_No
	}
	defer pl.call(cloPick(pl))
	if c, u = pl.selectFor(); c != nil {
		if css.IsExistCard(c) {
			return
		}
	}
	return nil, u
}

//  可选一个卡牌 如果超时则返回nil

func (pl *Player) selectHead(ci []interface{}) *Cards {
	css := NewCards(ci...)
	css.ForEach(func(c0 *Card) bool {
		if c0.GetSummoner() == pl && c0.IsInDeck() {
			c0.Peek()
		}
		return true
	})
	return css
}

func (pl *Player) SelectChoosable(lo lo_type, ci ...interface{}) *Card {
	css := pl.selectHead(ci)

	pl.call(setPick(lo, css, pl))
	defer pl.call(cloPick(pl))
	for {
		c, u := pl.selectFor()
		if c != nil && css.IsExistCard(c) {
			return c
		}

		if pl.IsOutTime() || u == LI_No || css.Len() == 0 {
			return nil
		}
	}
	return nil
}

func (pl *Player) SelectRequiredRange(i, j int, lo lo_type, ci ...interface{}) *Cards {
	if i > j || j <= 0 {
		return nil
	}

	css := pl.selectHead(ci)

	if css.Len() < i {
		return nil
	} else if css.Len() == i {
		return css
	}
	css0 := NewCards()
	pl.call(setPick(lo, css, pl))
	defer pl.call(cloPick(pl))
	for css0.Len() < j {

		c, u := pl.selectFor()
		if c != nil && css.IsExistCard(c) {
			css0.EndPush(c)
			css.PickedFor(c)
			pl.call(cloPickOne(pl, c))
		} else if u == LI_No && css.Len() >= i {
			break
		} else if pl.IsOutTime() {
			if css.Len() >= i {
				break
			}
			css0.EndPush(css.EndPop())
		}
	}
	return css0
}

// 必选一个卡牌 如果没有卡牌则返回nil  如果超时则返回最后一个 卡牌

func (pl *Player) SelectRequired(lo lo_type, ci ...interface{}) *Card {
	css := pl.selectHead(ci)
	pl.call(setPick(lo, css, pl))
	defer pl.call(cloPick(pl))
	for css.Len() != 0 {
		if pl.IsOutTime() {
			return css.EndPop()
		}
		c, _ := pl.selectFor()
		if c != nil && css.IsExistCard(c) {
			return c
		}
	}
	return nil
}

func (pl *Player) SelectRequiredShor(lo lo_type, ci ...interface{}) *Card {
	css := NewCards(ci...)
	if css.Len() == 1 {
		return css.EndPop()
	}
	return pl.SelectRequired(lo, css)
}
