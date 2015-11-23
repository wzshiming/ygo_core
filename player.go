package ygo_core

import (
	"fmt"
	"time"

	"github.com/wzshiming/base"
	"github.com/wzshiming/dispatcher"
	"github.com/wzshiming/server/agent"
)

type Player struct {
	//
	dispatcher.Events
	MsgChan
	Name    string         // 用户名
	Session *agent.Session // 会话
	Phases  lp_type

	// 规则属性
	Index     int           // 玩家索引
	game      *YGO          // 属于游戏
	OverTime  time.Duration // 允许超出的时间
	WaitTime  time.Duration // 每次动作等待的时间
	ReplyTime time.Duration // 回应等待的时间
	PassTime  time.Duration // 经过的时间
	// 基础属性
	Hp        int  // 生命值
	Camp      int  // 阵营
	RoundSize int  // 回合数
	DrawSize  uint // 抽卡数
	MaxHp     uint // 最大生命值
	MaxSdi    int  // 最大手牌

	// 卡牌区
	Deck    *Group // 卡组 40 ~ 60
	Hand    *Group // 手牌
	Extra   *Group // 额外卡组 <= 15 融合怪物 同调怪物 超量怪物
	Removed *Group // 排除卡
	Grave   *Group // 墓地
	Mzone   *Group // 怪物卡区 5
	Szone   *Group // 魔法卡陷阱卡区 5
	Field   *Group // 场地卡

	// 特殊区
	Side     *Group // 副卡组 <= 15 不参与游戏
	Portrait *Group // 玩家头像

	// 其他的
	lastSummonRound int // 最后召唤回合

	// 回合中
	rounding bool
	// 是否失败
	fail bool
}

func NewPlayer(yg *YGO) *Player {
	pl := &Player{
		Events:    dispatcher.NewForkEvent(yg.GetFork()),
		Camp:      1,
		Hp:        0,
		DrawSize:  1,
		MaxHp:     ^uint(0),
		MaxSdi:    6,
		OverTime:  time.Second * 120,
		WaitTime:  time.Second * 60,
		ReplyTime: time.Second * 20,
	}
	var pr uint
	pl.MsgChan = NewMsgChan(func(m MsgCode) bool {

		if m.Uniq != 0 {
			if ca := pl.Game().GetCard(m.Uniq); ca != nil {
				if m.Method == uint(LI_Over) {
					if pr != 0 {
						pl.Game().CallAll(offTouch(pr))
						pr = 0
					}
					pl.GetTarget().Call(onTouch(m.Uniq))
				} else if m.Method == uint(LI_Out) {
					if pr == m.Uniq {
						pr = 0
					}
					pl.Game().CallAll(offTouch(m.Uniq))
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
	pl.Deck = NewGroup(pl, LL_Deck)
	pl.Hand = NewGroup(pl, LL_Hand)
	pl.Extra = NewGroup(pl, LL_Extra)
	pl.Removed = NewGroup(pl, LL_Removed)
	pl.Grave = NewGroup(pl, LL_Grave)
	pl.Mzone = NewGroup(pl, LL_Mzone)
	pl.Szone = NewGroup(pl, LL_Szone)
	pl.Field = NewGroup(pl, LL_Field)
	pl.Portrait = NewGroup(pl, LL_Portrait)
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

	pl.AddEvent(ChangeHP, pl.changeHp)
	return pl
}

func (pl *Player) Dispatch(eventName string, args ...interface{}) {
	yg := pl.Game()
	if pl.IsOpen(eventName) {
		yg.Chain(eventName, nil, pl, append(args))
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
		a["self"] = pl.Name
	}
	if a["rival"] == nil {
		a["rival"] = pl.GetTarget().Name
	}
	pl.Call(message(fmts, a))
}

func (pl *Player) MsgPub(fmts string, a Arg) {
	if a == nil {
		a = map[string]interface{}{}
	}
	if a["self"] == nil {
		a["self"] = pl.Name
	}
	if a["rival"] == nil {
		a["rival"] = pl.GetTarget().Name
	}
	pl.CallAll(message(fmts, a))
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
	pl.Game().ForEachPlayer(fun)
}

func (pl *Player) Chain(eventName string, ca *Card, cs *Cards, a []interface{}) bool {
	t := pl.Phases
	r := pl.PassTime
	defer func() {
		if x := recover(); x != nil {
			if _, ok := x.(string); ok {

			} else {
				base.DebugStack()
			}
		}
	}()
	defer func() {
		pl.Phases = t
		pl.PassTime = r
		if pl.rounding {
			pl.CallAll(flashStep(pl))
		}
	}()
	yg := pl.Game()
	pl.Phases = LP_Chain
	pl.ResetReplyTime()
	pl.CallAll(flashStep(pl))

	for {
		if pl.IsOutTime() {
			break
		}
		cs0 := cs.Find(func(c *Card) bool {
			return c != ca && c.GetSummoner() == pl
		})

		if cs0.Len() == 0 {
			cs1 := pl.Szone.Find(func(c *Card) bool {
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
		c, u := pl.selectForWarn(cs0)
		if c == nil {
			if u == LI_No {
				break
			}
			continue
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
	return pl.RoundSize
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
	pl.RoundSize++
	pl.Dispatch(RoundBegin)

	pl.Phases = LP_Draw
	pl.CallAll(flashStep(pl))
	pl.Dispatch(DP, LP_Draw)

	pl.Phases = LP_Standby
	pl.CallAll(flashStep(pl))
	pl.Dispatch(SP, LP_Standby)

	pl.Phases = LP_Main1
	pl.CallAll(flashStep(pl))
	pl.Dispatch(MP, LP_Main1)

	pl.Phases = LP_Battle
	pl.CallAll(flashStep(pl))
	pl.Dispatch(BP, LP_Battle)

	pl.Phases = LP_Main2
	pl.CallAll(flashStep(pl))
	pl.Dispatch(MP, LP_Main2)

	pl.Phases = LP_End
	pl.CallAll(flashStep(pl))
	pl.Dispatch(EP, LP_End)
	pl.Dispatch(RoundEnd)
	return
}

func (pl *Player) draw(lp lp_type) {
	pl.ActionDraw(1)
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

	pl.ResetWaitTime()
	for {
		ca, u := pl.selectForWarn(pl.Hand, pl.Mzone, pl.Szone, func(c *Card) bool {
			if c.IsInHand() && c.IsMonster() {
				if !pl.IsCanSummon() {
					return false
				} else if c.GetLevel() >= 7 && pl.Mzone.Len() < 2 {
					return false
				} else if c.GetLevel() >= 5 && pl.Mzone.Len() < 1 {
					return false
				}
			} else if c.IsInSzone() && (c.IsTrap() || c.IsFaceUp()) {
				return false
			} else if c.IsInMzone() && !c.IsCanChange() {
				return false
			}
			return true
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

	pl.ResetWaitTime()
	for {
		ca, u := pl.selectForWarn(pl.Mzone, func(c *Card) bool {
			return c.IsFaceUpAttack() && c.IsCanAttack()
		})
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
		if tar.Mzone.Len() != 0 {
			if c, _ := pl.selectForWarn(tar.Mzone, tar.Portrait, func(c0 *Card) bool {
				return !(c0.IsPortrait() && !ca.IsCanDirect())
			}); c != nil {
				ca.Dispatch(Declaration, c)
			}
		} else {
			if c, _ := pl.selectForWarn(tar.Portrait); c != nil {
				ca.Dispatch(Declaration, c)
			}
		}
	}
}

func (pl *Player) end(lp lp_type) {
	if i := pl.Hand.Len() - pl.MaxSdi; i > 0 {
		pl.ResetReplyTime()
		pl.Msg("103", nil)
		for k := 0; k != i; k++ {
			ca := pl.SelectForWarn(pl.Hand)
			if ca == nil {
				ca = pl.Hand.EndPop()
			}
			ca.Dispatch(Discard)
		}
	}
}

func (pl *Player) init() {
	pl.ActionDraw(5)
}

func (pl *Player) InitPlayer(u int) {
	if pl.Portrait.Len() > 0 {
		return
	}
	c := NewNoneCardOriginal().Make(pl)
	pl.Portrait.EndPush(c)
	pl.CallAll(setPortrait(c, u))
}

func (pl *Player) initDeck(a []uint) {
	if pl.Deck.Len() > 0 {
		return
	}

	pl.Game().CardVer.Deck(pl, a)
	pl.ActionShuffle()
}

func (pl *Player) GetHp() int {
	return pl.Hp
}

func (pl *Player) ChangeHp(i int) {
	pl.Dispatch(ChangeHP, pl, i)
}

func (pl *Player) changeHp(i int) {
	if i < 0 {
		pl.MsgPub("msg.201", Arg{"num": fmt.Sprint(-i)})
	} else if i > 0 {
		pl.MsgPub("msg.202", Arg{"num": fmt.Sprint(i)})
	}
	pl.Hp += i
	if pl.Hp < 0 {
		pl.Fail()
	}
	pl.CallAll(changeHp(pl, pl.Hp))

}

func (pl *Player) GetTarget() *Player {
	if pl.Index == 0 {
		return pl.Game().GetPlayerForIndex(1)
	}
	return pl.Game().GetPlayerForIndex(0)
}

func (pl *Player) ActionShuffle() {
	pl.Deck.Shuffle()
}

func (pl *Player) ActionDraw(s int) {
	if s <= 0 {
		return
	}
	for i := 0; i != s; i++ {
		if pl.Deck.Len() == 0 {
			pl.Fail()
			return
		}
		pl.Dispatch(DrawNum, pl)
		t := pl.Deck.EndPop()
		pl.Hand.EndPush(t)
	}
	pl.Dispatch(Draw, pl)
}

func (pl *Player) Call(method string, reply interface{}) error {
	return pl.Game().Room.Push(Call{
		Method: method,
		Args:   reply,
	}, pl.Session)
}

func (pl *Player) CallAll(method string, reply interface{}) error {
	if pl.RoundSize != 0 {
		nap(1)
	}
	return pl.Game().CallAll(method, reply)
}
func (pl *Player) IsOutTime() bool {
	return pl.PassTime == 0
}
func (pl *Player) OutTime() {
	pl.PassTime = 0
}

func (pl *Player) ResetReplyTime() {
	pl.PassTime = pl.ReplyTime
}

func (pl *Player) ResetWaitTime() {
	pl.PassTime = pl.WaitTime
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

func (pl *Player) SelectWill() (p MsgCode) {
	pl.CallAll(flashStep(pl))
	for {
		select {
		case <-time.After(time.Second):
			pl.PassTime -= time.Second
			if pl.PassTime <= 0 {
				return
			}
		case p = <-pl.GetCode():
			return
		}
	}
	return
}

func (pl *Player) Select() (*Card, uint) {
	p := pl.SelectWill()
	if p.Uniq != 0 {
		return pl.Game().GetCard(p.Uniq), p.Method
	}
	return nil, p.Method
}
func (pl *Player) selectForPopup(ci ...interface{}) (c *Card, u uint) {
	css := NewCards(ci...)
	css.ForEach(func(c *Card) bool {
		c.Peek()
		return true
	})
	return pl.selectForWarn(ci...)
}
func (pl *Player) SelectForPopup(ci ...interface{}) *Card {
	c, _ := pl.selectForPopup(ci...)
	return c
}

func (pl *Player) selectForWarn(ci ...interface{}) (c *Card, u uint) {
	css := NewCards(ci...)
	pl.Call(setPick(css, pl))
	defer pl.Call(cloPick(pl))
	if c, u = pl.Select(); c != nil {
		if css.IsExistCard(c) {
			return
		}
	}
	return nil, u
}

func (pl *Player) SelectForWarn(ci ...interface{}) *Card {
	c, _ := pl.selectForWarn(ci...)
	return c
}
