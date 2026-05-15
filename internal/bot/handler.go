package bot

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anterekhova/egypt-tma/internal/game"
	"github.com/anterekhova/egypt-tma/internal/questions"
	"github.com/anterekhova/egypt-tma/internal/store"
	"github.com/anterekhova/egypt-tma/internal/transport"
)

const (
	questionDuration = 60 * time.Second
	votingDuration   = 30 * time.Second
	sethDuration     = 20 * time.Second
)

type Bot struct {
	api   *tgbotapi.BotAPI
	store store.Store
	bank  *questions.Bank
	log   *log.Logger
	ctx   context.Context

	// sethAnswers tracks who answered the Seth question in a given round
	// map[roomCode]map[userID]bool
	sethAnswers map[string]map[int64]bool
}

func New(api *tgbotapi.BotAPI, st store.Store, bank *questions.Bank, logger *log.Logger) *Bot {
	return &Bot{
		api:         api,
		store:       st,
		bank:        bank,
		log:         logger,
		ctx:         context.Background(),
		sethAnswers: make(map[string]map[int64]bool),
	}
}

// HandleUpdate routes incoming Telegram updates.
func (b *Bot) HandleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		b.handleMessage(update.Message)
	case update.CallbackQuery != nil:
		b.handleCallback(update.CallbackQuery)
	}
}

// ─── Message handlers ────────────────────────────────────────────────────────

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	userID, username := userFromMessage(msg)
	if userID == 0 {
		return
	}

	switch {
	case msg.IsCommand():
		b.handleCommand(msg, userID, username)
	default:
		b.handleTextAnswer(msg, userID)
	}
}

func (b *Bot) handleCommand(msg *tgbotapi.Message, userID int64, username string) {
	switch msg.Command() {
	case "start":
		b.cmdStart(msg, userID, username)
	case "new", "новая_комната":
		b.cmdNewRoom(msg, userID, username)
	case "join", "войти":
		b.cmdJoin(msg, userID, username)
	case "rooms":
		b.cmdRooms(msg, userID)
	default:
		b.sendText(msg.Chat.ID, "Неизвестная команда. /new — новая комната, /join ANKH42 — войти в комнату.")
	}
}

func (b *Bot) cmdStart(msg *tgbotapi.Message, userID int64, username string) {
	// deep link: /start ROOMCODE — join room directly
	if code := strings.TrimSpace(msg.CommandArguments()); code != "" {
		b.joinRoom(msg.Chat.ID, strings.ToUpper(code), userID, username)
		return
	}
	b.sendText(msg.Chat.ID, fmt.Sprintf(
		"⚖️ <b>Маат и Сет</b>\n\nДобро пожаловать, %s!\n\n/new — создать комнату\n/join КОД — войти в комнату",
		username,
	))
}

func (b *Bot) cmdNewRoom(msg *tgbotapi.Message, userID int64, username string) {
	// check if already in a room
	if existing, err := b.store.FindByPlayer(userID); err == nil {
		b.sendText(msg.Chat.ID, fmt.Sprintf("Вы уже в комнате <code>%s</code>.", existing.Code))
		return
	}

	code := generateCode()
	room := game.NewRoom(code, userID, username)
	if err := b.store.Create(room); err != nil {
		b.sendText(msg.Chat.ID, "Ошибка создания комнаты.")
		b.log.Printf("create room: %v", err)
		return
	}

	b.notifyLobby(room)
}

func (b *Bot) cmdJoin(msg *tgbotapi.Message, userID int64, username string) {
	parts := strings.Fields(msg.CommandArguments())
	if len(parts) == 0 {
		b.sendText(msg.Chat.ID, "Укажите код комнаты: /join ANKH42")
		return
	}
	b.joinRoom(msg.Chat.ID, strings.ToUpper(parts[0]), userID, username)
}

func (b *Bot) joinRoom(chatID int64, code string, userID int64, username string) {
	room, err := b.store.Get(code)
	if err != nil {
		b.sendText(chatID, fmt.Sprintf("Комната <code>%s</code> не найдена.", code))
		return
	}
	if room.Phase != game.PhaseLobby {
		b.sendText(chatID, "Войти после старта нельзя.")
		return
	}

	if !room.AddPlayer(userID, username) {
		b.sendText(chatID, "Вы уже в этой комнате.")
		return
	}
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	b.notifyLobby(room)
}

func (b *Bot) cmdRooms(msg *tgbotapi.Message, userID int64) {
	room, err := b.store.FindByPlayer(userID)
	if err != nil {
		b.sendText(msg.Chat.ID, "Вы не состоите ни в одной комнате.")
		return
	}
	b.sendText(msg.Chat.ID, fmt.Sprintf("Вы в комнате <code>%s</code>, фаза: %s.", room.Code, room.Phase))
}

// handleTextAnswer processes a free-text answer from the active player during PhaseQuestion.
func (b *Bot) handleTextAnswer(msg *tgbotapi.Message, userID int64) {
	room, err := b.store.FindByPlayer(userID)
	if err != nil {
		return
	}
	if room.Phase != game.PhaseQuestion {
		return
	}
	active := room.ActivePlayer()
	if active == nil || active.UserID != userID {
		return
	}

	room.CurrentAnswer = &game.AnswerRecord{Type: "text", Text: msg.Text}
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	b.startVoting(room)
}

// ─── Callback handlers ────────────────────────────────────────────────────────

func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
	userID, _ := userFromCallback(cb)
	if userID == 0 {
		return
	}

	// always ack
	ack := tgbotapi.NewCallback(cb.ID, "")
	if _, err := b.api.Request(ack); err != nil {
		b.log.Printf("ack callback: %v", err)
	}

	parsed, err := transport.Parse(cb.Data)
	if err != nil {
		b.log.Printf("parse callback %q: %v", cb.Data, err)
		return
	}

	room, err := b.store.Get(parsed.RoomCode)
	if err != nil {
		return
	}

	switch parsed.Kind {
	case transport.KindStart:
		b.cbStartGame(room, userID)
	case transport.KindVoice:
		b.cbVoiceAnswer(room, userID)
	case transport.KindVote:
		b.cbVote(room, userID, parsed.Value == 1)
	case transport.KindSeth:
		b.cbSethAnswer(room, userID, parsed.Value)
	}
}

func (b *Bot) cbStartGame(room *game.Room, userID int64) {
	if room.AdminID != userID {
		return
	}
	if !requirePhase(room, game.PhaseLobby) {
		return
	}
	if len(room.Players) < 2 {
		b.sendText(userID, "Нужно хотя бы 2 игрока.")
		return
	}

	room.Round = 0
	room.ActiveIdx = 0
	if err := room.Transition(game.PhaseQuestion); err != nil {
		b.log.Printf("transition: %v", err)
		return
	}
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	b.broadcast(room, "⚖️ <b>Суд начинается!</b>")
	b.startQuestion(room)
}

func (b *Bot) cbVoiceAnswer(room *game.Room, userID int64) {
	if !requirePhase(room, game.PhaseQuestion) {
		return
	}
	active := room.ActivePlayer()
	if active == nil || active.UserID != userID {
		return
	}

	room.CurrentAnswer = &game.AnswerRecord{Type: "voice"}
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}
	b.startVoting(room)
}

func (b *Bot) cbVote(room *game.Room, userID int64, trust bool) {
	if !requirePhase(room, game.PhaseVoting) {
		return
	}
	active := room.ActivePlayer()
	if active != nil && active.UserID == userID {
		return // active player cannot vote
	}

	if !room.AddVote(userID, trust) {
		return // already voted
	}
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	if room.AllVoted() {
		b.resolveVoting(room)
	}
}

func (b *Bot) cbSethAnswer(room *game.Room, userID int64, optIdx int) {
	if !requirePhase(room, game.PhaseSeth) {
		return
	}
	q := room.CurrentQuestion
	if q == nil || q.Type != "seth" {
		return
	}

	// record answer (once per player per Seth phase)
	if b.sethAnswers[room.Code] == nil {
		b.sethAnswers[room.Code] = make(map[int64]bool)
	}
	if _, already := b.sethAnswers[room.Code][userID]; already {
		return
	}

	correct := optIdx == q.CorrectOptIdx
	b.sethAnswers[room.Code][userID] = correct
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	// check if all answered
	if len(b.sethAnswers[room.Code]) >= len(room.Players) {
		b.resolveSeth(room)
	}
}

// ─── Phase transitions ────────────────────────────────────────────────────────

func (b *Bot) startQuestion(room *game.Room) {
	room.DrawCard(b.bank)
	room.PhaseDeadline = time.Now().Add(questionDuration)
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	b.notifyQuestion(room)

	cancel := game.StartTimer(b.ctx, questionDuration, func() {
		r, err := b.store.Get(room.Code)
		if err != nil || r.Phase != game.PhaseQuestion {
			return
		}
		b.notifyTimeout(r)
		active := r.ActivePlayer()
		if active != nil {
			r.AddScore(active.UserID, -1)
		}
		b.advanceNextTurn(r)
	})
	room.SetTimer(cancel)
}

func (b *Bot) startVoting(room *game.Room) {
	room.StopTimer()
	if err := room.Transition(game.PhaseVoting); err != nil {
		b.log.Printf("transition to voting: %v", err)
		return
	}
	room.PhaseDeadline = time.Now().Add(votingDuration)
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	b.notifyVoting(room)

	cancel := game.StartTimer(b.ctx, votingDuration, func() {
		r, err := b.store.Get(room.Code)
		if err != nil || r.Phase != game.PhaseVoting {
			return
		}
		b.resolveVoting(r)
	})
	room.SetTimer(cancel)
}

func (b *Bot) resolveVoting(room *game.Room) {
	room.StopTimer()
	trust, lie, _ := room.VoteResult()
	majorityTrust := trust >= lie // tie goes to trust

	b.notifyVoteResult(room, trust, lie, majorityTrust)

	active := room.ActivePlayer()
	if majorityTrust {
		if active != nil {
			room.AddScore(active.UserID, 3)
		}
		if winner := room.CheckWinner(); winner != nil {
			b.endGame(room, winner)
			return
		}
		b.advanceNextTurn(room)
		return
	}

	// lie majority → Seth penalty for everyone
	b.startSeth(room)
}

func (b *Bot) startSeth(room *game.Room) {
	if err := room.Transition(game.PhaseSeth); err != nil {
		b.log.Printf("transition to seth: %v", err)
		return
	}
	// draw a fresh Seth card
	room.DrawSethCard(b.bank)
	room.PhaseDeadline = time.Now().Add(sethDuration)
	delete(b.sethAnswers, room.Code)
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}

	b.notifySeth(room)

	cancel := game.StartTimer(b.ctx, sethDuration, func() {
		r, err := b.store.Get(room.Code)
		if err != nil || r.Phase != game.PhaseSeth {
			return
		}
		b.resolveSeth(r)
	})
	room.SetTimer(cancel)
}

func (b *Bot) resolveSeth(room *game.Room) {
	room.StopTimer()
	q := room.CurrentQuestion
	if q == nil {
		b.advanceNextTurn(room)
		return
	}

	answers := b.sethAnswers[room.Code]
	active := room.ActivePlayer()
	scores := make(map[int64]int, len(room.Players))

	for _, p := range room.Players {
		correct := answers[p.UserID]
		var delta int
		if p.UserID == active.UserID {
			if correct {
				delta = 2
			} else {
				delta = 0
			}
		} else {
			if correct {
				delta = 1
			} else {
				delta = -1
			}
		}
		room.AddScore(p.UserID, delta)
		scores[p.UserID] = delta
	}

	b.notifySethResult(room, q.Answer, scores)

	if winner := room.CheckWinner(); winner != nil {
		b.endGame(room, winner)
		return
	}
	b.advanceNextTurn(room)
}

func (b *Bot) advanceNextTurn(room *game.Room) {
	room.NextTurn()
	if err := room.Transition(game.PhaseQuestion); err != nil {
		b.log.Printf("transition to question: %v", err)
		return
	}
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}
	b.notifyScoreboard(room)
	b.startQuestion(room)
}

func (b *Bot) endGame(room *game.Room, winner *game.Player) {
	room.StopTimer()
	_ = room.Transition(game.PhaseFinished)
	if err := b.store.Save(room); err != nil {
		b.log.Printf("save room: %v", err)
	}
	b.notifyWinner(room, winner)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

const codeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = codeChars[rand.Intn(len(codeChars))]
	}
	return string(b)
}
