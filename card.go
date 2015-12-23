package ygo_core

import (
	"github.com/wzshiming/base"
	"github.com/wzshiming/dispatcher"
	"github.com/wzshiming/ffmt"
)

type CardOriginal struct {
	IsValid  bool    // 是否有效
	Id       uint    // 卡牌id
	Name     string  // 名字
	Password string  // 卡牌密码
	Lc       lc_type // 卡牌类型

	// 怪兽卡 属性
	La    la_type // 怪兽属性
	Lr    lr_type // 怪兽种族
	Level int     // 星级
	Atk   int     // 攻击力
	Def   int     // 防御力

	Initialize Action // 初始化

}

func (co CardOriginal) String() string {
	if co.Level != 0 {
		return ffmt.Sputs(map[string]interface{}{
			"Name": co.Name,
			"Id":   co.Id,
			"Pwd":  co.Password,
			"Type": co.Lc,
			"Arrt": co.La,
			"Race": co.Lr,
			"Lv":   co.Level,
			"Atk":  co.Atk,
			"Def":  co.Def,
		})
	} else {
		return ffmt.Sputs(map[string]interface{}{
			"Name": co.Name,
			"Id":   co.Id,
			"Pwd":  co.Password,
			"Type": co.Lc,
		})
	}
}

func (co *CardOriginal) Make(pl *Player) *Card {
	c := &Card{
		Events:       dispatcher.NewForkEvent(pl.GetFork()),
		baseOriginal: co,
		owner:        pl,
		summoner:     pl,
		le:           LE_FaceDownAttack,
	}
	c.InitUint()
	c.Init()
	pl.Game().registerCards(c)
	return c
}

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
	switch ca.baseOriginal.Lc {

	case LC_SpellQuickPlay: //速攻魔法 速度2
		return 2
	case LC_TrapNormal: //普通陷阱 速度2
		return 2
	case LC_TrapContinuous: //永续陷阱 速度2
		return 2
	case LC_TrapCounter: //反击陷阱 速度3
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
func (ca *Card) GetBaseType() lc_type {
	return ca.baseOriginal.Lc
}

func (ca *Card) Is(a ...interface{}) bool {
	for _, v := range a {
		switch s := v.(type) {
		case lc_type:
			if ca.appendOriginal.Lc&s == 0 {
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

func (ca *Card) IsExtra() bool {
	return !(ca.IsSpellAndTrap() || ca.IsMonsterNormal() || ca.IsMonsterEffect())
}

// 获得类型
func (ca *Card) GetType() lc_type {
	return ca.appendOriginal.Lc
}

// 是魔法卡
func (ca *Card) IsSpellAndTrap() bool {
	return (ca.GetType() & LC_SpellAndTrap) != 0
}

// 是魔法卡
func (ca *Card) IsSpell() bool {
	return (ca.GetType() & LC_Spell) != 0
}

// 是陷阱卡
func (ca *Card) IsTrap() bool {
	return (ca.GetType() & LC_Trap) != 0
}

// 是怪兽卡
func (ca *Card) IsMonster() bool {
	return (ca.GetType() & LC_Monster) != 0
}

// 是普通怪兽
func (ca *Card) IsMonsterNormal() bool {
	return (ca.GetType() & LC_MonsterNormal) != 0
}

// 是效果怪兽
func (ca *Card) IsMonsterEffect() bool {
	return (ca.GetType() & LC_MonsterEffect) != 0
}

// 是融合怪兽
func (ca *Card) IsMonsterFusion() bool {
	return (ca.GetType() & LC_MonsterFusion) != 0
}

// 是超量怪兽
func (ca *Card) IsMonsterXyz() bool {
	return (ca.GetType() & LC_MonsterXyz) != 0
}

// 是同调怪兽
func (ca *Card) IsMonsterSynchro() bool {
	return (ca.GetType() & LC_MonsterSynchro) != 0
}

// 是仪式怪兽
func (ca *Card) IsMonsterRitual() bool {
	return (ca.GetType() & LC_MonsterRitual) != 0
}

// 是普通魔法
func (ca *Card) IsSpellNormal() bool {
	return (ca.GetType() & LC_SpellNormal) != 0
}

// 是仪式魔法
func (ca *Card) IsSpellRitual() bool {
	return (ca.GetType() & LC_SpellRitual) != 0
}

// 是永续魔法
func (ca *Card) IsSpellContinuous() bool {
	return (ca.GetType() & LC_SpellContinuous) != 0
}

// 是装备魔法
func (ca *Card) IsSpellEquip() bool {
	return (ca.GetType() & LC_SpellEquip) != 0
}

// 是场地魔法
func (ca *Card) IsSpellField() bool {
	return (ca.GetType() & LC_SpellField) != 0
}

// 是速攻魔法
func (ca *Card) IsSpellQuickPlay() bool {
	return (ca.GetType() & LC_SpellQuickPlay) != 0
}

// 是普通陷阱
func (ca *Card) IsTrapNormal() bool {
	return (ca.GetType() & LC_TrapNormal) != 0
}

// 是永续陷阱
func (ca *Card) IsSustainsTrap() bool {
	return (ca.GetType() & LC_TrapContinuous) != 0
}

// 是反击陷阱
func (ca *Card) IsTrapCounter() bool {
	return (ca.GetType() & LC_TrapCounter) != 0
}

// 是特殊作用卡牌
func (ca *Card) IsNone() bool {
	return ca.GetType() == LC_None
}

// 设置类型
func (ca *Card) SetType(l lc_type) {
	ca.appendOriginal.Lc = l
}

// 获得基础属性
func (ca *Card) GetBaseAttribute() la_type {
	return ca.baseOriginal.La
}

// 获得属性
func (ca *Card) GetAttribute() la_type {
	return ca.appendOriginal.La
}

//  设置属性
func (ca *Card) SetAttribute(l la_type) {
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

//// 设置不能够数攻击
//func (ca *Card) SetSizeRoundNotCanAttack(i int) {
//	ca.lastAttackRound = ca.GetSummoner().GetRound() + i
//}

// 设置表示形式
func (ca *Card) setLE(l le_type) {
	b := ca.IsFaceDown() && ca.IsInMzone()
	ca.le = l
	if b && ca.IsFaceUp() {
		ca.Dispatch(FaceUp)
	}
	pl := ca.GetSummoner()
	pl.Dispatch(Expres, ca)
	pl.callAll(exprCard(ca, l))
	if ca.IsFaceUp() && ca.GetId() != 0 {
		pl.callAll(setFront(ca))
	}
}

// 判断是攻击表示
func (ca *Card) IsAttack() bool {
	return (ca.le & LE_Attack) == LE_Attack
}

// 设置攻击表示
func (ca *Card) SetFaceAttack() {
	ca.setLE(LE_Attack | (ca.le & LE_fd))
}

// 判断是防御表示
func (ca *Card) IsDefense() bool {
	return (ca.le & LE_Defense) == LE_Defense
}

// 设置防御表示
func (ca *Card) SetFaceDefense() {
	ca.setLE(LE_Defense | (ca.le & LE_fd))
}

// 判断是面朝
func (ca *Card) IsFaceUp() bool {
	return (ca.le & LE_FaceUp) == LE_FaceUp
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
	return (ca.le & LE_FaceDown) == LE_FaceDown
}

// 设置是面朝下
func (ca *Card) SetFaceDown() {
	ca.setLE(LE_FaceDown | (ca.le & LE_ad))
}

// 判断是面朝上攻击表示
func (ca *Card) IsFaceUpAttack() bool {
	return (ca.le & LE_FaceUpAttack) == LE_FaceUpAttack
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
	return (ca.le & LE_FaceDownAttack) == LE_FaceDownAttack
}

// 设置是面朝下攻击表示
func (ca *Card) SetFaceDownAttack() {
	ca.setLE(LE_FaceDownAttack)
}

// 判断是面朝上防御表示
func (ca *Card) IsFaceUpDefense() bool {
	return (ca.le & LE_FaceUpDefense) == LE_FaceUpDefense
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
	return (ca.le & LE_FaceDownDefense) == LE_FaceDownDefense
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
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Field
}

// 是在卡组
func (ca *Card) IsInDeck() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Deck
}

// 是在额外
func (ca *Card) IsInExtra() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Extra
}

// 是在墓地
func (ca *Card) IsInGrave() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Grave
}

// 是在手牌
func (ca *Card) IsInHand() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Hand
}

// 是在怪兽区
func (ca *Card) IsInMzone() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Mzone
}

// 是在魔陷区
func (ca *Card) IsInSzone() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Szone
}

// 是在手牌
func (ca *Card) IsInRemoved() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Removed
}

// 是头像
func (ca *Card) IsPortrait() bool {
	p := ca.GetPlace()
	return p != nil && p.GetName() == LL_Portrait
}

//战士族
func (ca *Card) RaceIsWarrior() bool {
	return (ca.appendOriginal.Lr & LR_Warrior) != LR_None
}

//魔法使用族
func (ca *Card) RaceIsSpellcaster() bool {
	return (ca.appendOriginal.Lr & LR_Spellcaster) != LR_None
}

//精灵族 天使族
func (ca *Card) RaceIsFairy() bool {
	return (ca.appendOriginal.Lr & LR_Fairy) != LR_None
}

//恶魔族
func (ca *Card) RaceIsFiend() bool {
	return (ca.appendOriginal.Lr & LR_Fiend) != LR_None
}

//不死族
func (ca *Card) RaceIsZombie() bool {
	return (ca.appendOriginal.Lr & LR_Zombie) != LR_None
}

//机械族
func (ca *Card) RaceIsMachine() bool {
	return (ca.appendOriginal.Lr & LR_Machine) != LR_None
}

//水族
func (ca *Card) RaceIsWater() bool {
	return (ca.appendOriginal.Lr & LR_Water) != LR_None
}

//炎族
func (ca *Card) RaceIsFire() bool {
	return (ca.appendOriginal.Lr & LR_Fire) != LR_None
}

//岩石族
func (ca *Card) RaceIsRock() bool {
	return (ca.appendOriginal.Lr & LR_Rock) != LR_None
}

//鸟兽族
func (ca *Card) RaceIsWingedBeast() bool {
	return (ca.appendOriginal.Lr & LR_WingedBeast) != LR_None
}

//植物族
func (ca *Card) RaceIsPlant() bool {
	return (ca.appendOriginal.Lr & LR_Plant) != LR_None
}

//昆虫族
func (ca *Card) RaceIsInsect() bool {
	return (ca.appendOriginal.Lr & LR_Insect) != LR_None
}

//雷族
func (ca *Card) RaceIsThunder() bool {
	return (ca.appendOriginal.Lr & LR_Thunder) != LR_None
}

//龙族
func (ca *Card) RaceIsDragon() bool {
	return (ca.appendOriginal.Lr & LR_Dragon) != LR_None
}

//兽族
func (ca *Card) RaceIsBeast() bool {
	return (ca.appendOriginal.Lr & LR_Beast) != LR_None
}

//兽战士族
func (ca *Card) RaceIsBeastWarrior() bool {
	return (ca.appendOriginal.Lr & LR_BeastWarrior) != LR_None
}

//恐龙族
func (ca *Card) RaceIsDinosaur() bool {
	return (ca.appendOriginal.Lr & LR_Dinosaur) != LR_None
}

//鱼族
func (ca *Card) RaceIsFish() bool {
	return (ca.appendOriginal.Lr & LR_Fish) != LR_None
}

//海龙族
func (ca *Card) RaceIsSeaSerpent() bool {
	return (ca.appendOriginal.Lr & LR_SeaSerpent) != LR_None
}

//爬虫族
func (ca *Card) RaceIsReptile() bool {
	return (ca.appendOriginal.Lr & LR_Reptile) != LR_None
}

//念动力族
func (ca *Card) RaceIsPsychic() bool {
	return (ca.appendOriginal.Lr & LR_Psychic) != LR_None
}

//幻神兽族
func (ca *Card) RaceIsDivineBeast() bool {
	return (ca.appendOriginal.Lr & LR_DivineBeast) != LR_None
}

//地
func (ca *Card) AttrIsEarth() bool {
	return ca.appendOriginal.La == LA_Earth
}

//水
func (ca *Card) AttrIsWater() bool {
	return ca.appendOriginal.La == LA_Water
}

//火
func (ca *Card) AttrIsFire() bool {
	return ca.appendOriginal.La == LA_Fire
}

//风
func (ca *Card) AttrIsWind() bool {
	return ca.appendOriginal.La == LA_Wind
}

//光
func (ca *Card) AttrIsLight() bool {
	return ca.appendOriginal.La == LA_Light
}

//暗
func (ca *Card) AttrIsDark() bool {
	return ca.appendOriginal.La == LA_Dark
}

//神
func (ca *Card) AttrIsDevine() bool {
	return ca.appendOriginal.La == LA_Devine
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
