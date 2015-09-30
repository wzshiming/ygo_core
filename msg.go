package ygo_cord

type MsgCode struct {
	Uniq   uint
	Method uint
}

type MsgChan struct {
	secondary chan MsgCode
	msgfunc   func(MsgCode) bool
}

func NewMsgChan(m func(MsgCode) bool) MsgChan {
	mc := MsgChan{
		secondary: make(chan MsgCode, 16),
		msgfunc:   m,
	}
	return mc
}

func (mc *MsgChan) AddCode(uniq, method uint) {
	c := MsgCode{
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

func (mc *MsgChan) GetCode() <-chan MsgCode {
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

func remind(t uint) (string, interface{}) {
	return "remind", map[string]interface{}{
		"uniq": t,
	}
}

func touch(t uint, x, y, z int) (string, interface{}) {
	return "touch", map[string]interface{}{
		"uniq": t,
		"x":    x,
		"y":    y,
		"z":    z,
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
		"message": format,
		"params":  a,
	}
}

func setPick(cs *Cards, pl *Player) (string, interface{}) {
	if cs != nil {
		return "setPick", map[string]interface{}{
			"master": pl.Index,
			"uniqs":  cs.Uniqs(),
		}
	}
	return "setPick", map[string]interface{}{
		"master": pl.Index,
		"uniqs":  []int{},
	}

}

func changeHp(t *Card, hp int) (string, interface{}) {
	return "changeHp", map[string]interface{}{
		"uniq": t.ToUint(),
		"hp":   hp,
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

func flashPhases(pl *Player) (string, interface{}) {
	return "flagStep", map[string]interface{}{
		"step": pl.Phases,
		"wait": pl.PassTime,
	}
}

func setFace(a Arg) (string, interface{}) {
	return "setFace", a
}

func flagName(pl *Player) (string, interface{}) {
	return "flagName", map[string]interface{}{
		"round":  pl.GetRound(),
		"player": pl.Index,
	}
}

func trigg(cs *Cards) (string, interface{}) {
	if cs != nil {
		return "trigg", map[string]interface{}{
			"uniqs": cs.Uniqs(),
		}
	}
	return "trigg", map[string]interface{}{
		"uniqs": []int{},
	}
}
