package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anterekhova/egypt-tma/internal/game"
	"github.com/anterekhova/egypt-tma/internal/transport"
)

func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.api.Send(msg); err != nil {
		b.log.Printf("send to %d: %v", chatID, err)
	}
}

func (b *Bot) sendMarkup(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = markup
	if _, err := b.api.Send(msg); err != nil {
		b.log.Printf("send markup to %d: %v", chatID, err)
	}
}

func (b *Bot) broadcast(room *game.Room, text string) {
	for _, p := range room.Players {
		b.sendText(p.UserID, text)
	}
}

// notifyLobby sends the lobby state to all players.
func (b *Bot) notifyLobby(room *game.Room) {
	admin := room.FindPlayer(room.AdminID)
	adminName := ""
	if admin != nil {
		adminName = "@" + admin.Username
	}

	botUsername := b.api.Self.UserName

	invite := fmt.Sprintf(
		"🪶 Игра Богов начнётся в Храме %s!\n\n"+
			"📜 Перешли это приглашение достойным.\n\n"+
			"👉 <a href=\"https://t.me/%s?start=%s\">Войти в Храм</a>",
		room.Code, botUsername, room.Code,
	)

	var players strings.Builder
	players.WriteString("<b>Участники:</b>\n")
	for _, p := range room.Players {
		players.WriteString(fmt.Sprintf("  • %s <i>«%s»</i>\n", p.Username, p.Title))
	}

	for _, p := range room.Players {
		if p.UserID == room.AdminID {
			markup := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("⚖️ Начать суд", transport.StartGameData(room.Code)),
				),
			)
			b.sendMarkup(p.UserID, invite+"\n\n"+players.String(), markup)
		} else {
			text := invite + "\n\n" + players.String() +
				fmt.Sprintf("\n<i>Игра начнётся когда %s начнёт её.</i>", adminName)
			b.sendText(p.UserID, text)
		}
	}
}

// notifyQuestion sends the drawn card to the active player and a spectator notice to others.
func (b *Bot) notifyQuestion(room *game.Room) {
	q := room.CurrentQuestion
	if q == nil {
		return
	}
	active := room.ActivePlayer()
	if active == nil {
		return
	}

	// Active player: card title matches type
	var cardTitle string
	if q.Type == "maat" {
		cardTitle = "🕊️ <b>Папирус Маат</b>"
	} else {
		cardTitle = "⚡ <b>Карта Сета</b>"
	}
	activeText := fmt.Sprintf(
		"%s\n«%s»\n\n💬 Если хочешь ответить текстом — просто напиши.\n🗣️ Если ответил вслух — нажми кнопку ниже.\n⏱️ У тебя 60 секунд.",
		cardTitle, q.Text,
	)
	activeMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗣️ Я ответил вслух", transport.VoiceData(room.Code)),
		),
	)

	// Spectators: only see that the active player received a question
	spectatorText := fmt.Sprintf(
		"👁️ <b>Глаз Гора наблюдает</b>\nИгрок @%s получил вопрос:\n«%s»\nЖди его ответа…",
		active.Username, q.Text,
	)

	for _, p := range room.Players {
		if p.UserID == active.UserID {
			b.sendMarkup(p.UserID, activeText, activeMarkup)
		} else {
			b.sendText(p.UserID, spectatorText)
		}
	}
}

// notifyVoting broadcasts the answer and sends vote buttons to non-active players.
func (b *Bot) notifyVoting(room *game.Room) {
	active := room.ActivePlayer()
	if active == nil {
		return
	}
	ans := room.CurrentAnswer
	var ansText string
	if ans == nil || ans.Type == "voice" {
		ansText = "<i>(ответил вслух)</i>"
	} else {
		ansText = fmt.Sprintf("«%s»", ans.Text)
	}

	text := fmt.Sprintf(
		"👂 <b>%s «%s»</b> отвечает:\n%s\n\n⚖️ Голосуйте! 30 секунд.",
		active.Username, active.Title, ansText,
	)

	voteMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🪶 Перо Маат (чисто)", transport.VoteData(room.Code, true)),
			tgbotapi.NewInlineKeyboardButtonData("😤 Гнев Аммит (лжёт)", transport.VoteData(room.Code, false)),
		),
	)

	for _, p := range room.Players {
		if p.UserID == active.UserID {
			b.sendText(p.UserID, text+"\n\n<i>Другие судят тебя...</i>")
		} else {
			b.sendMarkup(p.UserID, text, voteMarkup)
		}
	}
}

// notifyVoteResult broadcasts the vote outcome.
func (b *Bot) notifyVoteResult(room *game.Room, trust, lie int, majority bool) {
	active := room.ActivePlayer()
	if active == nil {
		return
	}
	var result string
	if majority {
		result = fmt.Sprintf("🪶 Большинство считает <b>%s «%s»</b> честным! +3 очка.", active.Username, active.Title)
	} else {
		result = fmt.Sprintf("😤 Большинство поймало <b>%s «%s»</b> на лжи! Штрафной вопрос Сета.", active.Username, active.Title)
	}
	b.broadcast(room, fmt.Sprintf("%s\n\nЗа: %d | Против: %d", result, trust, lie))
}

// notifySeth sends the Seth penalty question with answer options.
func (b *Bot) notifySeth(room *game.Room) {
	q := room.CurrentQuestion
	if q == nil || q.Type != "seth" {
		return
	}

	text := fmt.Sprintf("🌪 <b>Вопрос Сета!</b>\n\n%s\n\n<i>20 секунд. Активный: верно +2. Остальные: верно +1, неверно −1.</i>", q.Text)

	var rows [][]tgbotapi.InlineKeyboardButton
	for i, opt := range q.Options {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(opt, transport.SethData(room.Code, i)),
		))
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	b.broadcast(room, text)
	// send buttons separately to all players (they all answer Seth)
	for _, p := range room.Players {
		b.sendMarkup(p.UserID, "Выберите ответ:", markup)
	}
}

// notifySethResult broadcasts who answered correctly.
func (b *Bot) notifySethResult(room *game.Room, correct string, scores map[int64]int) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ Правильный ответ: <b>%s</b>\n\n", correct))
	for _, p := range room.Players {
		delta := scores[p.UserID]
		sign := "+"
		if delta < 0 {
			sign = ""
		}
		sb.WriteString(fmt.Sprintf("  %s «%s»: %s%d (итого %d)\n", p.Username, p.Title, sign, delta, p.Score))
	}
	b.broadcast(room, sb.String())
}

// notifyScoreboard shows current scores.
func (b *Bot) notifyScoreboard(room *game.Room) {
	lb := room.LeaderBoard()
	var sb strings.Builder
	sb.WriteString("📊 <b>Счёт:</b>\n")
	for i, p := range lb {
		sb.WriteString(fmt.Sprintf("  %d. %s «%s» — %d очков\n", i+1, p.Username, p.Title, p.Score))
	}
	b.broadcast(room, sb.String())
}

// notifyWinner sends the end-game message.
func (b *Bot) notifyWinner(room *game.Room, winner *game.Player) {
	text := fmt.Sprintf(
		"🏆 <b>%s «%s»</b> набрал(а) %d очков и победил(а)!\n\nСуд Осириса завершён. Маат торжествует.",
		winner.Username, winner.Title, winner.Score,
	)
	b.broadcast(room, text)
	b.notifyScoreboard(room)
}

// notifyTimeout tells everyone the active player ran out of time.
func (b *Bot) notifyTimeout(room *game.Room) {
	active := room.ActivePlayer()
	if active == nil {
		return
	}
	b.broadcast(room, fmt.Sprintf(
		"⏳ Время вышло! <b>%s «%s»</b> пропускает ход. −1 очко.",
		active.Username, active.Title,
	))
}
