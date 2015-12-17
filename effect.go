package ygo_core

// 魔法卡 定义事件

func (ca *Card) PushSpell(lo lo_type, e interface{}) {
	ca.PushChain(lo, func() {
		ca.Dispatch(UseSpell)
	})
	ca.AddEvent(UseSpell, e)
}

// 不是通常魔法 使用完不送墓地  有发动条件 需要PushSpell
func (ca *Card) RegisterSpellUnnormal(e interface{}) {
	ca.AddEvent(CheckSpell, e)
}

// 不是通常魔法 使用完不送墓地  没有发动条件
func (ca *Card) RegisterSpellUnnormalPush(lo lo_type, e interface{}) {
	ca.RegisterSpellUnnormal(func() {
		ca.PushSpell(lo, e)
	})
}

// 通常魔法 使用完就送墓地  有发动条件 需要PushSpell
func (ca *Card) RegisterSpellNormal(e interface{}) {
	ca.RegisterSpellUnnormal(e)
	ca.AddEventUsed(UseSpell, func() {
		ca.Depleted(ca)
	})
}

// 通常魔法 使用完就送墓地  没有发动条件
func (ca *Card) RegisterSpellNormalPuah(lo lo_type, e interface{}) {
	ca.RegisterSpellNormal(func() {
		ca.PushSpell(lo, e)
	})
}

//func (ca *Card) PushSpellEquip(css *Cards, e0 interface{}) {
//	flag := "pushSpellEquip"
//	ca.PushSpell(LO_Equip, func() {
//		pl := ca.GetSummoner()
//		if c := pl.SelectRequired(LO_Equip, css); c != nil {

//			ca.Range(flag, Disabled, Arg{
//				Equip:       e0,
//				EquipTarget: c,
//				EquipMissed: func() {
//					if ca.EventSize(EquipTarget) == 0 {
//						ca.Depleted(c)
//					}
//				},
//			})

//			ca.RangeForOther(c, flag, Disabled, Arg{
//				EquipList: ca,
//				OutMzone: func() {
//					ca.RemoveEvent(EquipTarget, c)
//					ca.Dispatch(EquipMissed)
//				},
//			})
//			ca.Dispatch(flag)
//			c.Dispatch(EquipFlash)
//		}

//	})
//}

// 装备魔法  有发动条件 需要PushSpellEquip
func (ca *Card) RegisterSpellEquip(e interface{}) {
	ca.RegisterSpellUnnormal(e)
}

func (ca *Card) registerSpellEquipEffect(a Action, e func(*Card), b func(*Card)) {
	ca.RegisterSpellEquip(func() {
		pl := ca.GetSummoner()
		tar := pl.GetTarget()
		css := NewCards(pl.Mzone(), tar.Mzone(), func(c0 *Card) bool {
			return c0.IsFaceUp() && a.Call(c0)
		})
		if css.Len() != 0 {
			e0 := func(c0 *Card) {
				if a.Call(c0) {
					e(c0)
				} else {
					ca.Destroy(c0)
				}
			}

			ca.PushSpell(LO_Equip, func() {
				pl := ca.GetSummoner()
				if c := pl.SelectRequired(LO_Equip, css); c != nil {

					ca.Range("", Disabled, Arg{
						Equip:       e0,
						EquipTarget: c,
						EquipMissed: func() {
							if ca.EventSize(EquipTarget) == 0 {
								ca.Depleted(c)
							}
						},
					})

					ca.RangeForOther(c, "", Disabled, Arg{
						EquipList: ca,
						OutMzone: func() {
							ca.RemoveEvent(EquipTarget, c)
							ca.Dispatch(EquipMissed)
						},
					})

					if b != nil {
						b(c)
					}
					c.Dispatch(EquipFlash)
				}

			})
		}
	})
}

// 装备魔法  封装的效果1  条件判断 和 装备效果
func (ca *Card) RegisterSpellEquipEffect1(a Action, e func(*Card)) {
	ca.registerSpellEquipEffect(a, e, nil)
}

// 装备魔法  封装的效果2  条件判断 和 装备效果
func (ca *Card) RegisterSpellEquipEffect2(a Action, e func(*Card), event string, eve func(*Card)) {
	ca.registerSpellEquipEffect(a, e, func(c *Card) {
		ca.RangeGlobal("", Disabled, Arg{
			event: func() {
				eve(c)
				c.Dispatch(EquipFlash)
			},
		})
	})
}

// 控制权变更
func (ca *Card) MzoneControlPower(p *Player) {
	ca.SetSummoner(p)
	ca.ToMzone()
}

// 控制权恢复
func (ca *Card) MzoneControlRestore() {
	ca.MzoneControlPower(ca.GetOwner())
}
