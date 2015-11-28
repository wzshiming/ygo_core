package ygo_core

type AskCode struct {
	Uniq   uint
	Method uint
}

type card struct {
	index uint
	size  uint
}

type deck struct {
	main []card
	side []card
}

func NewDeck() *deck {
	return &deck{}
}

func (d *deck) MainAddCard(index uint, size uint) {
	d.main = append(d.main, card{
		index: index,
		size:  size,
	})
	return
}

func (d *deck) SideAddCard(index uint, size uint) {
	d.side = append(d.side, card{
		index: index,
		size:  size,
	})
	return
}

//type session struct {
//	//ask  chan AskCode
//	sess *agent.Session
//	name string
//	deck *deck
//	id   uint
//}

//func NewSession(id uint, name string, deck *deck, sess *agent.Session) *session {
//	s := &session{
//		id: id,
//		//ask:  make(chan AskCode, 16),
//		sess: sess,
//		name: name,
//		deck: deck,
//	}
//	return s
//}

//func (s *session) Ask(method uint, uniq uint) {
//	s.ask <- AskCode{
//		Method: method,
//		Uniq:   uniq,
//	}
//	return
//}

//func (s *session) getAsk(method uint, uniq uint) <-chan AskCode {
//	return s.ask
//}
