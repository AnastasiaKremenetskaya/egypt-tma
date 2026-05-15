package transport

import (
	"fmt"
	"strconv"
	"strings"
)

type CallbackKind string

const (
	KindVote    CallbackKind = "vote"
	KindSeth    CallbackKind = "seth"
	KindDraw    CallbackKind = "draw"
	KindVoice   CallbackKind = "voice"
	KindStart   CallbackKind = "start_game"
	KindUnknown CallbackKind = ""
)

type Callback struct {
	Kind     CallbackKind
	RoomCode string
	Value    int // vote: 1=trust/0=lie; seth: answer index; draw/voice: unused
}

func Parse(data string) (Callback, error) {
	parts := strings.SplitN(data, ":", 3)
	if len(parts) < 2 {
		return Callback{}, fmt.Errorf("invalid callback: %q", data)
	}

	cb := Callback{
		Kind:     CallbackKind(parts[0]),
		RoomCode: parts[1],
	}

	if len(parts) == 3 {
		v, err := strconv.Atoi(parts[2])
		if err != nil {
			return Callback{}, fmt.Errorf("invalid value in callback %q: %w", data, err)
		}
		cb.Value = v
	}

	return cb, nil
}

func VoteData(roomCode string, trust bool) string {
	v := 0
	if trust {
		v = 1
	}
	return fmt.Sprintf("vote:%s:%d", roomCode, v)
}

func SethData(roomCode string, idx int) string {
	return fmt.Sprintf("seth:%s:%d", roomCode, idx)
}

func DrawData(roomCode string) string {
	return fmt.Sprintf("draw:%s", roomCode)
}

func VoiceData(roomCode string) string {
	return fmt.Sprintf("voice:%s", roomCode)
}

func StartGameData(roomCode string) string {
	return fmt.Sprintf("start_game:%s", roomCode)
}
