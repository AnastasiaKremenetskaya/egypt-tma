package game

import (
	"context"
	"time"
)

// Phase is the FSM state of a room.
type Phase string

const (
	PhaseLobby    Phase = "lobby"
	PhaseQuestion Phase = "question" // active player answers
	PhaseVoting   Phase = "voting"   // others vote
	PhaseSeth     Phase = "seth"     // penalty/bonus Seth question
	PhaseFinished Phase = "finished"
)

// QuestionRef holds the current question drawn for a turn.
type QuestionRef struct {
	ID   string `json:"id"`
	Type string `json:"type"` // "maat" | "seth"
	Text string `json:"text"`
	// For Seth questions only
	Answer        string   `json:"answer,omitempty"`
	Options       []string `json:"options,omitempty"`         // 4 shuffled choices shown to players
	CorrectOptIdx int      `json:"correct_opt_idx,omitempty"` // index in Options of the correct answer
}

// AnswerRecord stores what the active player submitted.
type AnswerRecord struct {
	Type string `json:"type"` // "text" | "voice"
	Text string `json:"text,omitempty"`
}

// Player represents one participant.
type Player struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Title    string `json:"title"`
	Score    int    `json:"score"`
}

// Room is the central game state.
type Room struct {
	Code    string   `json:"code"`
	AdminID int64    `json:"admin_id"`
	Phase   Phase    `json:"phase"`
	Players []Player `json:"players"`

	ActiveIdx       int            `json:"active_idx"`
	Round           int            `json:"round"`
	CurrentQuestion *QuestionRef   `json:"current_question,omitempty"`
	CurrentAnswer   *AnswerRecord  `json:"current_answer,omitempty"`
	Votes           map[int64]bool `json:"votes"` // userID → true=trust, false=lie

	PhaseDeadline time.Time `json:"phase_deadline"`

	// runtime-only, not persisted to Redis
	cancelTimer context.CancelFunc
}

func NewRoom(code string, adminID int64, username string) *Room {
	return &Room{
		Code:      code,
		AdminID:   adminID,
		Phase:     PhaseLobby,
		Players:   []Player{{UserID: adminID, Username: username, Title: RandomTitle(adminID), Score: 0}},
		ActiveIdx: 0,
		Round:     0,
		Votes:     make(map[int64]bool),
	}
}

func (r *Room) AddPlayer(userID int64, username string) bool {
	for _, p := range r.Players {
		if p.UserID == userID {
			return false // already in room
		}
	}
	r.Players = append(r.Players, Player{
		UserID:   userID,
		Username: username,
		Title:    RandomTitle(userID),
		Score:    0,
	})
	return true
}

func (r *Room) ActivePlayer() *Player {
	if len(r.Players) == 0 {
		return nil
	}
	return &r.Players[r.ActiveIdx%len(r.Players)]
}

func (r *Room) FindPlayer(userID int64) *Player {
	for i := range r.Players {
		if r.Players[i].UserID == userID {
			return &r.Players[i]
		}
	}
	return nil
}

func (r *Room) PlayerIDs() []int64 {
	ids := make([]int64, len(r.Players))
	for i, p := range r.Players {
		ids[i] = p.UserID
	}
	return ids
}

func (r *Room) NonActivePlayerIDs() []int64 {
	active := r.ActivePlayer()
	var ids []int64
	for _, p := range r.Players {
		if active == nil || p.UserID != active.UserID {
			ids = append(ids, p.UserID)
		}
	}
	return ids
}

func (r *Room) StopTimer() {
	if r.cancelTimer != nil {
		r.cancelTimer()
		r.cancelTimer = nil
	}
}

func (r *Room) SetTimer(cancel context.CancelFunc) {
	r.StopTimer()
	r.cancelTimer = cancel
}
