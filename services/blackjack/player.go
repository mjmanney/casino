package main

type PlayerStatus uint8

const (
	Active PlayerStatus = iota
	Idle
)

type Player struct {
	ID     string // unique identifier or seat label
	Name   string
	Hand   Hand
	Bet    int64
	Status PlayerStatus
}

func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:   id,
		Name: name,
		Hand: Hand{
			cards:  []Card{},
			Status: Qualified,
		},
		Bet:    0,
		Status: Active,
	}
}

func (p *Player) Active() { p.Status = Active }
func (p *Player) Idle()   { p.Status = Idle }
func (p *Player) Reset()  { p.Hand.cards = p.Hand.cards[:0]; p.Status = Active }
