package blackjack

// Executes a callback function for each player at the table.
func (g *Game) DoForEachPlayer(fn func(*Player)) {
	for _, p := range g.GetSeats() {
		if p != nil {
			fn(p)
		}
	}
}

// Returns the seats in dealing order, including nils.
func (g *Game) GetSeats() []*Player {
	return []*Player{g.Seat1, g.Seat2, g.Seat3}
}

// Returns the of most recent unpaid side bet of a given type, otherwise returns nil.
func latestUnpaidSideBet(bets []*SideBet, t SideBetType) *SideBet {
	for i := len(bets) - 1; i >= 0; i-- {
		sb := bets[i]
		if sb == nil {
			continue
		}
		if sb.Type == t && !sb.Paid {
			return sb
		}
	}
	return nil
}

// Returns true if all players have busted.
func (g *Game) AllPlayersBusted() bool {
	for _, p := range g.GetSeats() {
		if p == nil {
			continue
		}
		for _, h := range p.Hands {
			if h.Status != Busted {
				return false
			}
		}
	}
	return true
}
