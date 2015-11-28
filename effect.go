package ygo_core

// 怪兽区域光环效果生成
func (ca *Card) EffectMzoneHalo(a Action, f0 interface{}, f1 interface{}) func() {
	ca.AddEvent(Effect0, f0)
	ca.AddEvent(Effect1, f1)
	e0 := func(c *Card) bool {
		if a.Call(c) {
			ca.Dispatch(Effect0, c)
			ca.AddEvent(OutMzone, func() {
				ca.Dispatch(Effect1, c)
			}, c)
		}
		return true
	}

	e := func() {
		pl := ca.GetSummoner()
		tar := pl.GetTarget()
		cs := NewCards(tar.Mzone(), pl.Mzone())
		cs.ForEach(e0)
		ca.RegisterGlobalListen(InMzone, e0)
	}
	return e
}

// 出入区域效果生成
func (ca *Card) EffectAccessArea(area ll_type, a Action, f0 interface{}, f1 interface{}) func() {
	inflag := "inEffectAccessArea"
	outflag := "outEffectAccessArea"
	ca.AddEvent(inflag, f0)
	ca.AddEvent(outflag, f1)
	e0 := func(c *Card) bool {
		if a.Call(c) {
			ca.Dispatch(inflag, c)
		}
		return true
	}
	e1 := func(c *Card) bool {
		if a.Call(c) {
			ca.Dispatch(outflag, c)
		}
		return true
	}

	if area == LL_Mzone {
		return func() {
			pl := ca.GetSummoner()
			tar := pl.GetTarget()
			cs := NewCards(tar.Mzone(), pl.Mzone())
			cs.ForEach(e0)
			ca.RegisterGlobalListen(FaceUp, e0)
			ca.RegisterGlobalListen(OutMzone, e1)
		}
	}

	return func() {
		pl := ca.GetSummoner()
		ca.RegisterGlobalListen(In+string(area), e0)
		ca.RegisterGlobalListen(Out+string(area), e1)
		pl.getArea(area).ForEach(e0)
	}

}

// 怪兽效果 全场怪兽区域 类似光环效果 全场增幅
func (ca *Card) RegisterAllMzoneHalo(a Action, f0 interface{}, f1 interface{}) {
	ca.AddEvent(FaceUp, ca.EffectMzoneHalo(a, f0, f1))
}

// 怪兽效果 区域数量 某个区域存在符合条件的卡牌
func (ca *Card) RegisterAccessArea(area ll_type, a Action, f0 interface{}, f1 interface{}) {
	ca.AddEvent(FaceUp, ca.EffectAccessArea(area, a, f0, f1))
}
