package ygo_core

import "fmt"

type MsgChan struct {
	secondary chan AskCode
	msgfunc   func(AskCode) bool
}

func NewMsgChan(m func(AskCode) bool) MsgChan {
	mc := MsgChan{
		secondary: make(chan AskCode, 16),
		msgfunc:   m,
	}
	return mc
}

func (mc *MsgChan) AddCode(uniq, method uint) {
	c := AskCode{
		Uniq:   uniq,
		Method: method,
	}
	if mc.msgfunc(c) {
		for {
			select {
			case mc.secondary <- c:
				return
			default:
				mc.ClearCode()
			}
		}
	}
}

func (mc *MsgChan) GetCode() <-chan AskCode {
	mc.ClearCode()
	return mc.secondary
}

func (mc *MsgChan) ClearCode() {
	for {
		select {
		case <-mc.secondary:
		default:
			return
		}
	}
}

type Call struct {
	Method string      `json:"method"`
	Args   interface{} `json:"args"`
}

func Init() {

}

func onTouch(t uint) (string, Arg) {
	return "onTouch", Arg{
		"uniq": t,
	}
}

func offTouch(t uint) (string, Arg) {
	return "offTouch", Arg{
		"uniq": t,
	}
}

func exprCard(t *Card, le le_type) (string, Arg) {
	return "exprCard", Arg{
		"uniq": t.ToUint(),
		"expr": le,
	}
}

func setFront(t *Card) (string, Arg) {
	return "setFront", Arg{
		"desk": t.GetId(),
		"uniq": t.ToUint(),
	}
}

func message(format string, a Arg) (string, Arg) {
	return "message", Arg{
		"format": format,
		"params": a,
	}
}

func setPickRe(cs *Cards, pl *Player) (string, Arg) {
	mi := map[string][]lo_type{}
	cs.ForEach(func(c *Card) bool {
		mi[fmt.Sprint(c.ToUint())] = c.operate
		return true
	})
	return "setPickRe", Arg{
		"master":  pl.index,
		"operate": mi,
	}
}

//给用户选择
func setPick(use lo_type, cs *Cards, pl *Player) (string, Arg) {
	return "setPick", Arg{
		"master": pl.index,
		"uniqs":  cs.Uniqs(),
		"use":    use,
	}
}

func cloPick(pl *Player) (string, Arg) {
	return "cloPick", Arg{
		"master": pl.index,
	}
}

func changeHp(pl *Player, hp int) (string, Arg) {
	return "changeHp", Arg{
		"master": pl.index,
		"hp":     hp,
	}
}

func setPortrait(t *Card, d int) (string, Arg) {
	return "setPortrait", Arg{
		"uniq": t.ToUint(),
		"desk": d,
	}
}

func setCardFace(t *Card, a Arg) (string, Arg) {
	return "setCardFace", Arg{
		"uniq":   t.ToUint(),
		"params": a,
	}
}

func moveCard(t *Card, pos ll_type) (string, Arg) {
	i := 0
	if pos == LL_Mzone {
		i = t.GetSummoner().index
	} else {
		i = t.GetOwner().index
	}
	return "moveCard", Arg{
		"uniq":   t.ToUint(),
		"master": i,
		"pos":    pos,
	}
}

func flashStep(pl *Player) (string, Arg) {
	return "flagStep", Arg{
		"step":   pl.phases,
		"wait":   pl.passTime,
		"master": pl.index,
		"round":  pl.GetRound(),
	}
}

func impact(t1 *Card, t2 *Card) (string, Arg) {
	return "impact", Arg{
		"uniq":   t1.ToUint(),
		"target": t2.ToUint(),
	}
}

func over(yg *YGO) (string, Arg) {
	return "over", Arg{}
}
