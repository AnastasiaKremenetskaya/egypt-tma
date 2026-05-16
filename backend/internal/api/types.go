package api

import (
	"time"

	"github.com/anterekhova/egypt-tma/internal/game"
)

// RoomState is the sanitised room snapshot sent to Mini App clients.
// It intentionally hides individual vote choices and Seth correct-answer index.
type RoomState struct {
	Code            string        `json:"code"`
	AdminID         int64         `json:"admin_id"`
	Phase           game.Phase    `json:"phase"`
	Players         []PlayerView  `json:"players"`
	ActiveIdx       int           `json:"active_idx"`
	Round           int           `json:"round"`
	Question        *QuestionView `json:"question,omitempty"`
	Answer          *AnswerView   `json:"answer,omitempty"`
	VoteTrust       int           `json:"vote_trust"`
	VoteLie         int           `json:"vote_lie"`
	VotedIDs        []int64       `json:"voted_ids"`
	SethAnsweredIDs []int64       `json:"seth_answered_ids"`
	PhaseDeadline   time.Time     `json:"phase_deadline"`
}

type PlayerView struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Title    string `json:"title"`
	Score    int    `json:"score"`
}

// QuestionView omits the correct answer index during active Seth phase.
type QuestionView struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"` // "maat" | "seth"
	Text    string   `json:"text"`
	Options []string `json:"options,omitempty"`
}

type AnswerView struct {
	Type string `json:"type"` // "text" | "voice"
	Text string `json:"text,omitempty"`
}

// WSMessage wraps a RoomState push over WebSocket.
type WSMessage struct {
	Type  string    `json:"type"` // "state"
	State RoomState `json:"state"`
}

// RoomStateFrom converts a game.Room into a sanitised RoomState.
func RoomStateFrom(room *game.Room, sethAnsweredIDs []int64) RoomState {
	players := make([]PlayerView, len(room.Players))
	for i, p := range room.Players {
		players[i] = PlayerView{
			UserID:   p.UserID,
			Username: p.Username,
			Title:    p.Title,
			Score:    p.Score,
		}
	}

	var q *QuestionView
	if room.CurrentQuestion != nil {
		q = &QuestionView{
			ID:      room.CurrentQuestion.ID,
			Type:    room.CurrentQuestion.Type,
			Text:    room.CurrentQuestion.Text,
			Options: room.CurrentQuestion.Options,
		}
	}

	var ans *AnswerView
	if room.CurrentAnswer != nil {
		ans = &AnswerView{
			Type: room.CurrentAnswer.Type,
			Text: room.CurrentAnswer.Text,
		}
	}

	trust, lie := 0, 0
	votedIDs := make([]int64, 0, len(room.Votes))
	for id, t := range room.Votes {
		votedIDs = append(votedIDs, id)
		if t {
			trust++
		} else {
			lie++
		}
	}

	if sethAnsweredIDs == nil {
		sethAnsweredIDs = []int64{}
	}

	return RoomState{
		Code:            room.Code,
		AdminID:         room.AdminID,
		Phase:           room.Phase,
		Players:         players,
		ActiveIdx:       room.ActiveIdx,
		Round:           room.Round,
		Question:        q,
		Answer:          ans,
		VoteTrust:       trust,
		VoteLie:         lie,
		VotedIDs:        votedIDs,
		SethAnsweredIDs: sethAnsweredIDs,
		PhaseDeadline:   room.PhaseDeadline,
	}
}
