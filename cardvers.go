package ygo_core

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

func (cv *CardVersion) ListIsAll() (u []uint) {
	for k, _ := range cv.List {
		u = append(u, k)
	}
	return
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

func (cv *CardVersion) String() string {
	return ""
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

func (cv *CardVersion) IsExist(id uint) bool {
	return cv.List[id] != nil
}

func (cv *CardVersion) Register(co *CardOriginal) {
	if co == nil {
		Debug("RegisterCard: Nil")
	} else if co.Name == "" {
		Debug("RegisterCard: Name is empty")
	} else if co.Id == 0 || cv.List[co.Id] != nil {
		Debug("RegisterCard: Duplicate ID")
	} else {
		cv.List[co.Id] = co
		return
	}
	Debug(co)
	//cv.nameReg += fmt.Sprintf("~~~%s~~~", co.Name)
	//cv.nameList[co.Name] = co

	return
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

	for _, v := range player.decks.main {
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
