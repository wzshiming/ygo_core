package ygo_core

type event string

const (
	// 玩家事件  注意玩家事件没有 之前之后 的前缀符
	RoundBegin = "RoundBegin" // 回合开始之前
	DP         = "DP"         // 抽排阶段
	SP         = "SP"         // 预备阶段
	MP         = "MP"         // 主要阶段
	BP         = "BP"         // 战斗阶段
	EP         = "EP"         // 结束阶段
	RoundEnd   = "RoundEnd"   // 回合结束之后
	Chain      = "Chain"      // 连锁
	Draw       = "Draw"       // 每次抽排
	DrawNum    = "DrawNum"    // 每抽一张牌
	ChangeLp   = "ChangeLp"   // 生命值改变
	expres     = "expres"     // 改变表示形式 由用户指定主阶段的怪兽卡发出 由于卡牌效果也能触发改变表示形式的事件  这个独立出来做用户触发用

	Use1 = "Use1" // 用户按钮1
	Use2 = "Use2" // 用户按钮2

	//标识符
	Pre  = "Rre"  // 之前
	Suf  = "Suf"  // 之后
	Bear = "Bear" // 被
	In   = "In"   // 进入
	Out  = "Out"  // 离开

	// 主阶段事件
	Cover         = "Cover"         // 覆盖
	Onset         = "Onset"         // 主动发动
	Summon        = "Summon"        // 召唤
	SummonFlip    = "SummonFlip"    // 反转召唤
	SummonSpecial = "SummonSpecial" // 特殊召唤
	//特殊召唤 子事件
	SummonFusion = "SummonFusion" // 融合召唤

	// 战斗阶段事件
	Declaration = "Declaration" // 攻击宣言
	DamageStep  = "DamageStep"  // 伤害步骤

	// 怪兽事件
	Flip       = "Flip"       // 反转
	Deduct     = "Deduct"     // 对玩家造成伤害
	Fought     = "Fought"     // 战斗步骤结束对双方怪兽发出
	Expres     = "Expres"     // 改变表示形式
	FaceUp     = "FaceUp"     // 表侧表示  召唤 特殊召唤 翻转 翻转召唤
	BearAttack = "BearAttack" // 在伤害计算前向被攻击的怪兽发出
	Change     = "Change"     // 怪兽卡牌属性发生变化时

	// 多种进墓地和除外形式
	Removed         = "Removed"         // 移除
	Cost            = "Cost"            // 花费
	Disabled        = "Disabled"        // 失效
	Destroy         = "Destroy"         // 破坏 送去墓地
	DestroyBeBattle = "DestroyBeBattle" // 战斗破坏
	DestroyBeRule   = "DestroyBeRule"   // 规则破坏
	Discard         = "Discard"         // 丢弃
	Depleted        = "Depleted"        // 使用完的
	Freedom         = "Freedom"         // 解放

	// 卡牌事件
	InDeck     = In + string(LL_Deck)     // 进入卡组
	OutDeck    = Out + string(LL_Deck)    // 离开卡组
	InHand     = In + string(LL_Hand)     // 进入手牌
	OutHand    = Out + string(LL_Hand)    // 离开手牌
	InMzone    = In + string(LL_Mzone)    // 进入怪兽区
	OutMzone   = Out + string(LL_Mzone)   // 离开怪兽区
	InSzone    = In + string(LL_Szone)    // 进入魔陷区
	OutSzone   = Out + string(LL_Szone)   // 离开魔陷区
	InGrave    = In + string(LL_Grave)    // 进入墓地
	OutGrave   = Out + string(LL_Grave)   // 离开墓地
	InRemoved  = In + string(LL_Removed)  // 进入移除
	OutRemoved = Out + string(LL_Removed) // 离开移除
	InExtra    = In + string(LL_Extra)    // 进入额外
	OutExtra   = Out + string(LL_Extra)   // 离开额外
	InField    = In + string(LL_Field)    // 进入场地
	OutField   = Out + string(LL_Field)   // 离开场地

	//BearDestroy = Bear + Destroy // 卡牌被破坏时向被破坏的卡牌发出

	Trigger                   = "Trigger"                // 特殊的连锁诱发事件
	UnregisterAllGlobalListen = "UnregisterGlobalListen" // 注销全局事件监听

	// 由卡牌发出
	UseSpell = "UseSpell" // 使用魔法卡
	UseTrap  = "UseTrap"  // 使用陷阱卡

)
