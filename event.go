package ygo_core

func (ca *Card) Init() {
	ca.Empty()
	ca.original = *ca.baseOriginal
	ca.isValid = true
	ca.SetNotDirect()
	ca.RecoverSummoner()
	ca.registerNormal()
	ca.baseOriginal.Initialize.Call(ca)
}

func (ca *Card) registerNormal() {

	// 进入墓地和 移除时 卡牌翻面
	ca.AddEvent(InExtra, ca.SetFaceUpAttack)

	// 破坏
	ca.AddEvent(Destroy, func(c *Card) {
		ca.ToGrave()
		pl := ca.GetSummoner()
		r := Arg{"self": ca.ToUint()}
		if c != nil {
			r["rival"] = c.ToUint()
		}
		pl.MsgPub("msg.011", r)
	})

	// 战斗破坏
	ca.AddEvent(DestroyBeBattle, func(c *Card) {
		ca.Dispatch(Destroy, c)
	})

	// 规则破坏
	ca.AddEvent(DestroyBeRule, func(c *Card) {
		ca.ToGrave()
		pl := ca.GetSummoner()
		pl.MsgPub("msg.012", Arg{"self": ca.ToUint()})
	})

	// 花费
	ca.AddEvent(Cost, func() {
		ca.ToGrave()
		pl := ca.GetSummoner()
		pl.MsgPub("msg.013", Arg{"self": ca.ToUint()})
	})

	// 丢弃
	ca.AddEvent(Discard, func() {
		ca.ToGrave()
		pl := ca.GetSummoner()
		pl.MsgPub("msg.014", Arg{"self": ca.ToUint()})
	})

	// 使用完毕
	ca.AddEvent(Depleted, func() {
		ca.ToGrave()
		pl := ca.GetSummoner()
		pl.MsgPub("msg.015", Arg{"self": ca.ToUint()})
	})

	// 失效
	ca.AddEvent(Disabled, func() {
		ca.UnregisterAllGlobalListen()
		//pl := ca.GetSummoner()
		//pl.MsgPub("msg.016", Arg{"self": ca.ToUint()})
		ca.isValid = false
	})

	// 进入墓地和除外
	ca.AddEvent(InGrave, ca.SetFaceUpAttack)
	ca.AddEvent(InRemoved, ca.SetFaceUpAttack)

	// 被移除
	ca.AddEvent(Removed, func() {
		pl := ca.GetSummoner()
		ca.ToRemoved()
		pl.MsgPub("msg.017", Arg{"self": ca.ToUint()})
	})

	// 被抽到手牌
	ca.AddEvent(InHand, ca.Peek)

	// 覆盖
	ca.AddEvent(Use2, func() {
		ca.Dispatch(Cover)
	})
	// 使用
	ca.AddEvent(Use1, func() {
		ca.Dispatch(Onset)
	})

	e := func() {
		ca.Dispatch(Disabled)
	}

	if ca.IsMonster() {
		ca.AddEvent(InGrave, e)
		ca.AddEvent(InRemoved, e)
		ca.AddEvent(InSzone, e)
		ca.registerMonster()
		ca.AddEvent(InMzone, func() {
			ca.AddEvent(InHand, e)
			ca.AddEvent(InDeck, e)
		})
	} else if ca.IsMagicAndTrap() {
		ca.registerMagicAndTrap()
		ca.AddEvent(OutSzone, e)
	}
}

func (ca *Card) registerMagicAndTrap() {
	ca.Range(InHand, OutHand, Arg{
		// 代价 先覆盖
		Pay: func(s string) {
			if s != Onset && s != Cover {
				return
			}
			pl := ca.GetSummoner()
			ca.ToSzone()
			if ca.IsInSzone() {
				ca.SetFaceDownAttack()
				pl.Msg("021", Arg{"self": ca.ToUint()})
			} else {
				ca.StopOnce(s)
			}
		},
	})
}

// 注册一张魔法卡
func (ca *Card) registerMagic(e interface{}, only bool) {
	ca.RegisterPay(func(s string) {
		if s != UseMagic {
			return
		}
		ca.SetFaceUp()
		pl := ca.GetSummoner()
		pl.MsgPub("msg.022", Arg{"self": ca.ToUint()})
	})
	ca.AddEvent(Onset, func() {
		ca.Dispatch(UseMagic)
	})
	ca.AddEvent(Effect0, e)
	ca.AddEvent(UseMagic, func() {
		pl := ca.GetSummoner()
		if ca.IsValid() {
			ca.Dispatch(Effect0)
			pl.MsgPub("msg.023", Arg{"self": ca.ToUint()})
			if only {
				ca.Dispatch(Depleted)
			}
		} else {
			pl.MsgPub("msg.024", Arg{"self": ca.ToUint()})
		}
	})
}

// 注册一张不通常魔法卡
func (ca *Card) RegisterUnordinaryMagic(e interface{}) {
	ca.registerMagic(e, false)
}

// 注册一张通常魔法卡
func (ca *Card) RegisterOrdinaryMagic(e interface{}) {
	ca.registerMagic(e, true)
}

// 卡牌注册一个事件触发器 如果触发则发送给另一个事件
func (ca *Card) registerIgnitionSelector(event string, e interface{}, toevent string) {
	ca.RegisterGlobalListen(event, e)

	// 不能换成 onlyone 否则就无法多次注册触发器
	ca.AddEvent(Trigger, func() {

		// 注意 发动触发的事件时注销全部事件监听
		// 不然 发动一张 神之宣告 会把自己破坏掉
		ca.UnregisterAllGlobalListen()

		ca.Dispatch(toevent)
	}, toevent)
}

// 触发选择器
func (ca *Card) RegisterIgnitionSelector(event string, e interface{}) {
	ca.registerIgnitionSelector(event, e, Chain)
}

// 注册一张陷阱卡
func (ca *Card) registerTrap(event string, e interface{}, only bool) {
	ca.RegisterPay(func(s string) {
		if s != UseTrap {
			return
		}
		ca.SetFaceUp()
		pl := ca.GetSummoner()
		pl.MsgPub("msg.022", Arg{"self": ca.ToUint()})
	})
	ca.AddEvent(InSzone, func() {
		pl := ca.GetSummoner()
		pl.OnlyOnce(RoundEnd, func() {
			ca.registerIgnitionSelector(event, e, UseTrap)
		}, ca, event, e)
	}, event, e)
	ca.AddEvent(UseTrap, func() {
		pl := ca.GetSummoner()
		if ca.IsValid() {
			ca.Dispatch(Chain)
			pl.MsgPub("msg.023", Arg{"self": ca.ToUint()})
			if only {
				ca.Dispatch(Depleted)
			}
		} else {
			pl.MsgPub("msg.024", Arg{"self": ca.ToUint()})
		}
	})
}

// 注册一张不通常的陷阱卡
func (ca *Card) RegisterUnordinaryTrap(event string, e interface{}) {
	ca.registerTrap(event, e, false)
}

// 注册一张通常的陷阱卡
func (ca *Card) RegisterOrdinaryTrap(event string, e interface{}) {
	ca.registerTrap(event, e, true)
}

// 推送卡牌使之能连锁 玩家可以选择这张卡发动效果
func (ca *Card) PushChain(e interface{}) {
	yg := ca.GetSummoner().Game()
	yg.AddEvent(Chain, ca)
	ca.EmptyEvent(Chain)
	ca.AddEvent(Chain, e)
}

// 注册全局效果监听 直到注销
func (ca *Card) RegisterGlobalListen(event string, e interface{}) {
	yg := ca.GetSummoner().Game()
	yg.AddEvent(event, e, ca)
	ca.OnlyOnce(UnregisterAllGlobalListen, func() {
		yg.RemoveEvent(event, e, ca)
	}, event, e)
}

// 注销全局效果监听
func (ca *Card) UnregisterGlobalListen(event string, e interface{}) {
	yg := ca.GetSummoner().Game()
	yg.RemoveEvent(event, e, ca)
}

// 注销全部全局效果监听
func (ca *Card) UnregisterAllGlobalListen() {
	ca.Dispatch(UnregisterAllGlobalListen)
}

// 注册一个装备魔法卡  装备对象判断  装备上动作 装备下动作
func (ca *Card) RegisterEquipMagic(a Action, f1 interface{}, f2 interface{}) {
	ca.RegisterUnordinaryMagic(func() {
		pl := ca.GetSummoner()
		pl.MsgPub("msg.031", Arg{"self": ca.ToUint()})
		tar := pl.GetTarget()
		if c := pl.SelectForWarn(pl.Mzone(), tar.Mzone(), a); c != nil {

			// 装备卡 离开场地时
			ca.OnlyOnce(Disabled, func() {
				ca.Dispatch(Effect2, c)
			}, c)

			c.OnlyOnce(Disabled, func() {
				ca.Dispatch(Depleted)
			}, ca)

			// 监听目标的改变判断目标的改变是否允许
			c.AddEvent(Change, func() {
				if !c.IsInMzone() || !a.Call(c) {
					ca.Dispatch(Depleted)
				}
			}, ca)

			// 执行装备 上的效果
			ca.Dispatch(Effect1, c)
			pl.MsgPub("msg.032", Arg{"self": ca.ToUint()})
		} else {
			ca.Dispatch(DestroyBeRule)
			pl.MsgPub("msg.033", Arg{"self": ca.ToUint()})
		}
	})
	ca.AddEvent(Effect1, f1)
	ca.AddEvent(Effect2, f2)
}

func (ca *Card) RegisterPay(f interface{}) {
	ca.AddEvent(Pay, f)
}

// 注册翻转效果
func (ca *Card) RegisterFlip(f interface{}) {
	ca.AddEvent(Flip, f)
}

// 注册融合怪兽的融合材料
func (ca *Card) RegisterFusionMonster(names ...string) {
	h := map[string]int{}
	for _, v := range names {
		h[v]++
	}
	ca.Range(InExtra, OutExtra, Arg{
		Pay: func(s string) {
			if s != SummonFusion {
				return
			}
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
						tm := pl.SelectForWarn(is)
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

// 注册一张怪兽卡
func (ca *Card) registerMonster() {

	// 定义怪兽手牌默认事件是召唤
	ca.AddEvent(Onset, func() {
		ca.Dispatch(Summon)
	})

	// 召唤 特殊召唤 翻转召唤 设置卡片正面朝上攻击表示
	ca.RegisterPay(func(s string) {
		if s != SummonSpecial && s != SummonFlip && s != Summon {
			return
		}
		ca.SetFaceUpAttack()
	})

	// 翻转 设置卡片正面朝上
	ca.RegisterPay(func(s string) {
		if s != Flip {
			return
		}
		ca.SetFaceUp()
	})

	// 特殊召唤
	ca.AddEvent(SummonSpecial, func() {
		pl := ca.GetSummoner()
		ca.ToMzone()
		ca.SetNotCanChange()
		pl.MsgPub("msg.043", Arg{"self": ca.ToUint()})
	})

	e0 := func() {
		ca.Dispatch(FaceUp)
	}
	ca.AddEvent(Flip, e0)
	ca.AddEvent(Summon, e0)
	ca.AddEvent(SummonSpecial, e0)

	// 正面朝上时和改变属性时显示属性
	ca.AddEvent(Change, ca.ShowInfo)
	ca.AddEvent(FaceUp, ca.ShowInfo)
	// 手牌
	ca.Range(InHand, OutHand, Arg{
		// 代价
		Pay: func(s string) {
			if s != Summon && s != Cover {
				return
			}

			pl := ca.GetSummoner()
			pl.resetReplyTime()
			i := 0
			if ca.GetLevel() > 6 {
				i += 2
			} else if ca.GetLevel() > 4 {
				i += 1
			}
			if i != 0 {
				pl.MsgPub("msg.044", Arg{"self": ca.ToUint(), "size": i})
			}
			for k := 0; k < i; {
				if t := pl.SelectForWarn(pl.Mzone()); t != nil {
					t.Dispatch(Freedom, ca, &k)
				} else {
					ca.StopOnce(s)
					pl.MsgPub("msg.045", Arg{"self": ca.ToUint()})
					return
				}
			}
			pl.SetNotCanSummon()
		},
		// 召唤
		Summon: func() {
			if ca.IsValid() {
				pl := ca.GetSummoner()
				ca.ToMzone()
				ca.SetNotCanChange()
				pl.MsgPub("msg.046", Arg{"self": ca.ToUint()})
			}
		},
		// 覆盖
		Cover: func() {
			pl := ca.GetSummoner()
			if ca.IsInHand() {
				ca.ToMzone()
				ca.SetFaceDownDefense()
				ca.SetNotCanChange()
				pl.Msg("047", Arg{"self": ca.ToUint()})
			}
		},
	})

	ca.Range(InMzone, OutMzone, Arg{
		// 被解放
		Freedom: func(c *Card, i *int) {
			pl := ca.GetSummoner()
			if i != nil {
				*i++
			}
			ca.ToGrave()
			pl.MsgPub("msg.048", Arg{"self": ca.ToUint()})
		},
		// 改变表示形式
		Expression: func() {
			pl := ca.GetSummoner()
			if ca.IsCanChange() {
				if ca.IsFaceDownDefense() {
					ca.Dispatch(SummonFlip)
				} else if ca.IsFaceUpDefense() {
					ca.SetFaceUpAttack()
					pl.MsgPub("msg.051", Arg{"self": ca.ToUint()})
				} else if ca.IsFaceUpAttack() {
					ca.SetFaceUpDefense()
					pl.MsgPub("msg.052", Arg{"self": ca.ToUint()})
				} else {
					pl.Msg("053", Arg{"self": ca.ToUint()})
					return
				}
				ca.SetNotCanChange()
			} else {
				pl.Msg("053", Arg{"self": ca.ToUint()})
			}
		},

		// 翻转召唤
		SummonFlip: func() {
			pl := ca.GetSummoner()
			pl.MsgPub("msg.061", Arg{"self": ca.ToUint()})
			ca.Dispatch(Flip)
		},

		// 翻转
		Flip: func() {
			pl := ca.GetSummoner()
			pl.MsgPub("msg.062", Arg{"self": ca.ToUint()})
		},

		// 发出战斗宣言
		Declaration: func(c *Card) {
			pl := ca.GetSummoner()
			pl.callAll(impact(ca, c))
			if c != nil && c.IsPortrait() {
				c = nil

			}
			if c != nil {
				pl.MsgPub("msg.063", Arg{"self": ca.ToUint(), "rival": c.ToUint()})
			} else {
				pl.MsgPub("msg.063", Arg{"self": ca.ToUint()})
			}

			b := false
			if c != nil {
				b = c.IsFaceDown()
				ca.SetFaceUp()
				if ca.IsInMzone() && c.IsInMzone() {
					ca.Dispatch(DamageStep, c)
				}
			} else {
				if ca.IsInMzone() {
					ca.Dispatch(DamageStep)
				}
			}
			if b {
				c.Dispatch(Flip)
			}
		},
		Deduct: func(tar *Player, i int) {
			tar.ChangeHp(i)
		},
		// 战斗判定
		DamageStep: func(c *Card) {
			pl := ca.GetSummoner()
			if c != nil {

				tar := c.GetSummoner()
				pl.MsgPub("msg.064", Arg{"self": ca.ToUint(), "rival": c.ToUint()})
				c.Dispatch(BearAttack, ca)
				if c.IsAttack() {
					t := ca.GetAttack() - c.GetAttack()
					if t > 0 {
						ca.Dispatch(Deduct, tar, -t)
						c.Dispatch(DestroyBeBattle, ca)
					} else if t < 0 {
						c.Dispatch(Deduct, pl, t)
						ca.Dispatch(DestroyBeBattle, c)
					} else {
						c.Dispatch(DestroyBeBattle, ca)
						ca.Dispatch(DestroyBeBattle, c)
					}
				} else if c.IsDefense() {
					t := ca.GetAttack() - c.GetDefense()
					if t > 0 {
						c.Dispatch(DestroyBeBattle, ca)
					} else if t < 0 {
						c.Dispatch(Deduct, pl, t)
					}
				}
				ca.Dispatch(Fought, c)
				c.Dispatch(Fought, ca)
			} else {
				tar := pl.GetTarget()
				ca.Dispatch(Deduct, tar, -ca.GetAttack())
				pl.MsgPub("msg.065", Arg{"self": ca.ToUint()})
			}

			ca.SetNotCanAttack()
			ca.SetNotCanChange()
		},
	})
}
