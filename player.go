package ygo_core

import (
	"fmt"
	"time"

	"github.com/wzshiming/base"
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

	// 规则属性
	index     int           // 玩家索引
	game      *YGO          // 属于游戏
	overTime  time.Duration // 允许超出的时间
	waitTime  time.Duration // 每次动作等待的时间
	replyTime time.Duration // 回应等待的时间
	passTime  time.Duration // 经过的时间
	// 基础属性
	hp        int  // 生命值
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
	// 是否失败
	fail bool
}

func newPlayer(yg *YGO) *Player {
	pl := &Player{
		Events:    dispatcher.NewForkEvent(yg.GetFork()),
		camp:      1,
		hp:        0,
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

	pl.AddEvent(DP, pl.draw)
	//pl.AddEvent(SP, pl.standby)
	pl.AddEvent(MP, pl.main)
	pl.AddEvent(BP, pl.battle)
	pl.AddEvent(EP, pl.end)

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

func (pl *Player) Mzone() *Group {
	return pl.getArea(LL_Mzone)
}

func (pl *Player) Szone() *Group {
	return pl.getArea(LL_Szone)
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

func (pl *Player) Dispatch(eventName string, args ...interface{}) {
	yg := pl.Game()
	if pl.IsOpen(eventName) {
		yg.chain(eventName, nil, pl, append(args))
	}
	pl.Events.Dispatch(eventName, args...)
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
	//pl.Empty()
	pl.AddCode(0, uint(LP_End))
	pl.AddCode(0, uint(LP_End))
}

func (pl *Player) IsFail() bool {
	return pl.fail
}

func (pl *Player) ForEachPlayer(fun func(p *Player)) {
	pl.Game().forEachPlayer(fun)
}

func (pl *Player) chain(eventName string, ca *Card, cs *Cards, a []interface{}) bool {
	t := pl.phases
	r := pl.passTime
	defer func() {
		if x := recover(); x != nil {
			if _, ok := x.(string); ok {

			} else {
				base.DebugStack()
			}
		}
	}()
	defer func() {
		pl.phases = t
		pl.passTime = r
		if pl.rounding {
			pl.callAll(flashStep(pl))
		}
	}()
	yg := pl.Game()
	pl.phases = LP_Chain
	pl.resetReplyTime()
	pl.callAll(flashStep(pl))

	for {
		if pl.IsOutTime() {
			break
		}
		cs0 := cs.Find(func(c *Card) bool {
			return c != ca && c.GetSummoner() == pl
		})

		if cs0.Len() == 0 {
			cs1 := pl.Szone().Find(func(c *Card) bool {
				return c.IsFaceDown()
			})
			if cs1.Len() == 0 {
				break
			}
		}

		if ca != nil {
			pl.MsgPub("msg.004", Arg{"rival": ca.ToUint(), "event": eventName})
		} else {
			pl.MsgPub("msg.005", Arg{"event": eventName})
		}
		c, _ := pl.selectForWarn(LO_Chain, cs0)
		if c == nil {
			break
			//			if u == LI_No {
			//				break
			//			}
			//			continue
		}

		if ca == nil {
			pl.MsgPub("msg.006", Arg{"self": c.ToUint(), "event": eventName})
			c.Dispatch(Trigger, a...)
		} else if ca.Priority() > c.Priority() {
			ca.OnlyOnce(eventName, func() {
				pl.MsgPub("msg.007", Arg{"self": c.ToUint(), "rival": ca.ToUint(), "event": eventName})
				c.Dispatch(Trigger, a...)
			}, c)
		} else {
			pl.MsgPub("msg.008", Arg{"self": c.ToUint(), "rival": ca.ToUint(), "event": eventName})
			c.Dispatch(Trigger, a...)
		}

		if !yg.multi[eventName] {
			break
		}
		cs.PickedFor(c)
	}

	return false
}

func (pl *Player) GetRound() int {
	return pl.roundSize
}

func (pl *Player) round() (err error) {
	defer func() {
		if x := recover(); x != nil {
			if _, ok := x.(string); ok {

			} else {
				base.DebugStack()
			}
		}
	}()
	pl.roundSize++
	pl.Dispatch(RoundBegin)

	pl.phases = LP_Draw
	pl.callAll(flashStep(pl))
	pl.Dispatch(DP, LP_Draw)

	pl.phases = LP_Standby
	pl.callAll(flashStep(pl))
	pl.Dispatch(SP, LP_Standby)

	pl.phases = LP_Main1
	pl.callAll(flashStep(pl))
	pl.Dispatch(MP, LP_Main1)

	pl.phases = LP_Battle
	pl.callAll(flashStep(pl))
	pl.Dispatch(BP, LP_Battle)

	pl.phases = LP_Main2
	pl.callAll(flashStep(pl))
	pl.Dispatch(MP, LP_Main2)

	pl.phases = LP_End
	pl.callAll(flashStep(pl))
	pl.Dispatch(EP, LP_End)

	pl.Dispatch(RoundEnd)
	pl.phases = LP_Chain
	return
}

func (pl *Player) draw(lp lp_type) {
	pl.DrawCard(1)
}

func (pl *Player) main(lp lp_type) {
	defer func() {
		if x := recover(); x != nil {
			if _, ok := x.(string); ok {

			} else {
				base.DebugStack()
			}
		}
	}()

	pl.resetWaitTime()
	for {

		ca, u := pl.selectForMain(
			LO_Onset, LO_Cover, pl.Hand(), func(c *Card) bool {
				return c.IsSpell()
			}, LO_Summon, LO_Cover, pl.Hand(), func(c *Card) bool {
				return pl.IsCanSummon() && c.IsMonster() && c.GetLevel() <= 4
			}, LO_SummonFreedom, LO_CoverFreedom, pl.Hand(), func(c *Card) bool {
				return pl.IsCanSummon() && c.IsMonster() && ((c.GetLevel() > 4 && pl.Mzone().Len() >= 1) || (c.GetLevel() > 6 && pl.Mzone().Len() >= 2))
			}, LO_Cover, pl.Hand(), func(c *Card) bool {
				return c.IsTrap()
			}, LO_Onset, pl.Szone(), func(c *Card) bool {
				return c.IsSpell()
			}, LO_Expres, pl.Mzone(), func(c *Card) bool {
				return c.IsMonster() && c.IsCanChange()
			})

		if ca == nil {
			if u == uint(LP_Battle) && lp == LP_Main1 {
				break
			} else if u == uint(LP_End) {
				if lp == LP_Main1 {
					pl.StopOnce(MP)
					pl.StopOnce(BP)
				}
				break
			}
			if u == 0 {
				break
			}
			continue
		}

		if ca.IsInHand() {
			if u == uint(LI_Use1) {
				ca.Dispatch(Use1)
			} else if u == uint(LI_Use2) {
				ca.Dispatch(Use2)
			}
		} else if ca.IsInMzone() {
			ca.Dispatch(Expression)
		} else if ca.IsInSzone() {
			ca.Dispatch(Onset)
		} else {
			Debug(ca)
			pl.Msg("101", nil)
		}
	}
}

func (pl *Player) battle(lp lp_type) {
	defer func() {
		if x := recover(); x != nil {
			if _, ok := x.(string); ok {

			} else {
				base.DebugStack()
			}
		}
	}()

	pl.resetWaitTime()
	for {

		css := NewCards(pl.Mzone(), func(c *Card) bool {
			return c.IsFaceUpAttack() && c.IsCanAttack()
		})
		if css.Len() == 0 {
			break
		}
		ca, u := pl.selectForWarn(LO_Attack, css)
		if ca == nil {
			if u == uint(LP_Main2) {
				break
			} else if u == uint(LP_End) {
				pl.StopOnce(MP)
				break
			}
			if u == 0 {
				break
			}
			continue
		}

		tar := pl.GetTarget()
		pl.Msg("102", nil)

		var c *Card
		if tar.Mzone().Len() != 0 {
			if ca.IsCanDirect() {
				c = pl.SelectForWarnShort(LO_Target, 1, tar.Mzone(), tar.Portrait())
			} else {
				c = pl.SelectForWarnShort(LO_Target, 1, tar.Mzone())
			}
		} else {
			c = pl.SelectForWarnShort(LO_Target, 1, tar.Portrait())
		}
		if c != nil {
			ca.Dispatch(Declaration, c)
		}

	}
}

func (pl *Player) end(lp lp_type) {
	if i := pl.Hand().Len() - pl.maxSdi; i > 0 {
		pl.resetReplyTime()
		pl.Msg("103", nil)
		for k := 0; k != i; k++ {
			ca := pl.SelectForWarn(LO_Discard, pl.Hand())
			if ca == nil {
				ca = pl.Hand().EndPop()
			}
			ca.Dispatch(Discard)
		}
	}
}

func (pl *Player) init() {
	pl.DrawCard(5)
}

func (pl *Player) initPlayer(u int) {
	if pl.Portrait().Len() > 0 {
		return
	}
	c := NewNoneCardOriginal().Make(pl)
	pl.Portrait().EndPush(c)
	pl.callAll(setPortrait(c, u))
}

func (pl *Player) initDeck() {
	if pl.Deck().Len() > 0 {
		return
	}

	pl.Game().cardVer.Deck(pl)
	pl.ActionShuffle()
}

func (pl *Player) GetHp() int {
	return pl.hp
}

func (pl *Player) ChangeHp(i int) {
	if i < 0 {
		pl.MsgPub("msg.201", Arg{"num": fmt.Sprint(-i)})
	} else if i > 0 {
		pl.MsgPub("msg.202", Arg{"num": fmt.Sprint(i)})
	}
	pl.hp += i
	if pl.hp < 0 {
		pl.Fail()
	}
	pl.Dispatch(ChangeHP, pl, i)
	pl.callAll(changeHp(pl, pl.hp))

}

func (pl *Player) GetTarget() *Player {
	if pl.index == 0 {
		return pl.Game().getPlayerForIndex(1)
	}
	return pl.Game().getPlayerForIndex(0)
}

func (pl *Player) ActionShuffle() {
	pl.Deck().Shuffle()
}

func (pl *Player) DrawCard(s int) {
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
	return pl.lastSummonRound < pl.GetRound()
}

func (pl *Player) SetCanSummon() {
	pl.lastSummonRound = 0
}

func (pl *Player) SetNotCanSummon() {
	pl.lastSummonRound = pl.GetRound()
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

func (pl *Player) SelectForPopup(lo lo_type, ci ...interface{}) *Card {
	css := NewCards(ci...)
	css.ForEach(func(c *Card) bool {
		c.Peek()
		return true
	})
	return pl.SelectForWarn(lo, css)
}

func (pl *Player) selectForMain(cc ...interface{}) (c *Card, u uint) {
	b := map[lo_type][]interface{}{}
	e := true
	n := LO_None
	for _, v := range cc {
		if s, ok := v.(lo_type); ok {
			if !e {
				n = LO_None
				e = true
			}
			if n != LO_None {
				n += ","
			}
			n += s
		} else {
			b[n] = append(b[n], v)
			e = false
		}
	}
	css := NewCards()
	for k, v := range b {
		cs := NewCards(v...)
		css.Join(cs)
		pl.call(setPick(k, cs, pl))
	}

	defer pl.call(cloPick(pl))
	if c, u = pl.selectFor(); c != nil {
		if css.IsExistCard(c) {
			return
		}
	}
	return nil, u
}

func (pl *Player) selectForWarn(lo lo_type, ci ...interface{}) (c *Card, u uint) {
	css := NewCards(ci...)
	pl.call(setPick(lo, css, pl))
	defer pl.call(cloPick(pl))
	if c, u = pl.selectFor(); c != nil {
		if css.IsExistCard(c) {
			return
		}
	}
	return nil, u
}

// 用户选择卡牌 如果 少于i则直接返回 最后一个卡牌

func (pl *Player) SelectForWarnShort(lo lo_type, i int, ci ...interface{}) (c *Card) {
	css := NewCards(ci...)
	if css.Len() == 0 {
		return nil
	} else if css.Len() <= i {
		return css.EndPop()
	}
	return pl.SelectForWarn(lo, css)

}

// 直到用户选择正确的卡牌

func (pl *Player) SelectForWarn(lo lo_type, ci ...interface{}) *Card {
	css := NewCards(ci...)
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
