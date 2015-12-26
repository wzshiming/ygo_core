package ygo_core

// 魔法卡 定义事件

func (ca *Card) PushSpell(lo lo_type, e interface{}) {
	ca.PushChain(lo, func() {
		ca.Dispatch(UseSpell)
	})
	ca.EmptyEvent(UseSpell)
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

// 装备魔法  有发动条件 需要PushSpellEquip
func (ca *Card) registerSpellEquip(e interface{}) {
	ca.RegisterSpellUnnormal(e)
}

// 装备效果注册
func (ca *Card) effectEquipRegister(e func(*Card)) {
	ca.AddEvent(equipEffect, e)
}

// 装备刷新
func (ca *Card) EquipFlash() {
	ca.appendOriginal = *ca.baseOriginal
	ca.ForEventEach(equipList, func(s string, i interface{}) {
		if v, ok := i.(*Card); ok {
			v.Dispatch(equipEffect, ca)
		} /*else {
			ca.OnlyOnce("", i)
			ca.DispatchLocal("", ca)
		}*/
	})
	if ca.IsFaceUp() && ca.IsInMzone() {
		ca.ShowInfo()
	}
}

// 给卡牌 添加 简易的 装备效果
func (ca *Card) effectEquipSimple(c *Card) {
	ca.Range("", usedOver, Arg{
		equipTarget: c,
	})

	ca.RangeForOther(c, "", usedOver, Arg{
		equipList: ca,
	})
	c.EquipFlash()
}

// 简易的 装备效果 结束
func (ca *Card) effectEquipSimpleOver() {
	cs := NewCards()
	ca.ForEventEach(equipTarget, func(s string, i interface{}) {
		if c, ok := i.(*Card); ok {
			cs.EndPush(c)
		}
	})
	ca.Dispatch(usedOver)
	cs.ForEach(func(c *Card) bool {
		c.EquipFlash()
		return true
	})
}

// 刷新装备对象
func (ca *Card) EquipTargetFlash() {
	ca.ForEventEach(equipTarget, func(s string, i interface{}) {
		if c, ok := i.(*Card); ok {
			c.EquipFlash()
		}
	})
}

// 给卡牌 添加 在场上的 装备效果
func (ca *Card) effectEquipBind(c *Card) {
	// 注册 ca 直到失效时
	ca.Range("", Disabled, Arg{
		equipTarget: c,
		// 丢失目标时
		equipMissed: func() {
			if ca.EventSize(equipTarget) == 0 {
				ca.Depleted(c)
			}
		},
	})

	// 注册 ca 直到失效时 c的事件
	ca.RangeForOther(c, "", Disabled, Arg{
		equipList: ca,
		// c离开场地时  注意 现在这种离开怪兽区的写法 对改变控制者 也会失效
		Suf + OutMzone: func() {
			ca.RemoveEvent(equipTarget, c)
			ca.Dispatch(equipMissed)
		},
	})
	c.EquipFlash()
}

// 装备 默认普通装备卡流程
func (ca *Card) equipBind(c *Card, e func(*Card)) {
	ca.effectEquipRegister(e)
	ca.effectEquipBind(c)
}

// 装备 简易的装备流程
func (ca *Card) equipSimple(c *Card, e func(*Card)) {
	ca.effectEquipRegister(e)
	ca.effectEquipSimple(c)
}

// 光环 场上所有符合条件的目标
func (ca *Card) effectMzoneHalo(e func(*Card), b bool) {
	ca.effectEquipRegister(e)

	// 给卡牌添加 简单装备
	e0 := func(c *Card) bool {
		ca.effectEquipSimple(c)
		return true
	}
	pl := ca.GetSummoner()
	tar := pl.GetTarget()
	/// 对已经在场上的卡生效
	pl.Mzone().ForEach(e0)
	tar.Mzone().ForEach(e0)

	if b {
		// 对后加入的卡牌也生效
		ca.RangeGlobal("", usedOver, Arg{
			Pre + InMzone: func(c *Card) {
				ca.effectEquipSimple(c)
			},
		})
	}
}

// 单个怪兽 添加效果 直到某一个事件
func (ca *Card) EquipSimpleTemp(c *Card, eve string, e func(*Card)) {
	ca.equipSimple(c, e)
	ca.RangeGlobal("", usedOver, Arg{
		eve: func() {
			ca.effectEquipSimpleOver()
		},
	})
}

// 光环 场上所有符合条件的目标
func (ca *Card) EffectMzoneHalo(e func(*Card)) {
	ca.effectMzoneHalo(e, true)
}

// 光环 场上所有符合条件的目标 直到自身被破坏
func (ca *Card) EffectExistMzoneHalo(e func(*Card)) {
	ca.EffectMzoneHalo(e)

	ca.Range("", usedOver, Arg{
		Disabled: func() {
			ca.effectEquipSimpleOver()
		},
	})
}

// 光环 场上所有符合条件的目标 直到某一个事件
func (ca *Card) EffectTempMzoneHalo(eve string, e func(*Card)) {
	ca.EffectMzoneHalo(e)
	ca.RangeGlobal("", usedOver, Arg{
		eve: func() {
			ca.effectEquipSimpleOver()
		},
	})
}

// 装备效果 根据自己某个地方的牌的数量 刷新自身
func (ca *Card) effectAreaSize(c *Card, ll ll_type, b bool, a Action, f func(i int, c0 *Card)) {
	ca.RangeGlobal("", Disabled, Arg{
		Suf + In + string(ll):  c.EquipFlash,
		Suf + Out + string(ll): c.EquipFlash,
	})
	ca.equipSimple(c, func(c0 *Card) {
		pl := c.GetSummoner()
		i := 0
		e := func(c0 *Card) bool {
			if a.Call(c0) {
				i++
			}
			return true
		}
		pl.getArea(ll).ForEach(e)
		if b {
			tar := pl.GetTarget()
			tar.getArea(ll).ForEach(e)
		}
		f(i, c0)
	})
}

// 装备魔法卡的效果
func (ca *Card) registerSpellEquipEffect(a Action, e func(*Card), b func(*Card)) {
	// 这是一张装备魔法
	ca.registerSpellEquip(func() {
		pl := ca.GetSummoner()
		tar := pl.GetTarget()

		// 场上要有符合要求的怪
		css := NewCards(pl.Mzone(), tar.Mzone(), func(c0 *Card) bool {
			return c0.IsFaceUp() && a.Call(c0)
		})
		if css.Len() != 0 {

			// 判断该卡是否还符合要求
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
					ca.equipBind(c, e0)
					if b != nil {
						b(c)
					}
				}
			})
		}
	})
}

// 装备魔法  封装的效果 效果增幅固定
func (ca *Card) RegisterSpellEquipEffect1(a Action, e func(*Card)) {
	ca.registerSpellEquipEffect(a, e, nil)
}

// 装备魔法  封装的效果 效果增幅在某个事件刷新
func (ca *Card) RegisterSpellEquipEffect2(a Action, e func(*Card), event string, eve func(*Card)) {
	ca.registerSpellEquipEffect(a, e, func(c *Card) {
		ca.RangeGlobal("", Disabled, Arg{
			event: func() {
				eve(c)
				c.EquipFlash()
			},
		})
	})
}

// 通常魔法 封装效果 场上符合条件类型的怪兽破坏
func (ca *Card) RegisterSpellEffectDestroyFor(a Action) {
	ca.RegisterSpellNormalPuah(LO_DestroyMonster, func() {
		pl := ca.GetSummoner()
		tar := pl.GetTarget()
		css := NewCards(pl.Mzone(), tar.Mzone(), func(c0 *Card) bool {
			return a.Call(c0)
		})
		css.ForEach(func(c0 *Card) bool {
			c0.Destroy(ca)
			return true
		})

	})
}

// 陷阱卡

// 推送陷阱卡
func (ca *Card) PushTrap(lo lo_type, e interface{}) {
	ca.PushChain(lo, func() {
		ca.Dispatch(UseTrap)
	})
	ca.EmptyEvent(UseTrap)
	ca.AddEvent(UseTrap, e)
}

// 不是通常陷阱卡 使用完不送墓地  有发动条件 需要PushTrap
func (ca *Card) RegisterTrapUnnormal(e interface{}, events ...string) {
	ca.AddEvent(InSzone, func() {
		pl := ca.GetSummoner()
		//注册 下回合才能 连锁事件
		pl.OnlyOnce(RoundEnd, func() {
			if !ca.IsInSzone() {
				return
			}
			ar := Arg{}
			for _, v := range events {
				ar[v] = e
			}
			ca.RangeGlobal("", OutSzone, ar)

		}, ca, e)
	}, e)
}
func (ca *Card) RegisterTrapUnnormalPush(lo lo_type, e interface{}, events ...string) {
	ca.RegisterTrapUnnormal(func() {
		ca.PushTrap(lo, e)
	}, events...)
}

// 通常陷阱 使用完就送墓地  有发动条件 需要PushTrap
func (ca *Card) RegisterTrapNormal(e interface{}, events ...string) {
	ca.RegisterTrapUnnormal(e, events...)
	ca.AddEventUsed(UseTrap, func() {
		ca.Depleted(ca)
	})
}

func (ca *Card) RegisterTrapNormalPush(lo lo_type, e interface{}, events ...string) {
	ca.RegisterTrapNormal(func() {
		ca.PushTrap(lo, e)
	}, events...)
}

// 不是通常陷阱卡 使用完不送墓地  没有发动条件 需要PushTrap
func (ca *Card) RegisterTrapUnnormalAny(e interface{}) {
	ca.AddEvent(InSzone, func() {
		pl := ca.GetSummoner()
		//注册 下回合才能 连锁事件
		pl.OnlyOnce(RoundEnd, func() {
			if !ca.IsInSzone() {
				return
			}
			yg := pl.Game()
			yg.OnAny(ca)
			ca.AddEvent(OutSzone, func() {
				yg.OffAny(ca)
			})
		}, ca, e)
	}, e)
}

func (ca *Card) RegisterTrapUnnormalAnyPush(lo lo_type, e interface{}) {
	ca.RegisterTrapUnnormalAny(func() {
		ca.PushTrap(lo, e)
	})
}

// 通常陷阱 使用完就送墓地  没有发动条件 需要PushTrap
func (ca *Card) RegisterTrapNormalAny(e interface{}) {
	ca.RegisterTrapUnnormalAny(e)
	ca.AddEventUsed(UseTrap, func() {
		ca.Depleted(ca)
	})
}

func (ca *Card) RegisterTrapNormalAnyPush(lo lo_type, e interface{}) {
	ca.RegisterTrapNormalAny(func() {
		ca.PushTrap(lo, e)
	})
}

// 注册融合怪兽的融合材料
func (ca *Card) RegisterMonsterFusion(names ...string) {
	h := map[string]int{}
	for _, v := range names {
		h[v]++
	}
	ca.Range(InExtra, OutExtra, Arg{
		Pre + SummonFusion: func(s string) {
			pl := ca.GetSummoner()
			se := NewCards()
			cs := NewCards(pl.Hand(), pl.Mzone())
			for k, v := range h {
				is := cs.Find(func(c *Card) bool {
					return c.GetName() == k
				})
				if is.Len() == v {
					is.ForEach(func(c *Card) bool {
						se.EndPush(c)
						return true
					})
				} else if is.Len() > v {
					for i := 0; i != v; i++ {
						tm := pl.SelectRequired(LO_Fusion, is)
						if tm == nil {
							pl.MsgPub("msg.041", Arg{"self": ca.ToUint()})
							ca.StopOnce(s)
							return
						}
						is.PickedFor(tm)
						se.EndPush(tm)
					}
				} else {
					pl.MsgPub("msg.042", Arg{"self": ca.ToUint()})
					ca.StopOnce(s)
					return
				}
			}
			se.ForEach(func(c *Card) bool {
				c.Dispatch(Cost)
				return true
			})
		},
		SummonFusion: func() {
			ca.Dispatch(SummonSpecial)
		},
	})
}

// 怪兽效果 封装 当从怪兽区送往墓地时 从卡组选一张符合条件的卡加入手牌
func (ca *Card) RegisterMonsterEffect1(a Action) {
	ca.AddEvent(InGrave, func() {
		pl := ca.GetSummoner()
		if ca.GetLastPlace() != pl.Mzone() {
			return
		}
		css := NewCards(pl.Deck(), func(c0 *Card) bool {
			return a.Call(c0)
		})
		if css.Len() != 0 {
			if c := pl.SelectRequiredShor(LO_JoinHand, css); c != nil {
				c.ToHand()
			}
		}
	})
}

// 怪兽效果 封装 对全场怪兽进行光环
func (ca *Card) RegisterMonsterEffect2(e func(*Card)) {
	e0 := func() {
		ca.effectMzoneHalo(e, true)
	}
	e1 := func() {
		ca.effectEquipSimpleOver()
	}
	ca.AddEvent(FaceUp, e0)
	ca.AddEvent(OutMzone, e1)
}

func (ca *Card) RegisterMonsterFaceUp(e interface{}) {
	ca.AddEvent(FaceUp, e)
}

// 怪兽效果 封装 根据自己的某个地方的牌的数量 刷新自身
func (ca *Card) RegisterMonsterSelfAreaSize(ll ll_type, a Action, f func(i int, c0 *Card)) {
	ca.AddEvent(FaceUp, func() {
		ca.effectAreaSize(ca, ll, false, a, f)
	})
}

// 怪兽效果 封装 根据自己的某个地方的牌的数量 刷新自身
func (ca *Card) RegisterMonsterAllAreaSize(ll ll_type, a Action, f func(i int, c0 *Card)) {
	ca.AddEvent(FaceUp, func() {
		ca.effectAreaSize(ca, ll, true, a, f)
	})
}

// 控制权变更
func (ca *Card) ControlPower(p *Player) {
	ca.SetSummoner(p)
	ca.ToMzone()
}

// 控制权恢复
func (ca *Card) ControlRestore() {
	ca.ControlPower(ca.GetOwner())
}
