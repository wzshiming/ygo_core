package ygo_core

import (
	"github.com/wzshiming/base"
	"github.com/wzshiming/dispatcher"
	"github.com/wzshiming/ffmt"
)

type Card struct {
	dispatcher.Events
	base.Unique

	isValid        bool
	appendOriginal CardOriginal
	baseOriginal   *CardOriginal
	place          *Group  // 所在位置
	lastPlace      *Group  // 上一个位置
	summoner       *Player // 召唤者
	owner          *Player // 所有者
	le             le_type // 表示形式
	//怪兽卡 属性

	lastAttackRound int  // 最后攻击回合
	lastChangeRound int  // 最后改变表示形式回合
	direct          bool // 直接攻击玩家

	operateUniq uint
	operate     []lo_type

	lastEvent string
	Counter   int // 计数器 给卡牌效果操作
}

func (ca *Card) String() string {
	if ca.GetLevel() == 0 {
		return ffmt.Sputs(map[string]interface{}{
			"Base":  *ca.baseOriginal,
			"Pos":   ca.place,
			"Expre": ca.le,
		})
	}
	return ffmt.Sputs(map[string]interface{}{
		"Curr":  ca.appendOriginal,
		"Pos":   ca.place,
		"Expre": ca.le,
	})
}

func (ca *Card) Init() {
	ca.Empty()
	ca.appendOriginal = *ca.baseOriginal
	ca.Counter = 0
	ca.isValid = true
	ca.SetNotDirect()
	ca.SetCanAttack()
	ca.SetCanChange()
	ca.RecoverSummoner()
	ca.registerNormal()
	ca.baseOriginal.Initialize.Call(ca)
}

func (ca *Card) StopEvent() {
	ca.StopOnce(ca.lastEvent)
}

func (ca *Card) RangeGlobal(eventin string, eventout string, events map[string]interface{}) {
	yg := ca.GetSummoner().Game()
	ca.RangeForOther(yg, eventin, eventout, events)
}

func (ca *Card) Peek() {
	pl := ca.GetSummoner()
	pl.call(setFront(ca))
	pl.call(exprCard(ca, LE_FaceUpAttack))
	ca.le |= LE_peek
	pl.GetTarget().call(exprCard(ca, LE_FaceDownAttack))
}

func (ca *Card) PeekFor(p *Player) {
	p.call(setFront(ca))
}

func (ca *Card) IsCanDirect() bool {
	return ca.direct
}
func (ca *Card) SetCanDirect() {
	ca.direct = true
}

func (ca *Card) SetNotDirect() {
	ca.direct = false
}

func (ca *Card) ShowInfo() {
	pl := ca.GetSummoner()
	pl.callAll(setCardFace(ca, Arg{"ATK": ca.GetAtk(), "DEF": ca.GetDef()}))
}

func (ca *Card) HideInfo() {
	pl := ca.GetSummoner()
	pl.callAll(setCardFace(ca, Arg{}))
}

func (ca *Card) IsValid() bool {
	return ca.isValid
}

func (ca *Card) Priority() int {
	switch ca.baseOriginal.Lt {
	case LT_SpellQuickPlay: //速攻魔法 速度2
		return 2
	case LT_TrapNormal: //普通陷阱 速度2
		return 2
	case LT_TrapContinuous: //永续陷阱 速度2
		return 2
	case LT_TrapCounter: //反击陷阱 速度3
		return 3
	default:
		return 1
	}
	return 0
}

func (ca *Card) DispatchLocal(eventName string, args ...interface{}) {
	ca.Events.Dispatch(eventName, args...)
}

func (ca *Card) DispatchGlobal(eventName string, args ...interface{}) {
	if ca.Events.IsOpen(eventName) {
		pl := ca.GetSummoner()
		yg := pl.Game()
		yg.chain(eventName, ca, pl, args)
	}
	ca.Events.Dispatch(eventName, args...)
}

func (ca *Card) Dispatch(eventName string, args ...interface{}) {
	//defer DebugStack()
	yg := ca.GetSummoner().Game()
	if !ca.Events.IsOpen(eventName) {
		return
	}
	ca.DispatchGlobal(Pre+eventName, eventName)
	if ca.Events.IsOpen(eventName) {
		yg.chain(eventName, ca, ca.GetSummoner(), args)
	}
	if ca.Events.IsOpen(eventName) {
		ca.Events.Dispatch(eventName, args...)
		ca.DispatchGlobal(Used+eventName, eventName)
	}
	ca.DispatchGlobal(Suf+eventName, eventName)
}

func (ca *Card) GetPlace() *Group {
	return ca.place
}

func (ca *Card) GetLastPlace() *Group {
	return ca.lastPlace
}

// 设置名字
func (ca *Card) SetName(name string) {
	ca.appendOriginal.Name = name
}

// 获得名字
func (ca *Card) GetName() string {
	return ca.appendOriginal.Name
}

// 获得召唤者
func (ca *Card) GetSummoner() *Player {
	return ca.summoner
}

// 设置召唤者
func (ca *Card) SetSummoner(c *Player) {
	ca.IsCanAttack()
	ca.summoner = c
}

// 恢复召唤者
func (ca *Card) RecoverSummoner() {
	ca.summoner = ca.GetOwner()
}

// 获得所有者
func (ca *Card) GetOwner() *Player {
	return ca.owner
}

// 获得 id
func (ca *Card) GetId() uint {
	return ca.appendOriginal.Id
}

// 获得基础类型
func (ca *Card) GetBaseType() lt_type {
	return ca.baseOriginal.Lt
}

func (ca *Card) Is(a ...interface{}) bool {
	for _, v := range a {
		switch s := v.(type) {
		case lt_type:
			if ca.appendOriginal.Lt&s == 0 {
				return false
			}
		case la_type:
			if ca.appendOriginal.La&s == 0 {
				return false
			}
		case le_type:
			if ca.le&s == 0 {
				return false
			}
		case lr_type:
			if ca.appendOriginal.Lr&s == 0 {
				return false
			}
		case ll_type:
			if ca.place.name != s {
				return false
			}
		case *Player:
			if ca.GetSummoner() != s {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// 获得类型
func (ca *Card) GetType() lt_type {
	return ca.appendOriginal.Lt
}

// 设置类型
func (ca *Card) SetType(l lt_type) {
	ca.appendOriginal.Lt = l
}

// 获得基础属性
func (ca *Card) GetBaseAttribute() la_type {
	return ca.baseOriginal.La
}

// 获得属性
func (ca *Card) GetAttr() la_type {
	return ca.appendOriginal.La
}

//  设置属性
func (ca *Card) SetAttr(l la_type) {
	ca.appendOriginal.La = l

}

// 获得基础种族
func (ca *Card) GetBaseRace() lr_type {
	return ca.baseOriginal.Lr
}

// 获得种族
func (ca *Card) GetRace() lr_type {
	return ca.appendOriginal.Lr
}

// 设置种族
func (ca *Card) SetRace(l lr_type) {
	ca.appendOriginal.Lr = l

}

// 获得基础攻击值
func (ca *Card) GetBaseAtk() int {
	return ca.baseOriginal.Atk
}

// 获得攻击值
func (ca *Card) GetAtk() int {
	return ca.appendOriginal.Atk
}

// 设置攻击值
func (ca *Card) SetAtk(i int) {
	if i < 0 {
		i = 0
	}
	ca.appendOriginal.Atk = i

}

// 改变攻击值
func (ca *Card) AddAtk(i int) {
	ca.SetAtk(ca.GetAtk() + i)
}

// 获得基础防御值
func (ca *Card) GetBaseDef() int {
	return ca.baseOriginal.Def
}

// 获得防御值
func (ca *Card) GetDef() int {
	return ca.appendOriginal.Def
}

// 设置防御值
func (ca *Card) SetDef(i int) {
	if i < 0 {
		i = 0
	}
	ca.appendOriginal.Def = i

}

// 改变防御值
func (ca *Card) AddDef(i int) {
	ca.SetDef(ca.GetDef() + i)
}

// 获得基础等级
func (ca *Card) GetBaseLevel() int {
	return ca.baseOriginal.Level
}

// 获得等级
func (ca *Card) GetLevel() int {
	return ca.appendOriginal.Level
}

// 设置等级
func (ca *Card) SetLevel(i int) {
	ca.appendOriginal.Level = i

}

// 判断能够改变表示形式
func (ca *Card) IsCanChange() bool {
	return ca.lastChangeRound != 0
}

// 设置能够改变表示形式
func (ca *Card) SetCanChange() {
	ca.lastChangeRound = 1

}

// 设置不能够改变表示形式
func (ca *Card) SetNotCanChange() {
	ca.lastChangeRound = 0
}

// 判断能够攻击
func (ca *Card) IsCanAttack() bool {
	return ca.lastAttackRound != 0
}

// 设置能够攻击
func (ca *Card) SetCanAttack() {
	ca.lastAttackRound = 1
}

// 设置不能够攻击
func (ca *Card) SetNotCanAttack() {
	ca.lastAttackRound = 0
}

// 设置表示形式
func (ca *Card) setLE(l le_type) {
	if ca.le == l {
		return
	}
	l &= ^LE_peek
	oo := ca.le
	b := ca.IsFaceDown() && ca.IsInMzone()
	ca.le = l
	if b && ca.IsFaceUp() {
		ca.Dispatch(FaceUp)
	}
	pl := ca.GetSummoner()

	ca.Dispatch(Expres, oo)
	pl.callAll(exprCard(ca, l))
	if ca.IsFaceUp() && ca.GetId() != 0 {
		pl.callAll(setFront(ca))
	}
}

// 判断是攻击表示
func (ca *Card) IsAttack() bool {
	return ca.le.IsAttack()
}

// 设置攻击表示
func (ca *Card) SetFaceAttack() {
	ca.setLE(LE_Attack | (ca.le & LE_fd))
}

// 判断是防御表示
func (ca *Card) IsDefense() bool {
	return ca.le.IsDefense()
}

// 设置防御表示
func (ca *Card) SetFaceDefense() {
	ca.setLE(LE_Defense | (ca.le & LE_fd))
}

// 判断是面朝
func (ca *Card) IsFaceUp() bool {
	return ca.le.IsFaceUp()
}

// 设置是面朝上
func (ca *Card) setFaceUp() {
	ca.setLE(LE_FaceUp | (ca.le & LE_ad))
}

func (ca *Card) SetFaceUp() {
	b := ca.IsFaceDown()
	ca.setFaceUp()
	if b {
		ca.Dispatch(Flip)
	}
}

// 判断是面朝下
func (ca *Card) IsFaceDown() bool {
	return ca.le.IsFaceDown()
}

// 设置是面朝下
func (ca *Card) SetFaceDown() {
	ca.setLE(LE_FaceDown | (ca.le & LE_ad))
}

// 判断是面朝上攻击表示
func (ca *Card) IsFaceUpAttack() bool {
	return ca.le.IsFaceUpAttack()
}

// 设置是面朝上攻击表示
func (ca *Card) setFaceUpAttack() {
	ca.setLE(LE_FaceUpAttack)
}

func (ca *Card) SetFaceUpAttack() {
	b := ca.IsFaceDown()
	ca.setFaceUpAttack()
	if b {
		ca.Dispatch(Flip)
	}
}

// 判断是面朝下攻击表示
func (ca *Card) IsFaceDownAttack() bool {
	return ca.le.IsFaceDownAttack()
}

// 设置是面朝下攻击表示
func (ca *Card) SetFaceDownAttack() {
	ca.setLE(LE_FaceDownAttack)
}

// 判断是面朝上防御表示
func (ca *Card) IsFaceUpDefense() bool {
	return ca.le.IsFaceUpDefense()
}

// 设置是面朝上防御表示
func (ca *Card) setFaceUpDefense() {
	ca.setLE(LE_FaceUpDefense)
}

func (ca *Card) SetFaceUpDefense() {
	b := ca.IsFaceDown()
	ca.setFaceUpDefense()
	if b {
		ca.Dispatch(Flip)
	}
}

// 判断是面朝下防御表示
func (ca *Card) IsFaceDownDefense() bool {
	return ca.le.IsFaceDownDefense()
}

// 设置是面朝下防御表示
func (ca *Card) SetFaceDownDefense() {
	ca.setLE(LE_FaceDownDefense)
}

// 拿起  不属于任何牌堆里
func (ca *Card) Placed() {
	if ca.place != nil {
		ca.place.PickedFor(ca)
	}
}

// 移动到墓地
func (ca *Card) ToGrave() {
	ca.GetOwner().Grave().EndPush(ca)
}

// 移动到除外
func (ca *Card) ToRemoved() {
	ca.GetOwner().Removed().EndPush(ca)
}

// 移动到手牌
func (ca *Card) ToHand() {
	ca.GetOwner().Hand().EndPush(ca)
}

// 移动到额外
func (ca *Card) ToExtra() {
	ca.GetOwner().Extra().EndPush(ca)
}

// 移动到怪兽
func (ca *Card) ToMzone() {
	pl := ca.GetSummoner()
	if pl.Mzone().Len() >= 5 {
		ca.Dispatch(DestroyBeRule)
	} else {
		pl.Mzone().EndPush(ca)
	}
}

// 移动到魔法
func (ca *Card) ToSzone() {
	pl := ca.GetOwner()
	if pl.Szone().Len() >= 5 {
		ca.Dispatch(DestroyBeRule)
	} else {
		pl.Szone().EndPush(ca)
	}
}

// 移动到卡组
func (ca *Card) ToDeckTop() {
	ca.GetOwner().Deck().EndPush(ca)
}

func (ca *Card) ToDeckBot() {
	ca.GetOwner().Deck().BeginPush(ca)
}

// 移动到场地
func (ca *Card) ToField() {
	f := ca.GetOwner().Field()
	f.ForEach(func(c *Card) bool {
		c.ToGrave()
		return true
	})
	f.EndPush(ca)
}

// 是在场地
func (ca *Card) IsInField() bool {
	return ca.GetPlace().IsField()
}

// 是在卡组
func (ca *Card) IsInDeck() bool {
	return ca.GetPlace().IsDeck()
}

// 是在额外
func (ca *Card) IsInExtra() bool {
	return ca.GetPlace().IsExtra()
}

// 是在墓地
func (ca *Card) IsInGrave() bool {
	return ca.GetPlace().IsGrave()
}

// 是在手牌
func (ca *Card) IsInHand() bool {
	return ca.GetPlace().IsHand()
}

// 是在怪兽区
func (ca *Card) IsInMzone() bool {
	return ca.GetPlace().IsMzone()
}

// 是在魔陷区
func (ca *Card) IsInSzone() bool {
	return ca.GetPlace().IsSzone()
}

// 是在除外
func (ca *Card) IsInRemoved() bool {
	return ca.GetPlace().IsRemoved()
}

// 是头像
func (ca *Card) IsPortrait() bool {
	return ca.GetPlace().IsPortrait()
}

// 被破坏
func (ca *Card) Destroy(c *Card) {
	ca.Dispatch(Destroy, c)
}

// 被支付
func (ca *Card) Cost(c *Card) {
	ca.Dispatch(Cost, c)
}

// 被丢弃
func (ca *Card) Discard(c *Card) {
	ca.Dispatch(Discard, c)
}

// 被移除
func (ca *Card) Removed(c *Card) {
	ca.Dispatch(Removed, c)
}

// 被使用
func (ca *Card) Depleted(c *Card) {
	ca.Dispatch(Depleted, c)
}

// 特殊召唤
func (ca *Card) SummonSpecial(c *Card) {
	ca.Init()
	ca.Dispatch(SummonSpecial, c)
}

// 通常召唤
func (ca *Card) SummonNormal(c *Card) {
	ca.Dispatch(Summon, c)
}
