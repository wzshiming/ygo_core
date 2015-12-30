package ygo_core

import "fmt"

func NewPortraitCardOriginal() *CardOriginal {
	return &CardOriginal{
		IsValid: true,
		Lt:      LT_None,
		Initialize: func(ca *Card) bool {

			ca.RangeGlobal(InPortrait, OutPortrait, Arg{
				Initiative: func() {
					pl := ca.GetSummoner()
					pl.DrawCard(5)
					pl.ChangeLp(8000)
				},
				First: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}

					pl.MsgPub("msg.001", nil)
					ca.RegisterGlobalListen(BP, func(tar *Player) {
						tar.Mzone().ForEach(func(c *Card) bool {
							c.SetNotCanAttack()
							return true
						})
					})
					ca.RegisterGlobalListen(RoundEnd, func() {
						ca.UnregisterAllGlobalListen()
					})
				},
				RoundBegin: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}
					pl.SetCanSummon()
					pl.Mzone().ForEach(func(c0 *Card) bool {
						c0.SetCanAttack()
						c0.SetCanChange()
						return true
					})
				},
				RoundEnd: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}
					tar := pl.GetTarget()
					e := func(c0 *Card) bool {
						c0.EquipFlash()
						return true
					}
					pl.Mzone().ForEach(e)
					tar.Mzone().ForEach(e)
				},
				DP: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}
					pl.DrawCard(1)
				},
				MP: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}
					if pl.phases == LP_Main1 {
						ca.PushChain(LO_BP, func() {
							pl.Skip(LP_Battle)
						})
						ca.PushChain(LO_EP, func() {
							pl.Skip(LP_End)
						})
					} else {
						ca.PushChain(LO_EP, func() {
							pl.Skip(LP_End)
						})
					}

				},
				BP: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}

					ca.PushChain(LO_MP, func() {
						pl.Skip(LP_Main2)

					})
					ca.PushChain(LO_EP, func() {
						pl.Skip(LP_End)
					})
				},
				EP: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}
					if i := pl.Hand().Len() - pl.maxSdi; i > 0 {
						pl.resetReplyTime()
						pl.Msg("103", nil)
						for k := 0; k != i; k++ {
							ca := pl.SelectRequired(LO_Discard, pl.Hand())
							if ca == nil {
								ca = pl.Hand().EndPop()
							}
							ca.Dispatch(Discard)
						}
					}
				},
			})
			return true
		},
	}
}

func (ca *Card) registerNormal() {
	if ca.GetType() == LT_None {
		return
	}

	// 进入额外和 移除时 卡牌翻面
	ca.AddEvent(InExtra, ca.SetFaceUpAttack)
	ca.AddEvent(InGrave, ca.SetFaceUpAttack)

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
	ca.AddEventSuf(Disabled, func() {
		ca.UnregisterAllGlobalListen()
		//pl := ca.GetSummoner()
		//pl.MsgPub("msg.016", Arg{"self": ca.ToUint()})
		//ca.isValid = false
	})

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

	if ca.GetType().IsTrap() {
		ca.AddEvent(Use1, func() {
			ca.Dispatch(Cover)
		})
	} else {
		ca.AddEvent(Use2, func() {
			ca.Dispatch(Cover)
		})
		ca.AddEvent(Use1, func() {
			ca.Dispatch(Onset)
		})
	}

	if ca.GetType().IsMonster() {
		ca.registerMonster()
		e := func() {
			lp := ca.GetLastPlace()
			if lp.IsMzone() || lp.IsHand() {
				ca.Dispatch(Disabled)
			}
		}

		ca.AddEvent(InGrave, e)
		ca.AddEvent(InRemoved, e)
		ca.AddEvent(InSzone, e)
		ca.AddEvent(InHand, e)
		ca.AddEvent(InDeck, e)

		if ca.GetType().IsExtra() {
			ca.AddEvent(InDeck, ca.ToExtra)
			ca.AddEvent(InHand, ca.ToExtra)
		}

	} else if ca.GetType().IsSpellField() {
		ca.registerSpellAndTrap()
		e := func() {
			ca.Dispatch(Disabled)
		}
		ca.AddEvent(OutField, e)
	} else if ca.GetType().IsSpellAndTrap() {
		ca.registerSpellAndTrap()
		e := func() {
			ca.Dispatch(Disabled)
		}
		ca.AddEvent(OutSzone, e)
	} else {
		Debug("registerNormal", ca)
	}
}

// 魔法卡陷阱卡 操作
func (ca *Card) registerSpellAndTrap() {
	// 先覆盖
	e1 := func(s string) {
		pl := ca.GetSummoner()
		if ca.GetType().IsSpellField() {
			ca.ToField()
		} else {
			ca.ToSzone()
		}

		if ca.IsInSzone() || ca.IsInField() {
			ca.SetFaceDownAttack()
			pl.Msg("021", Arg{"self": ca.ToUint()})
		} else {
			ca.StopOnce(s)
		}
	}

	// 发动先翻面
	e2 := func() {
		ca.SetFaceUp()
		pl := ca.GetSummoner()
		pl.MsgPub("msg.022", Arg{"self": ca.ToUint()})
	}

	if ca.GetType().IsTrap() {
		ca.Range(InHand, OutHand, Arg{
			Pre + Cover: e1,
		})

		ca.Range(InSzone, OutSzone, Arg{
			Pre + UseTrap: e2,
		})
		ca.RangeGlobal(InHand, OutHand, Arg{
			MP: func() {
				pl := ca.GetSummoner()
				if !pl.IsCurrent() {
					return
				}
				if pl.Szone().Len() >= 5 {
					return
				}

				ca.PushChain(LO_Cover, func() {
					ca.Dispatch(Cover)
				})
			},
		})

	} else if ca.GetType().IsSpell() {
		// 注册 使用 为 发动魔法卡

		ff := string(LL_Szone)
		if ca.GetType().IsSpellField() {
			ff = string(LL_Field)
		}

		ca.Range(In+ff, Out+ff, Arg{
			Pre + UseSpell: e2,
		})

		ca.Range(InHand, OutHand, Arg{
			// 代价 先覆盖
			Pre + UseSpell: func(s string) {
				e1(s)
				if ca.GetType().IsSpellField() {
					if ca.IsInField() {
						e2()
					}
				} else {
					if ca.IsInSzone() {
						e2()
					}
				}

			},
			Pre + Cover: e1,
		})

		ca.RangeGlobal(InHand, OutHand, Arg{
			MP: func() {
				pl := ca.GetSummoner()
				if !pl.IsCurrent() {
					return
				}
				if pl.Szone().Len() >= 5 {
					return
				}

				ca.DispatchLocal(CheckSpell)

				ca.PushChain(LO_Cover, func() {
					ca.Dispatch(Cover)
				})
			},
		})

		if !ca.GetType().IsSpellQuickPlay() {
			ca.RangeGlobal(In+ff, Out+ff, Arg{
				MP: func() {
					pl := ca.GetSummoner()
					if !pl.IsCurrent() {
						return
					}
					if !ca.IsFaceDown() {
						return
					}
					ca.DispatchLocal(CheckSpell)
				},
			})
		}
	} else {
		Debug("registerSpellAndTrap", ca)
	}

}

// 推送卡牌使之能连锁 玩家可以选择这张卡发动效果
func (ca *Card) PushChain(lo lo_type, e interface{}) {
	yg := ca.GetSummoner().Game()
	u := yg.GetEventUniq()
	if ca.operateUniq != u {
		ca.operate = []lo_type{}
		ca.operateUniq = u
	}
	ca.operate = append(ca.operate, lo)
	yg.AddEvent(Chain, ca)

	cn := Chain + fmt.Sprint(len(ca.operate))
	ca.EmptyEvent(cn)
	ca.AddEvent(cn, e)
}

//
func (ca *Card) PushConst() int {
	return len(ca.operate)
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

func (ca *Card) AddEventPre(event string, f interface{}, token ...interface{}) {
	ca.AddEvent(Pre+event, f, token...)
}

func (ca *Card) AddEventSuf(event string, f interface{}, token ...interface{}) {
	ca.AddEvent(Suf+event, f, token...)
}

func (ca *Card) AddEventUsed(event string, f interface{}, token ...interface{}) {
	ca.AddEvent(Used+event, f, token...)
}

// 注册翻转效果
func (ca *Card) RegisterFlip(f interface{}) {
	ca.AddEvent(Flip, f)
}

// 注册一张怪兽卡
func (ca *Card) registerMonster() {

	// 召唤 特殊召唤 翻转召唤 设置卡片正面朝上攻击表示 不变
	ca.AddEventPre(SummonSpecial, ca.SetFaceUpAttack)
	ca.AddEventPre(SummonFlip, ca.SetFaceUpAttack)
	ca.AddEventPre(Summon, ca.SetFaceUpAttack)
	ca.AddEventPre(Summon, ca.SetFaceUpAttack)

	// 翻转 设置卡片正面朝上 不变
	ca.AddEventPre(Flip, ca.SetFaceUp)

	// 面朝下 结束 效果 或者  离开场地时
	ca.AddEventSuf(OutMzone, func() {
		ca.Dispatch(FaceDown)
	})

	// 特殊召唤  不变
	ca.AddEvent(SummonSpecial, func() {
		pl := ca.GetSummoner()
		ca.ToMzone()
		ca.SetNotCanChange()
		pl.MsgPub("msg.043", Arg{"self": ca.ToUint()})
	})

	ca.AddEvent(InMzone, func() {
		if ca.IsFaceUp() {
			ca.Dispatch(FaceUp)
			ca.ShowInfo()
		}
	})

	e := func(s string) {

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
			if t := pl.SelectRequired(LO_Freedom, pl.Mzone()); t != nil {
				t.Dispatch(Freedom, ca, &k)
			} else {
				ca.StopOnce(s)
				pl.MsgPub("msg.045", Arg{"self": ca.ToUint()})
				return
			}
		}
		pl.SetNotCanSummon()
	}

	// 手牌
	ca.Range(InHand, OutHand, Arg{
		// 代价
		Pre + Summon: e,
		Pre + Cover:  e,
		// 召唤
		Summon: func() {
			pl := ca.GetSummoner()
			ca.ToMzone()
			ca.SetNotCanChange()
			pl.MsgPub("msg.046", Arg{"self": ca.ToUint()})
		},

		// 覆盖
		Cover: func() {
			pl := ca.GetSummoner()
			ca.ToMzone()
			ca.SetFaceDownDefense()
			ca.SetNotCanChange()
			pl.Msg("047", Arg{"self": ca.ToUint()})
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

		// 翻转召唤
		SummonFlip: func() {
			pl := ca.GetSummoner()
			pl.MsgPub("msg.061", Arg{"self": ca.ToUint()})
			//ca.Dispatch(Flip)
		},

		// 翻转
		Flip: func() {
			ca.ShowInfo()
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
				ca.setFaceUp()
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
			tar.ChangeLp(i)
		},
		// 战斗判定
		DamageStep: func(c *Card) {
			pl := ca.GetSummoner()
			if c != nil {

				tar := c.GetSummoner()
				pl.MsgPub("msg.064", Arg{"self": ca.ToUint(), "rival": c.ToUint()})
				c.Dispatch(BearAttack, ca)
				if c.IsAttack() {
					t := ca.GetAtk() - c.GetAtk()
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
					t := ca.GetAtk() - c.GetDef()
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
				ca.Dispatch(Deduct, tar, -ca.GetAtk())
				pl.MsgPub("msg.065", Arg{"self": ca.ToUint()})
			}

			ca.SetNotCanAttack()
			ca.SetNotCanChange()
		},
	})

	// 在手牌的事件
	ca.RangeGlobal(InHand, OutHand, Arg{
		MP: func() {
			pl := ca.GetSummoner()
			if !pl.IsCurrent() {
				return
			}
			if !pl.IsCanSummon() {
				return
			}
			if ca.GetLevel() <= 4 {
				ca.PushChain(LO_Summon, func() {
					ca.Dispatch(Summon)
				})
				ca.PushChain(LO_Cover, func() {
					ca.Dispatch(Cover)
				})
			} else {
				if ca.GetLevel() == 5 || ca.GetLevel() == 6 {
					if pl.Mzone().Len() == 0 {
						return
					}
				} else {
					if pl.Mzone().Len() <= 1 {
						return
					}
				}
				ca.PushChain(LO_SummonFreedom, func() {
					ca.Dispatch(Summon)
				})
				ca.PushChain(LO_CoverFreedom, func() {
					ca.Dispatch(Cover)
				})
			}
		},
	})

	// 在怪兽区事件
	ca.RangeGlobal(InMzone, OutMzone, Arg{
		MP: func() {
			pl := ca.GetSummoner()
			if !pl.IsCurrent() {
				return
			}
			if !ca.IsCanChange() {
				return
			}
			if pl.Mzone().Len() >= 5 {
				return
			}

			if ca.IsFaceDownDefense() {
				ca.PushChain(LO_SummonFlip, func() {
					ca.Dispatch(SummonFlip)
					ca.SetNotCanChange()
				})
			} else if ca.IsFaceUpDefense() {
				ca.PushChain(LO_SetAttack, func() {
					ca.SetFaceUpAttack()
					pl.MsgPub("msg.051", Arg{"self": ca.ToUint()})
					ca.SetNotCanChange()
				})

			} else if ca.IsFaceUpAttack() {
				ca.PushChain(LO_SetDefense, func() {
					ca.SetFaceUpDefense()
					pl.MsgPub("msg.051", Arg{"self": ca.ToUint()})
					ca.SetNotCanChange()
				})
			}
		},
		BP: func() {
			pl := ca.GetSummoner()
			if !pl.IsCurrent() {
				return
			}
			if !ca.IsCanAttack() {
				return
			}
			if !ca.IsFaceUpAttack() {
				return
			}
			ca.PushChain(LO_Attack, func() {
				tar := pl.GetTarget()
				if tar.Mzone().Len() == 0 {
					c := pl.SelectRequiredShor(LO_Target, tar.Portrait())
					ca.Dispatch(Declaration, c)
				} else {
					if ca.IsCanDirect() {
						c := pl.SelectRequiredShor(LO_Target, tar.Mzone(), tar.Portrait())
						ca.Dispatch(Declaration, c)
					} else {
						c := pl.SelectRequiredShor(LO_Target, tar.Mzone())
						ca.Dispatch(Declaration, c)
					}
				}
			})
		},
	})
}
