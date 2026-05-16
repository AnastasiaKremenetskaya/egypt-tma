package game

import "sync"

var voteMu sync.Mutex

// AddVote records a vote atomically. Returns false if already voted.
func (r *Room) AddVote(voterID int64, trust bool) bool {
	voteMu.Lock()
	defer voteMu.Unlock()
	if _, exists := r.Votes[voterID]; exists {
		return false
	}
	r.Votes[voterID] = trust
	return true
}

// VoteResult returns (trustCount, lieCount, totalExpected).
func (r *Room) VoteResult() (trust, lie, expected int) {
	voteMu.Lock()
	defer voteMu.Unlock()
	for _, v := range r.Votes {
		if v {
			trust++
		} else {
			lie++
		}
	}
	expected = len(r.Players) - 1 // everyone except active
	return
}

// AllVoted returns true when all eligible voters have cast their vote.
func (r *Room) AllVoted() bool {
	trust, lie, expected := r.VoteResult()
	return trust+lie >= expected
}

func (r *Room) ResetVotes() {
	voteMu.Lock()
	defer voteMu.Unlock()
	r.Votes = make(map[int64]bool)
}
