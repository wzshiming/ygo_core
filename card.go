package ygo_cord

import (
	"github.com/wzshiming/base"
	"github.com/wzshiming/dispatcher"
)

type CardOriginal struct {
	IsValid  bool    // 是否有效
	Id       uint    // 卡牌id
	Name     string  // 名字
	Password string  // 卡牌密码
	Lc       lc_type // 卡牌类型

	// 怪兽卡 属性
	La      la_type // 怪兽属性
	Lr      lr_type // 怪兽种族
	Level   int     // 星级
	Attack  int     // 攻击力
	Defense int     // 防御力

	Initialize Action // 初始化

}

func NewNoneCardOriginal() *CardOriginal {
	return &CardOriginal{
		IsValid: true,
		Lc:      LC_None,
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
	pl.Game().RegisterCards(c)
	return c
}

type Card struct {
	dispatcher.Events
	base.Unique
	isValid      bool
	baseOriginal *CardOriginal
	original     CardOriginal
	place        *Group  // 所在位置
	summoner     *Player // 召唤者
	owner        *Player // 所有者
	le           le_type // 表示形式
	//怪兽卡 属性
	counter         int  // 计数器
	lastAttackRound int  // 最后攻击回合
	lastChangeRound int  // 最后改变表示形式回合
	direct          bool // 直接攻击玩家
}

func (ca *Card) Peek() {
	pl := ca.GetSummoner()
	pl.Call(setFront(ca))
	pl.Call(exprCard(ca, LE_FaceUpAttack))
	pl.GetTarget().Call(exprCard(ca, LE_FaceDownAttack))
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
	pl.CallAll(setCardFace(ca, Arg{"攻击力": ca.GetAttack(), "防御力": ca.GetDefense()}))
}

func (ca *Card) HideInfo() {
	pl := ca.GetSummoner()
	pl.CallAll(setCardFace(ca, Arg{}))
}

func (ca *Card) IsValid() bool {
	return ca.isValid
}

func (ca *Card) Priority() int {
	switch ca.baseOriginal.Lc {

	case LC_RushMagic: //速攻魔法 速度2
		return 2
	case LC_OrdinaryTrap: //普通陷阱 速度2
		return 2
	case LC_SustainsTrap: //永续陷阱 速度2
		return 2
	case LC_ReactionTrap: //反击陷阱 速度3
		return 3
	default:
		return 1
	}
	return 0
}

func (ca *Card) Dispatch(eventName string, args ...interface{}) {
	yg := ca.GetSummoner().Game()
	if Pay != eventName && Chain != eventName {
		ca.Events.Dispatch(Pay, eventName)
		if ca.IsOpen(eventName) {
			yg.Chain(eventName, ca, ca.GetSummoner(), args)
		}
	}
	ca.Events.Dispatch(eventName, args...)
}

func (ca *Card) GetPlace() *Group {
	return ca.place
}

// 设置名字
func (ca *Card) SetName(name string) {
	ca.original.Name = name
}

// 获得名字
func (ca *Card) GetName() string {
	return ca.original.Name
}

// 获得召唤者
func (ca *Card) GetSummoner() *Player {
	return ca.summoner
}

// 设置召唤者
func (ca *Card) SetSummoner(c *Player) {
	ca.IsCanAttack()
	ca.summoner = c
	ca.Dispatch(Change)
}

// 获得所有者
func (ca *Card) GetOwner() *Player {
	return ca.owner
}

// 获得 id
func (ca *Card) GetId() uint {
	return ca.original.Id
}

// 获得基础类型
func (ca *Card) GetBaseType() lc_type {
	return ca.baseOriginal.Lc
}

func (ca *Card) Is(a ...interface{}) bool {
	for _, v := range a {
		switch s := v.(type) {
		case lc_type:
			if ca.original.Lc&s == 0 {
				return false
			}
		case la_type:
			if ca.original.La&s == 0 {
				return false
			}
		case le_type:
			if ca.le&s == 0 {
				return false
			}
		case lr_type:
			if ca.original.Lr&s == 0 {
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
func (ca *Card) GetType() lc_type {
	return ca.original.Lc
}

// 是魔法卡
func (ca *Card) IsMagicAndTrap() bool {
	return (ca.GetType() & LC_MagicAndTrap) != 0
}

// 是魔法卡
func (ca *Card) IsMagic() bool {
	return (ca.GetType() & LC_Magic) != 0
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
func (ca *Card) IsOrdinaryMonster() bool {
	return (ca.GetType() & LC_OrdinaryMonster) != 0
}

// 是效果怪兽
func (ca *Card) IsEffectMonster() bool {
	return (ca.GetType() & LC_EffectMonster) != 0
}

// 是融合怪兽
func (ca *Card) IsFusionMonster() bool {
	return (ca.GetType() & LC_FusionMonster) != 0
}

// 是超量怪兽
func (ca *Card) IsExcessMonster() bool {
	return (ca.GetType() & LC_ExcessMonster) != 0
}

// 是同调怪兽
func (ca *Card) IsHomologyMonster() bool {
	return (ca.GetType() & LC_HomologyMonster) != 0
}

// 是仪式怪兽
func (ca *Card) IsRiteMonster() bool {
	return (ca.GetType() & LC_RiteMonster) != 0
}

// 是普通魔法
func (ca *Card) IsOrdinaryMagic() bool {
	return (ca.GetType() & LC_OrdinaryMagic) != 0
}

// 是仪式魔法
func (ca *Card) IsRiteMagic() bool {
	return (ca.GetType() & LC_RiteMagic) != 0
}

// 是永续魔法
func (ca *Card) IsSustainsMagic() bool {
	return (ca.GetType() & LC_SustainsMagic) != 0
}

// 是装备魔法
func (ca *Card) IsEquipMagic() bool {
	return (ca.GetType() & LC_EquipMagic) != 0
}

// 是场地魔法
func (ca *Card) IsPlaceMagic() bool {
	return (ca.GetType() & LC_PlaceMagic) != 0
}

// 是速攻魔法
func (ca *Card) IsRushMagic() bool {
	return (ca.GetType() & LC_RushMagic) != 0
}

// 是普通陷阱
func (ca *Card) IsOrdinaryTrap() bool {
	return (ca.GetType() & LC_OrdinaryTrap) != 0
}

// 是永续陷阱
func (ca *Card) IsSustainsTrap() bool {
	return (ca.GetType() & LC_SustainsTrap) != 0
}

// 是反击陷阱
func (ca *Card) IsReactionTrap() bool {
	return (ca.GetType() & LC_ReactionTrap) != 0
}

// 是特殊作用卡牌
func (ca *Card) IsNone() bool {
	return ca.GetType() == LC_None
}

// 设置类型
func (ca *Card) SetType(l lc_type) {
	ca.original.Lc = l
	ca.Dispatch(Change)
}

// 获得基础属性
func (ca *Card) GetBaseAttribute() la_type {
	return ca.baseOriginal.La
}

// 获得属性
func (ca *Card) GetAttribute() la_type {
	return ca.original.La
}

//  设置属性
func (ca *Card) SetAttribute(l la_type) {
	ca.original.La = l
	ca.Dispatch(Change)
}

// 获得基础种族
func (ca *Card) GetBaseRace() lr_type {
	return ca.baseOriginal.Lr
}

// 获得种族
func (ca *Card) GetRace() lr_type {
	return ca.original.Lr
}

// 设置种族
func (ca *Card) SetRace(l lr_type) {
	ca.original.Lr = l
	ca.Dispatch(Change)
}

// 获得基础攻击
func (ca *Card) GetBaseAttack() int {
	return ca.baseOriginal.Attack
}

// 获得攻击
func (ca *Card) GetAttack() int {
	return ca.original.Attack
}

// 设置攻击
func (ca *Card) SetAttack(i int) {
	if i < 0 {
		i = 0
	}
	ca.original.Attack = i
	ca.Dispatch(Change)
	pl := ca.GetSummoner()
	pl.CallAll(setCardFace(ca, Arg{"攻击力": i}))
}

// 获得基础防御
func (ca *Card) GetBaseDefense() int {
	return ca.baseOriginal.Defense
}

// 获得防御
func (ca *Card) GetDefense() int {
	return ca.original.Defense
}

// 设置防御
func (ca *Card) SetDefense(i int) {
	if i < 0 {
		i = 0
	}
	ca.original.Defense = i
	ca.Dispatch(Change)
	pl := ca.GetSummoner()
	pl.CallAll(setCardFace(ca, Arg{"防御力": i}))
}

// 获得基础等级
func (ca *Card) GetBaseLevel() int {
	return ca.baseOriginal.Level
}

// 获得等级
func (ca *Card) GetLevel() int {
	return ca.original.Level
}

// 设置等级
func (ca *Card) SetLevel(i int) {
	ca.original.Level = i
	ca.Dispatch(Change)
}

// 判断能够改变表示形式
func (ca *Card) IsCanChange() bool {
	return ca.lastChangeRound < ca.GetSummoner().GetRound()
}

// 设置能够改变表示形式
func (ca *Card) SetCanChange() {
	ca.lastChangeRound = 0
	ca.Dispatch(Change)
}

// 设置不能够改变表示形式
func (ca *Card) SetNotCanChange() {
	ca.lastChangeRound = ca.GetSummoner().GetRound()
}

// 判断能够攻击
func (ca *Card) IsCanAttack() bool {
	return ca.lastAttackRound < ca.GetSummoner().GetRound()
}

// 设置能够攻击
func (ca *Card) SetCanAttack() {
	ca.lastAttackRound = 0
}

// 设置不能够攻击
func (ca *Card) SetNotCanAttack() {
	ca.lastAttackRound = ca.GetSummoner().GetRound()
}

// 设置不能够数攻击
func (ca *Card) SetSizeRoundNotCanAttack(i int) {
	ca.lastAttackRound = ca.GetSummoner().GetRound() + i
}

// 设置表示形式
func (ca *Card) setLE(l le_type) {
	ca.le = l
	pl := ca.GetSummoner()
	pl.Dispatch(Expres, ca)
	pl.CallAll(exprCard(ca, l))
	if ca.IsFaceUp() && ca.GetId() != 0 {
		pl.CallAll(setFront(ca))
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
func (ca *Card) SetFaceUp() {
	ca.setLE(LE_FaceUp | (ca.le & LE_ad))
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
func (ca *Card) SetFaceUpAttack() {
	ca.setLE(LE_FaceUpAttack)
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
func (ca *Card) SetFaceUpDefense() {
	ca.setLE(LE_FaceUpDefense)
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
	ca.GetOwner().Grave.EndPush(ca)
}

// 移动到除外
func (ca *Card) ToRemoved() {
	ca.GetOwner().Removed.EndPush(ca)
}

// 移动到手牌
func (ca *Card) ToHand() {
	ca.GetOwner().Hand.EndPush(ca)
}

// 移动到额外
func (ca *Card) ToExtra() {
	ca.GetOwner().Extra.EndPush(ca)
}

// 移动到怪兽
func (ca *Card) ToMzone() {
	pl := ca.GetSummoner()
	if pl.Mzone.Len() >= 5 {
		ca.Dispatch(DestroyBeRule)
	} else {
		pl.Mzone.EndPush(ca)
	}
}

// 移动到魔法
func (ca *Card) ToSzone() {
	pl := ca.GetOwner()
	if pl.Szone.Len() >= 5 {
		ca.Dispatch(DestroyBeRule)
	} else {
		pl.Szone.EndPush(ca)
	}
}

// 移动到卡组
func (ca *Card) ToDeck() {
	ca.GetOwner().Deck.EndPush(ca)
}

// 移动到场地
func (ca *Card) ToField() {
	f := ca.GetOwner().Field
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
	return ca.original.Lr == LR_Warrior
}

//魔法使用族
func (ca *Card) RaceIsSpellCaster() bool {
	return ca.original.Lr == LR_SpellCaster
}

//精灵族
func (ca *Card) RaceIsFairy() bool {
	return ca.original.Lr == LR_Fairy
}

//恶魔族
func (ca *Card) RaceIsFiend() bool {
	return ca.original.Lr == LR_Fiend
}

//不死族
func (ca *Card) RaceIsZombie() bool {
	return ca.original.Lr == LR_Zombie
}

//机械族
func (ca *Card) RaceIsMachine() bool {
	return ca.original.Lr == LR_Machine
}

//水族
func (ca *Card) RaceIsAqua() bool {
	return ca.original.Lr == LR_Aqua
}

//炎族
func (ca *Card) RaceIsPyro() bool {
	return ca.original.Lr == LR_Pyro
}

//岩石族
func (ca *Card) RaceIsRock() bool {
	return ca.original.Lr == LR_Rock
}

//鸟兽族
func (ca *Card) RaceIsWindBeast() bool {
	return ca.original.Lr == LR_WindBeast
}

//植物族
func (ca *Card) RaceIsPlant() bool {
	return ca.original.Lr == LR_Plant
}

//昆虫族
func (ca *Card) RaceIsInsect() bool {
	return ca.original.Lr == LR_Insect
}

//雷族
func (ca *Card) RaceIsThunder() bool {
	return ca.original.Lr == LR_Thunder
}

//龙族
func (ca *Card) RaceIsDragon() bool {
	return ca.original.Lr == LR_Dragon
}

//兽族
func (ca *Card) RaceIsBeast() bool {
	return ca.original.Lr == LR_Beast
}

//兽战士族
func (ca *Card) RaceIsBeastWarror() bool {
	return ca.original.Lr == LR_BeastWarror
}

//恐龙族
func (ca *Card) RaceIsDinosaur() bool {
	return ca.original.Lr == LR_Dinosaur
}

//鱼族
func (ca *Card) RaceIsFRaceIsh() bool {
	return ca.original.Lr == LR_Fish
}

//海龙族
func (ca *Card) RaceIsSeaserpent() bool {
	return ca.original.Lr == LR_Seaserpent
}

//爬虫族
func (ca *Card) RaceIsReptile() bool {
	return ca.original.Lr == LR_Reptile
}

//念动力族
func (ca *Card) RaceIsPsycho() bool {
	return ca.original.Lr == LR_Psycho
}

//幻神兽族
func (ca *Card) RaceIsDivineBeast() bool {
	return ca.original.Lr == LR_DivineBeast
}

//天使族
func (ca *Card) RaceIsAngel() bool {
	return ca.original.Lr == LR_Angel
}

//创造神族
func (ca *Card) RaceIsCreatorGod() bool {
	return ca.original.Lr == LR_CreatorGod
}

//幻龙族
func (ca *Card) RaceIsPhantomDragon() bool {
	return ca.original.Lr == LR_PhantomDragon
}

//地
func (ca *Card) AttrIsEarth() bool {
	return ca.original.La == LA_Earth
}

//水
func (ca *Card) AttrIsWater() bool {
	return ca.original.La == LA_Water
}

//火
func (ca *Card) AttrIsFire() bool {
	return ca.original.La == LA_Fire
}

//风
func (ca *Card) AttrIsWind() bool {
	return ca.original.La == LA_Wind
}

//光
func (ca *Card) AttrIsLight() bool {
	return ca.original.La == LA_Light
}

//暗
func (ca *Card) AttrIsDark() bool {
	return ca.original.La == LA_Dark
}

//神
func (ca *Card) AttrIsDevine() bool {
	return ca.original.La == LA_Devine
}
