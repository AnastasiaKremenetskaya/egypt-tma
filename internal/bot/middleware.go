package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anterekhova/egypt-tma/internal/game"
)

// userFromMessage extracts userID and username from a message.
func userFromMessage(msg *tgbotapi.Message) (int64, string) {
	if msg == nil || msg.From == nil {
		return 0, ""
	}
	name := msg.From.UserName
	if name == "" {
		name = msg.From.FirstName
	}
	return msg.From.ID, name
}

// userFromCallback extracts userID and username from a callback query.
func userFromCallback(cb *tgbotapi.CallbackQuery) (int64, string) {
	if cb == nil || cb.From == nil {
		return 0, ""
	}
	name := cb.From.UserName
	if name == "" {
		name = cb.From.FirstName
	}
	return cb.From.ID, name
}

// requirePhase returns the room only if it is in one of the allowed phases.
func requirePhase(room *game.Room, phases ...game.Phase) bool {
	for _, p := range phases {
		if room.Phase == p {
			return true
		}
	}
	return false
}
