package ygo_core

import "errors"

//"rego"

type CardRets struct {
	List []CardRet `json:"list"`
}

type CardRet struct {
	Id    uint `json:"id,string"`
	Limit uint `json:"limit,string"`
	State uint `json:"state,string"`
}

type CardVersion struct {
	List map[uint]*CardOriginal
	//nameList map[string]*CardOriginal
	//nameReg string
}

func NewCardVersion() *CardVersion {
	return &CardVersion{
		List: make(map[uint]*CardOriginal),
		//nameList: make(map[string]*CardOriginal),
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

func (cv *CardVersion) ListIsValid() (u []uint) {
	for k, v := range cv.List {
		if v.IsValid {
			u = append(u, k)
		}
	}
	return
}

func (cv *CardVersion) AllIsValid() (cr CardRets) {
	for k, v := range cv.List {
		if v.IsValid {
			cr.List = append(cr.List, CardRet{
				Id:    k,
				Limit: 3,
				State: 1,
			})
		}
	}
	return
}

//func (cv *CardVersion) Filter(name string, i ...interface{}) (cr CardRets) {
//	// weiwan daixu

//	for k, _ := range cv.FindForOriginal(name, true) {
//		cr.List = append(cr.List, CardRet{
//			Id:    k,
//			Limit: 3,
//			State: 1,
//		})
//	}
//	return
//}

//func (cv *CardVersion) FindForOriginal(name string, val bool) (co map[uint]*CardOriginal) {
//	co = make(map[uint]*CardOriginal)
//	reg := regexp.MustCompile(fmt.Sprintf("~([^(~~)]*%s[^(~~)]*)~", name))
//	al := reg.FindAllStringSubmatch(cv.nameReg, -1)
//	for _, v := range al {
//		if len(v) == 2 {
//			if d := cv.nameList[v[1]]; d != nil {
//				if val {
//					if d.IsValid {
//						co[d.Id] = d
//					}
//				} else {
//					co[d.Id] = d
//				}
//			}
//		}
//	}
//	return
//}

//func (cv *CardVersion) Find(name string, val bool) (c []uint) {
//	reg := regexp.MustCompile(fmt.Sprintf("~([^(~~)]*%s[^(~~)]*)~", name))
//	al := reg.FindAllStringSubmatch(cv.nameReg, -1)
//	for _, v := range al {
//		if len(v) == 2 {
//			if d := cv.nameList[v[1]]; d != nil {
//				if val {
//					if d.IsValid {
//						c = append(c, d.Id)
//					}
//				} else {
//					c = append(c, d.Id)
//				}
//			}
//		}
//	}
//	return
//}

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
	//cv.nameReg += fmt.Sprintf("~~~%s~~~", co.Name)
	//cv.nameList[co.Name] = co

	//os.MkdirAll("./img", 0666)
	//os.MkdirAll("./info", 0666)
	//exec.Command("cp", fmt.Sprintf("../static/web/static/cards/img/%v.jpg", co.Id), fmt.Sprintf("./img/%v.jpg", co.Id)).Start()
	//exec.Command("cp", fmt.Sprintf("../static/web/static/cards/i18n/zh-CN/%v.json", co.Id), fmt.Sprintf("./info/%v.json", co.Id)).Start()

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

func (cv *CardVersion) Deck(player *Player) {

	for _, v := range player.Decks.main {
		t := cv.Get(v.index)
		if t != nil {
			for i := uint(0); i != v.size; i++ {
				c := t.Make(player)
				if c.IsExtra() {
					player.Extra().EndPush(c)
				} else {
					player.Deck().EndPush(c)
				}
			}
		}
	}
}
