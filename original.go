package ygo_core

import (
	"github.com/wzshiming/dispatcher"
	"github.com/wzshiming/ffmt"
)

type CardOriginal struct {
	IsValid  bool    // 是否有效
	Id       uint    // 卡牌id
	Name     string  // 名字
	Password string  // 卡牌密码
	Lt       lt_type // 卡牌类型
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
			"Type": co.Lt,
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
			"Type": co.Lt,
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
