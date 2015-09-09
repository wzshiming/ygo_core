package ygo_cord

import (
	"errors"
	"fmt"
	"regexp"
	//"rego"
)

type CardVersion struct {
	List     map[uint]*CardOriginal
	nameList map[string]*CardOriginal
	nameReg  string
}

func NewCardVersion() *CardVersion {
	return &CardVersion{
		List:     make(map[uint]*CardOriginal),
		nameList: make(map[string]*CardOriginal),
	}
}

func (cv *CardVersion) Keys() (c []uint) {
	for k, _ := range cv.List {
		c = append(c, k)
	}
	return
}

func (cv *CardVersion) Get(id uint) *CardOriginal {
	return cv.List[id]
}

func (cv *CardVersion) Find(name string, val bool) (c []uint) {
	reg := regexp.MustCompile(fmt.Sprintf("~([^(~~)]*%s[^(~~)]*)~", name))
	al := reg.FindAllStringSubmatch(cv.nameReg, -1)
	for _, v := range al {
		if len(v) == 2 {
			if d := cv.nameList[v[1]]; d != nil {
				if val {
					if d.IsValid {
						c = append(c, d.Id)
					}
				} else {
					c = append(c, d.Id)
				}

			}
		}
	}
	return
}

func (cv *CardVersion) Register(co *CardOriginal) error {
	if co == nil {
		return errors.New("RegisterCard: Nil")
	}
	if co.Name == "" {
		return errors.New("RegisterCard: Name is empty")
	}
	if co.Id == 0 || cv.List[co.Id] != nil {
		return errors.New("RegisterCard: Duplicate ID")
	}
	cv.List[co.Id] = co
	cv.nameReg += fmt.Sprintf("~~~%s~~~", co.Name)
	cv.nameList[co.Name] = co

	//os.MkdirAll("./img", 0666)
	//exec.Command("cp", fmt.Sprintf("../web/static/cards/img/%v.jpg", co.Id), fmt.Sprintf("./img/%v.jpg", co.Id)).Start()

	return nil
}

func (cv *CardVersion) Sum(cv2 *CardVersion) *CardVersion {
	rcv := NewCardVersion()
	for _, v := range cv.List {
		cv.Register(v)
	}
	for _, v := range cv2.List {
		cv.Register(v)
	}
	return rcv
}

func (cv *CardVersion) Deck(cp *Group, player *Player, deck []uint) {
	for _, v := range deck {
		t := cv.Get(v)
		if t != nil {
			c := t.Make(player)
			cp.EndPush(c)
		}
	}
}
