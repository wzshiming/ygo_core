package ygo_core

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

func onTouch(t uint) (string, interface{}) {
	return "onTouch", map[string]interface{}{
		"uniq": t,
	}
}

func offTouch(t uint) (string, interface{}) {
	return "offTouch", map[string]interface{}{
		"uniq": t,
	}
}

func exprCard(t *Card, le le_type) (string, interface{}) {
	return "exprCard", map[string]interface{}{
		"uniq": t.ToUint(),
		"expr": le,
	}
}

func setFront(t *Card) (string, interface{}) {
	return "setFront", map[string]interface{}{
		"desk": t.GetId(),
		"uniq": t.ToUint(),
	}
}

func message(format string, a Arg) (string, interface{}) {
	return "message", map[string]interface{}{
		"format": format,
		"params": a,
	}
}

//给用户选择
func setPick(cs *Cards, pl *Player) (string, interface{}) {
	return "setPick", map[string]interface{}{
		"master": pl.Index,
		"uniqs":  cs.Uniqs(),
	}
}

func cloPick(pl *Player) (string, interface{}) {
	return "cloPick", map[string]interface{}{
		"master": pl.Index,
	}
}

func changeHp(pl *Player, hp int) (string, interface{}) {
	return "changeHp", map[string]interface{}{
		"master": pl.Index,
		"hp":     hp,
	}
}

func setPortrait(t *Card, d int) (string, interface{}) {
	return "setPortrait", map[string]interface{}{
		"uniq": t.ToUint(),
		"desk": d,
	}
}

func setCardFace(t *Card, a Arg) (string, interface{}) {
	return "setCardFace", map[string]interface{}{
		"uniq":   t.ToUint(),
		"params": a,
	}
}

func moveCard(t *Card, pos ll_type) (string, interface{}) {
	i := 0
	if pos == LL_Mzone {
		i = t.GetSummoner().Index
	} else {
		i = t.GetOwner().Index
	}
	return "moveCard", map[string]interface{}{
		"uniq":   t.ToUint(),
		"master": i,
		"pos":    pos,
	}
}

func flashStep(pl *Player) (string, interface{}) {
	return "flagStep", map[string]interface{}{
		"step":   pl.Phases,
		"wait":   pl.PassTime,
		"master": pl.Index,
		"round":  pl.GetRound(),
	}
}
