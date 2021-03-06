package ygo_core

import (
	"fmt"

	"github.com/wzshiming/dispatcher"
)

type Group struct {
	Cards
	dispatcher.Events
	name  ll_type
	owner *Player // 所有者
}

func NewGroup(pl *Player, name ll_type) *Group {
	return &Group{
		Events: dispatcher.NewForkEvent(pl.GetFork()),
		owner:  pl,
		name:   name,
		Cards:  *NewCards(),
	}
}

func (cp *Group) String() string {
	return fmt.Sprintf("%s(%d)", cp.GetName(), cp.Len())
}

func (cp *Group) GetOwner() *Player {
	return cp.owner
}

//func (cp *Group) SetName(name ll_type) {
//	cp.name = name
//}

func (cp *Group) GetName() ll_type {
	return cp.name
}

func (cp *Group) BeginPush(c *Card) {
	cp.Insert(c, cp.Len())
}

func (cp *Group) EndPush(c *Card) {
	cp.Insert(c, 0)
}

func (cp *Group) Insert(c *Card, index int) {
	c.Placed()
	c.GetSummoner().callAll(moveCard(c, cp.GetName()))
	c.place = cp
	cp.Cards.Insert(c, index)
	c.Dispatch(In + string(cp.GetName()))
}

func (cp *Group) BeginPop() (c *Card) {
	return cp.Remove(cp.Len() - 1)
}

func (cp *Group) EndPop() (c *Card) {
	return cp.Remove(0)
}

func (cp *Group) EndPeek(i int) *Cards {
	cs := NewCards()
	if i <= 0 {
		return cs
	}
	l := cp.Len()
	if i > l {
		i = l
	}
	for j := 0; j != i; j++ {
		cs.EndPush(cp.Cards[j])
	}
	return cs
}

func (cp *Group) BeginPeek(i int) *Cards {
	cs := NewCards()
	if i <= 0 {
		return cs
	}
	l := cp.Len()
	if i > l {
		i = l
	}

	for j := 0; j != i; j++ {
		cs.EndPush(cp.Cards[l-j-1])
	}
	return cs
}

func (cp *Group) Remove(index int) (c *Card) {
	c = cp.Cards.Remove(index)
	if c != nil {
		c.place = nil
		c.lastPlace = cp
		c.Dispatch(Out + string(cp.GetName()))
	}
	return
}

func (cp *Group) PickedForUniq(uniq uint) (c *Card) {
	c = cp.Cards.PickedForUniq(uniq)
	if c != nil {
		c.place = nil
		c.lastPlace = cp
		c.Dispatch(Out + string(cp.GetName()))
	}
	return
}

func (cp *Group) PickedFor(c *Card) {
	if c != nil {
		cp.PickedForUniq(c.ToUint())
	}
}

// 是在场地
func (p *Group) IsField() bool {
	return p != nil && p.GetName() == LL_Field
}

// 是在卡组
func (p *Group) IsDeck() bool {
	return p != nil && p.GetName() == LL_Deck
}

// 是在额外
func (p *Group) IsExtra() bool {
	return p != nil && p.GetName() == LL_Extra
}

// 是在墓地
func (p *Group) IsGrave() bool {
	return p != nil && p.GetName() == LL_Grave
}

// 是在手牌
func (p *Group) IsHand() bool {
	return p != nil && p.GetName() == LL_Hand
}

// 是在怪兽区
func (p *Group) IsMzone() bool {
	return p != nil && p.GetName() == LL_Mzone
}

// 是在魔陷区
func (p *Group) IsSzone() bool {
	return p != nil && p.GetName() == LL_Szone
}

// 是在手牌
func (p *Group) IsRemoved() bool {
	return p != nil && p.GetName() == LL_Removed
}

// 是头像
func (p *Group) IsPortrait() bool {
	return p != nil && p.GetName() == LL_Portrait
}
