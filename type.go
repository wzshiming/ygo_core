package ygo_core

// 卡牌类型 Types
type lt_type uint32

const (
	LT_None lt_type = 0

	LT_MonsterNormal   lt_type = 1 << (32 - 1 - iota) //普通怪兽 黄色
	LT_MonsterEffect                                  //效果怪兽 橙色
	LT_MonsterFusion                                  //融合怪兽 紫色
	LT_MonsterXyz                                     //超量怪兽 黑色
	LT_MonsterSynchro                                 //同调怪兽 白色
	LT_MonsterRitual                                  //仪式怪兽 蓝色
	LT_SpellNormal                                    //普通魔法 通常
	LT_SpellRitual                                    //仪式魔法
	LT_SpellContinuous                                //永续魔法
	LT_SpellEquip                                     //装备魔法
	LT_SpellField                                     //场地魔法
	LT_SpellQuickPlay                                 //速攻魔法 速度2
	LT_TrapNormal                                     //普通陷阱 速度2
	LT_TrapContinuous                                 //永续陷阱 速度2
	LT_TrapCounter                                    //反击陷阱 速度3

	// 怪物卡
	LT_Monster = LT_MonsterNormal | LT_MonsterEffect | LT_MonsterXyz | LT_MonsterSynchro | LT_MonsterFusion | LT_MonsterRitual
	// 魔法卡
	LT_Spell = LT_SpellNormal | LT_SpellRitual | LT_SpellContinuous | LT_SpellEquip | LT_SpellField | LT_SpellQuickPlay
	// 陷阱卡
	LT_Trap = LT_TrapNormal | LT_TrapContinuous | LT_TrapCounter
	// 魔法卡与陷阱卡
	LT_SpellAndTrap = LT_Spell | LT_Trap
)

var lc = map[lt_type]string{
	LT_None:            "NoneType",
	LT_MonsterNormal:   "MonsterNormal",   //普通怪兽 黄色
	LT_MonsterEffect:   "MonsterEffect",   //效果怪兽 橙色
	LT_MonsterFusion:   "MonsterFusion",   //融合怪兽 紫色
	LT_MonsterXyz:      "MonsterXyz",      //超量怪兽 黑色
	LT_MonsterSynchro:  "MonsterSynchro",  //同调怪兽 白色
	LT_MonsterRitual:   "MonsterRitual",   //仪式怪兽 蓝色
	LT_SpellNormal:     "SpellNormal",     //普通魔法 通常
	LT_SpellRitual:     "SpellRitual",     //仪式魔法
	LT_SpellContinuous: "SpellContinuous", //永续魔法
	LT_SpellEquip:      "SpellEquip",      //装备魔法
	LT_SpellField:      "SpellField",      //场地魔法
	LT_SpellQuickPlay:  "SpellQuickPlay",  //速攻魔法 速度2
	LT_TrapNormal:      "TrapNormal",      //普通陷阱 速度2
	LT_TrapContinuous:  "TrapContinuous",  //永续陷阱 速度2
	LT_TrapCounter:     "TrapCounter",     //反击陷阱 速度3
}

func (c lt_type) String() (s string) {
	if c != LT_None {
		s = lc[c]
		if s != "" {
			return
		}
		for k, v := range lc {
			if (k & c) != 0 {
				s += v
				s += ""
			}
		}
		return
	}
	return lc[c]
}

// 是特殊作用卡牌
func (c lt_type) IsNone() bool {
	return c == LT_None
}

// 是额外的
func (c lt_type) IsExtra() bool {
	return !(c.IsSpellAndTrap() || c.IsMonsterNormal() || c.IsMonsterEffect())
}

// 是魔法卡
func (c lt_type) IsSpellAndTrap() bool {
	return (c & LT_SpellAndTrap) != 0
}

// 是魔法卡
func (c lt_type) IsSpell() bool {
	return (c & LT_Spell) != 0
}

// 是陷阱卡
func (c lt_type) IsTrap() bool {
	return (c & LT_Trap) != 0
}

// 是怪兽卡
func (c lt_type) IsMonster() bool {
	return (c & LT_Monster) != 0
}

// 是普通怪兽
func (c lt_type) IsMonsterNormal() bool {
	return (c & LT_MonsterNormal) != 0
}

// 是效果怪兽
func (c lt_type) IsMonsterEffect() bool {
	return (c & LT_MonsterEffect) != 0
}

// 是融合怪兽
func (c lt_type) IsMonsterFusion() bool {
	return (c & LT_MonsterFusion) != 0
}

// 是超量怪兽
func (c lt_type) IsMonsterXyz() bool {
	return (c & LT_MonsterXyz) != 0
}

// 是同调怪兽
func (c lt_type) IsMonsterSynchro() bool {
	return (c & LT_MonsterSynchro) != 0
}

// 是仪式怪兽
func (c lt_type) IsMonsterRitual() bool {
	return (c & LT_MonsterRitual) != 0
}

// 是普通魔法
func (c lt_type) IsSpellNormal() bool {
	return (c & LT_SpellNormal) != 0
}

// 是仪式魔法
func (c lt_type) IsSpellRitual() bool {
	return (c & LT_SpellRitual) != 0
}

// 是永续魔法
func (c lt_type) IsSpellContinuous() bool {
	return (c & LT_SpellContinuous) != 0
}

// 是装备魔法
func (c lt_type) IsSpellEquip() bool {
	return (c & LT_SpellEquip) != 0
}

// 是场地魔法
func (c lt_type) IsSpellField() bool {
	return (c & LT_SpellField) != 0
}

// 是速攻魔法
func (c lt_type) IsSpellQuickPlay() bool {
	return (c & LT_SpellQuickPlay) != 0
}

// 是普通陷阱
func (c lt_type) IsTrapNormal() bool {
	return (c & LT_TrapNormal) != 0
}

// 是永续陷阱
func (c lt_type) IsSustainsTrap() bool {
	return (c & LT_TrapContinuous) != 0
}

// 是反击陷阱
func (c lt_type) IsTrapCounter() bool {
	return (c & LT_TrapCounter) != 0
}

// 怪兽属性 Attributes
type la_type uint32

const (
	LA_None la_type = 0

	LA_Earth  la_type = 1 << (32 - 1 - iota) //地
	LA_Water                                 //水
	LA_Fire                                  //火
	LA_Wind                                  //风
	LA_Light                                 //光
	LA_Dark                                  //暗
	LA_Devine                                //神
)

var la = map[la_type]string{
	LA_None:   "NoneAttr",
	LA_Earth:  "Earth",  //地
	LA_Water:  "Water",  //水
	LA_Fire:   "Fire",   //火
	LA_Wind:   "Wind",   //风
	LA_Light:  "Light",  //光
	LA_Dark:   "Dark",   //暗
	LA_Devine: "Devine", //神
}

func (c la_type) String() (s string) {
	if c != LA_None {
		s = la[c]
		if s != "" {
			return
		}
		for k, v := range la {
			if (k & c) != 0 {
				s += v
				s += ""
			}
		}
		return
	}
	return la[c]
}

//地
func (c la_type) IsEarth() bool {
	return c == LA_Earth
}

//水
func (c la_type) IsWater() bool {
	return c == LA_Water
}

//火
func (c la_type) IsFire() bool {
	return c == LA_Fire
}

//风
func (c la_type) IsWind() bool {
	return c == LA_Wind
}

//光
func (c la_type) IsLight() bool {
	return c == LA_Light
}

//暗
func (c la_type) IsDark() bool {
	return c == LA_Dark
}

//神
func (c la_type) IsDevine() bool {
	return c == LA_Devine
}

// 怪兽种族 Races
type lr_type uint32

const (
	LR_None lr_type = 0

	LR_Warrior      lr_type = 1 << (32 - 1 - iota) //战士族
	LR_Spellcaster                                 //魔法使用族
	LR_Fairy                                       //精灵族 天使族
	LR_Fiend                                       //恶魔族
	LR_Zombie                                      //不死族
	LR_Machine                                     //机械族
	LR_Water                                       //水族
	LR_Fire                                        //炎族
	LR_Rock                                        //岩石族
	LR_WingedBeast                                 //鸟兽族
	LR_Plant                                       //植物族
	LR_Insect                                      //昆虫族
	LR_Thunder                                     //雷族
	LR_Dragon                                      //龙族
	LR_Beast                                       //兽族
	LR_BeastWarrior                                //兽战士族
	LR_Dinosaur                                    //恐龙族
	LR_Fish                                        //鱼族
	LR_SeaSerpent                                  //海龙族
	LR_Reptile                                     //爬虫族
	LR_Psychic                                     //念动力族
	LR_DivineBeast                                 //幻神兽族
	//LR_Fairy     //天使族
	//LR_CreatorGod //创造神族
	//LR_DivineBeast   //幻龙族
)

var lr = map[lr_type]string{
	LR_None:         "NoneRaces",
	LR_Warrior:      "Warrior",      //战士族
	LR_Spellcaster:  "Spellcaster",  //魔法使用族
	LR_Fairy:        "Fairy",        //精灵族 天使族
	LR_Fiend:        "Fiend",        //恶魔族
	LR_Zombie:       "Zombie",       //不死族
	LR_Machine:      "Machine",      //机械族
	LR_Water:        "Water",        //水族
	LR_Fire:         "Fire",         //炎族
	LR_Rock:         "Rock",         //岩石族
	LR_WingedBeast:  "WingedBeast",  //鸟兽族
	LR_Plant:        "Plant",        //植物族
	LR_Insect:       "Insect",       //昆虫族
	LR_Thunder:      "Thunder",      //雷族
	LR_Dragon:       "Dragon",       //龙族
	LR_Beast:        "Beast",        //兽族
	LR_BeastWarrior: "BeastWarrior", //兽战士族
	LR_Dinosaur:     "Dinosaur",     //恐龙族
	LR_Fish:         "Fish",         //鱼族
	LR_SeaSerpent:   "SeaSerpent",   //海龙族
	LR_Reptile:      "Reptile",      //爬虫族
	LR_Psychic:      "Psychic",      //念动力族
	LR_DivineBeast:  "DivineBeast",  //幻神兽族
}

func (c lr_type) String() (s string) {
	if c != LR_None {
		s = lr[c]
		if s != "" {
			return
		}
		for k, v := range lr {
			if (k & c) != 0 {
				s += v
				s += ""
			}
		}
		return
	}
	return lr[c]
}

//战士族
func (c lr_type) IsWarrior() bool {
	return (c & LR_Warrior) != LR_None
}

//魔法使用族
func (c lr_type) IsSpellcaster() bool {
	return (c & LR_Spellcaster) != LR_None
}

//精灵族 天使族
func (c lr_type) IsFairy() bool {
	return (c & LR_Fairy) != LR_None
}

//恶魔族
func (c lr_type) IsFiend() bool {
	return (c & LR_Fiend) != LR_None
}

//不死族
func (c lr_type) IsZombie() bool {
	return (c & LR_Zombie) != LR_None
}

//机械族
func (c lr_type) IsMachine() bool {
	return (c & LR_Machine) != LR_None
}

//水族
func (c lr_type) IsWater() bool {
	return (c & LR_Water) != LR_None
}

//炎族
func (c lr_type) IsFire() bool {
	return (c & LR_Fire) != LR_None
}

//岩石族
func (c lr_type) IsRock() bool {
	return (c & LR_Rock) != LR_None
}

//鸟兽族
func (c lr_type) IsWingedBeast() bool {
	return (c & LR_WingedBeast) != LR_None
}

//植物族
func (c lr_type) IsPlant() bool {
	return (c & LR_Plant) != LR_None
}

//昆虫族
func (c lr_type) IsInsect() bool {
	return (c & LR_Insect) != LR_None
}

//雷族
func (c lr_type) IsThunder() bool {
	return (c & LR_Thunder) != LR_None
}

//龙族
func (c lr_type) IsDragon() bool {
	return (c & LR_Dragon) != LR_None
}

//兽族
func (c lr_type) IsBeast() bool {
	return (c & LR_Beast) != LR_None
}

//兽战士族
func (c lr_type) IsBeastWarrior() bool {
	return (c & LR_BeastWarrior) != LR_None
}

//恐龙族
func (c lr_type) IsDinosaur() bool {
	return (c & LR_Dinosaur) != LR_None
}

//鱼族
func (c lr_type) IsFish() bool {
	return (c & LR_Fish) != LR_None
}

//海龙族
func (c lr_type) IsSeaSerpent() bool {
	return (c & LR_SeaSerpent) != LR_None
}

//爬虫族
func (c lr_type) IsReptile() bool {
	return (c & LR_Reptile) != LR_None
}

//念动力族
func (c lr_type) IsPsychic() bool {
	return (c & LR_Psychic) != LR_None
}

//幻神兽族
func (c lr_type) IsDivineBeast() bool {
	return (c & LR_DivineBeast) != LR_None
}

// 表示形式 Expression
type le_type uint32

const (
	LE_None le_type = 0

	LE_FaceUp   le_type = 1 << (32 - 1 - iota) // 正面朝上
	LE_FaceDown                                // 正面朝下
	LE_Attack                                  // 攻击状态
	LE_Defense                                 // 守备状态

	LE_FaceUpAttack    le_type = LE_FaceUp | LE_Attack    // 朝上攻击
	LE_FaceDownAttack  le_type = LE_FaceDown | LE_Attack  // 朝下攻击
	LE_FaceUpDefense   le_type = LE_FaceUp | LE_Defense   // 朝上防御
	LE_FaceDownDefense le_type = LE_FaceDown | LE_Defense // 朝下防御

	LE_ad   le_type = LE_Attack | LE_Defense
	LE_fd   le_type = LE_FaceUp | LE_FaceDown
	LE_peek le_type = 1
)

var le = map[le_type]string{
	LE_None:            "NoneExpre",
	LE_FaceUp:          "FaceUp",   // 正面朝上
	LE_FaceDown:        "FaceDown", // 正面朝下
	LE_Attack:          "Attack",   // 攻击状态
	LE_Defense:         "Defense",  // 守备状态
	LE_FaceUpAttack:    "FaceUpAttack",
	LE_FaceDownAttack:  "FaceDownAttack",
	LE_FaceUpDefense:   "FaceUpDefense",
	LE_FaceDownDefense: "FaceDownDefense",
}

func (c le_type) String() (s string) {
	return le[c&(LE_ad|LE_fd)]
}

// 判断是攻击表示
func (c le_type) IsAttack() bool {
	return (c & LE_Attack) == LE_Attack
}

// 判断是防御表示
func (c le_type) IsDefense() bool {
	return (c & LE_Defense) == LE_Defense
}

// 判断是面朝
func (c le_type) IsFaceUp() bool {
	return (c & LE_FaceUp) == LE_FaceUp
}

// 判断是面朝下
func (c le_type) IsFaceDown() bool {
	return (c & LE_FaceDown) == LE_FaceDown
}

// 判断是面朝上攻击表示
func (c le_type) IsFaceUpAttack() bool {
	return (c & LE_FaceUpAttack) == LE_FaceUpAttack
}

// 判断是面朝下攻击表示
func (c le_type) IsFaceDownAttack() bool {
	return (c & LE_FaceDownAttack) == LE_FaceDownAttack
}

// 判断是面朝上防御表示
func (c le_type) IsFaceUpDefense() bool {
	return (c & LE_FaceUpDefense) == LE_FaceUpDefense
}

// 判断是面朝下防御表示
func (c le_type) IsFaceDownDefense() bool {
	return (c & LE_FaceDownDefense) == LE_FaceDownDefense
}

// 手牌主动方法 Initiative
type li_type uint32

const (
	LI_None li_type = iota

	LI_Use1   = 1   // 使用
	LI_Use2   = 2   // 覆盖
	LI_Yes    = 10  // 是
	LI_No     = 11  // 否
	LI_Defeat = 666 // 认输
	LI_Over   = 101 // 鼠标悬浮
	LI_Out    = 102 // 鼠标离开

)

// 召唤方式 Summon
//type ls_type uint32

//const (
//	LS_None ls_type = 0

//	LS_Normal   ls_type = 1 << (32 - 1 - iota) // 通常
//	LS_Advance                                 // 上级
//	LS_Dual                                    // 二重
//	LS_Flip                                    // 翻转
//	LS_Special                                 // 特殊
//	LS_Fusion                                  // 融合
//	LS_Ritual                                  // 仪式
//	LS_Synchro                                 // 同调
//	LS_Excess                                  // 超量
//	LS_Pendulum                                // 摇摆
//)

// 游戏阶段 Phase
type lp_type uint32

const (
	//LP_None    lp_type = 0
	LP_Chain   lp_type = iota // 连锁
	LP_Draw                   // 抽牌
	LP_Standby                // 预备
	LP_Main1                  // 主阶段1
	LP_Battle                 // 战斗
	LP_Main2                  // 主阶段2
	LP_End                    // 结束

	LP_Damage    // 战斗
	LP_DamageCal // 战斗计算
)

var lp = map[lp_type]string{
	//LP_None:    "NonePhase",
	LP_Chain:   "Chain",   // 连锁
	LP_Draw:    "Draw",    // 抽牌
	LP_Standby: "Standby", // 预备
	LP_Main1:   "Main1",   // 主阶段1
	LP_Battle:  "Battle",  // 战斗
	LP_Main2:   "Main2",   // 主阶段2
	LP_End:     "End",     // 结束
}

func (c lp_type) String() (s string) {

	s = lp[c]
	if s != "" {
		return
	}
	for k, v := range lp {
		if (k & c) != 0 {
			s += v
			s += ""
		}
	}
	return
}

// 卡牌放置位置 Locations
type ll_type string

const (
	LL_None ll_type = ""

	LL_Deck     ll_type = "deck"     // 卡组
	LL_Hand     ll_type = "hand"     // 手牌
	LL_Mzone    ll_type = "mzone"    // 怪兽区
	LL_Szone    ll_type = "szone"    // 魔陷区
	LL_Grave    ll_type = "grave"    // 墓地
	LL_Removed  ll_type = "removed"  // 移除
	LL_Extra    ll_type = "extra"    // 额外
	LL_Field    ll_type = "field"    // 场地
	LL_OverLay  ll_type = "overLay"  //
	LL_Fzone    ll_type = "fzone"    //
	LL_Pzone    ll_type = "pzone"    //
	LL_Portrait ll_type = "portrait" // 玩家头像
)

func (c ll_type) String() (s string) {
	return string(c)
}

// 卡牌操作提示 Operation
type lo_type string

const (
	LO_None lo_type = ""

	LO_BP  lo_type = "BP"
	LO_MP2 lo_type = "MP2"
	LO_EP  lo_type = "EP"

	LO_Chain               lo_type = "Chain"
	LO_Onset               lo_type = "Onset"
	LO_Cover               lo_type = "Cover"
	LO_Discard             lo_type = "Discard"
	LO_Attack              lo_type = "Attack"
	LO_Target              lo_type = "Target"
	LO_Select              lo_type = "Select"
	LO_Destroy             lo_type = "Destroy"
	LO_DestroyAny          lo_type = "DestroyAny"
	LO_DestroyMonster      lo_type = "DestroyMonster"
	LO_DestroySpell        lo_type = "DestroySpell"
	LO_DestroyTrap         lo_type = "DestroyTrap"
	LO_DestroySpellAndTrap lo_type = "DestroySpellAndTrap"
	LO_SummonMonster       lo_type = "SummonMonster"
	LO_Freedom             lo_type = "Freedom"
	LO_Cost                lo_type = "Cost"
	LO_Popup               lo_type = "Popup"
	LO_Summon              lo_type = "Summon"
	LO_SummonFreedom       lo_type = "SummonFreedom"
	LO_SummonSpecial       lo_type = "SummonSpecial"
	LO_SummonFlip          lo_type = "SummonFlip"
	LO_SummonFusion        lo_type = "SummonFusion"
	LO_Flip                lo_type = "Flip"
	LO_SetAttack           lo_type = "SetAttack"
	LO_SetDefense          lo_type = "SetDefense"
	LO_CoverFreedom        lo_type = "CoverFreedom"
	LO_JoinHand            lo_type = "JoinHand"
	LO_Puppet              lo_type = "Puppet"
	LO_Equip               lo_type = "Equip"
	LO_Fusion              lo_type = "Fusion"
	LO_JoinDeckBot         lo_type = "JoinDeckBot"
	LO_JoinDeckTop         lo_type = "JoinDeckTop"
	LO_PutBack             lo_type = "PutBack"
	LO_Reply               lo_type = "Reply"
	LO_Hurt                lo_type = "Hurt"
	LO_Fence               lo_type = "Fence"
	LO_DrawCard            lo_type = "DrawCard"
	LO_RemovedMonster      lo_type = "RemovedMonster"
	LO_Removed             lo_type = "Removed"
	LO_TempChange          lo_type = "TempChange"
	LO_Mitigation          lo_type = "Mitigation"
	LO_Peek                lo_type = "Peek"
	LO_LetDiscard          lo_type = "LetDiscard"
	LO_Intercept           lo_type = "Intercept"
)
