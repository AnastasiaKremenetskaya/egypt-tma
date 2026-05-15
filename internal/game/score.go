package game

const WinScore = 50

// AddScore adjusts a player's score and returns the new value.
func (r *Room) AddScore(userID int64, delta int) int {
	p := r.FindPlayer(userID)
	if p == nil {
		return 0
	}
	p.Score += delta
	return p.Score
}

// CheckWinner returns the first player who reached WinScore, or nil.
func (r *Room) CheckWinner() *Player {
	for i := range r.Players {
		if r.Players[i].Score >= WinScore {
			return &r.Players[i]
		}
	}
	return nil
}

// LeaderBoard returns players sorted by score descending (copy).
func (r *Room) LeaderBoard() []Player {
	out := make([]Player, len(r.Players))
	copy(out, r.Players)
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Score > out[i].Score {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
