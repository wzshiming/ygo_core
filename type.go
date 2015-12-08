package ygo_core

// 卡牌类型 Card Types
type lc_type uint32

const (
	LC_None lc_type = 0

	LC_MonsterNormal   lc_type = 1 << (32 - 1 - iota) //普通怪兽 黄色
	LC_MonsterEffect                                  //效果怪兽 橙色
	LC_MonsterFusion                                  //融合怪兽 紫色
	LC_MonsterXyz                                     //超量怪兽 黑色
	LC_MonsterSynchro                                 //同调怪兽 白色
	LC_MonsterRitual                                  //仪式怪兽 蓝色
	LC_SpellNormal                                    //普通魔法 通常
	LC_SpellRitual                                    //仪式魔法
	LC_SpellContinuous                                //永续魔法
	LC_SpellEquip                                     //装备魔法
	LC_SpellField                                     //场地魔法
	LC_SpellQuickPlay                                 //速攻魔法 速度2
	LC_TrapNormal                                     //普通陷阱 速度2
	LC_TrapContinuous                                 //永续陷阱 速度2
	LC_TrapCounter                                    //反击陷阱 速度3

	// 怪物卡
	LC_Monster = LC_MonsterNormal | LC_MonsterEffect | LC_MonsterXyz | LC_MonsterSynchro | LC_MonsterFusion | LC_MonsterRitual
	// 魔法卡
	LC_Spell = LC_SpellNormal | LC_SpellRitual | LC_SpellContinuous | LC_SpellEquip | LC_SpellField | LC_SpellQuickPlay
	// 陷阱卡
	LC_Trap = LC_TrapNormal | LC_TrapContinuous | LC_TrapCounter
	// 魔法卡与陷阱卡
	LC_SpellAndTrap = LC_Spell | LC_Trap
)

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

// 表示形式 Expression
type le_type uint32

const (
	LE_None le_type = 0

	LE_FaceUp   le_type = 1 << (32 - 1 - iota) // 正面朝上
	LE_FaceDown                                // 正面朝下
	LE_Attack                                  // 攻击状态
	LE_Defense                                 // 守备状态

	LE_FaceUpAttack    = LE_FaceUp | LE_Attack    // 朝上攻击
	LE_FaceDownAttack  = LE_FaceDown | LE_Attack  // 朝下攻击
	LE_FaceUpDefense   = LE_FaceUp | LE_Defense   // 朝上防御
	LE_FaceDownDefense = LE_FaceDown | LE_Defense // 朝下防御

	LE_ad = LE_Attack | LE_Defense
	LE_fd = LE_FaceUp | LE_FaceDown
)

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
type ls_type uint32

const (
	LS_None ls_type = 0

	LS_Normal   ls_type = 1 << (32 - 1 - iota) // 通常
	LS_Advance                                 // 上级
	LS_Dual                                    // 二重
	LS_Flip                                    // 翻转
	LS_Special                                 // 特殊
	LS_Fusion                                  // 融合
	LS_Ritual                                  // 仪式
	LS_Synchro                                 // 同调
	LS_Excess                                  // 超量
	LS_Pendulum                                // 摇摆
)

// 游戏阶段 Phase
type lp_type uint32

const (
	//LP_None lp_type = 0
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

// 卡牌放置位置 Locations
type ll_type string

const (
	LL_None ll_type = ""

	LL_Deck    ll_type = "deck"    // 卡组
	LL_Hand    ll_type = "hand"    // 手牌
	LL_Mzone   ll_type = "mzone"   // 怪兽区
	LL_Szone   ll_type = "szone"   // 魔陷区
	LL_Grave   ll_type = "grave"   // 墓地
	LL_Removed ll_type = "removed" // 移除
	LL_Extra   ll_type = "extra"   // 额外
	LL_Field   ll_type = "field"   // 场地
	LL_OverLay ll_type = "overLay" //
	LL_Fzone   ll_type = "fzone"   //
	LL_Pzone   ll_type = "pzone"   //

	LL_Portrait ll_type = "portrait" // 玩家头像
)

// 卡牌操作提示 Operation
type lo_type string

const (
	LO_None          lo_type = ""
	LO_Chain         lo_type = "Chain"
	LO_Onset         lo_type = "Onset"
	LO_Cover         lo_type = "Cover"
	LO_Discard       lo_type = "Discard"
	LO_Attack        lo_type = "Attack"
	LO_Target        lo_type = "Target"
	LO_Select        lo_type = "Select"
	LO_Destroy       lo_type = "Destroy"
	LO_Freedom       lo_type = "Freedom"
	LO_Cost          lo_type = "Cost"
	LO_Expres        lo_type = "Expres"
	LO_Popup         lo_type = "Popup"
	LO_Summon        lo_type = "Summon"
	LO_SummonFreedom lo_type = "SummonFreedom"
	LO_SummonSpecial lo_type = "SummonSpecial"
	LO_CoverFreedom  lo_type = "CoverFreedom"
	LO_Removed       lo_type = "Removed"
	LO_JoinHand      lo_type = "JoinHand"
	LO_Puppet        lo_type = "Puppet"
	LO_Equip         lo_type = "Equip"
	LO_Fusion        lo_type = "Fusion"
)
