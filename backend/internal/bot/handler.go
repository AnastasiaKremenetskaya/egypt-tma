package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anterekhova/egypt-tma/internal/api"
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
	hub   *api.Hub

	// sethAnswers tracks who answered the Seth question in a given round
	// map[roomCode]map[userID]bool
	sethAnswers map[string]map[int64]bool
}

func New(tgAPI *tgbotapi.BotAPI, st store.Store, bank *questions.Bank, logger *log.Logger) *Bot {
	return &Bot{
		api:         tgAPI,
		store:       st,
		bank:        bank,
		log:         logger,
		ctx:         context.Background(),
		sethAnswers: make(map[string]map[int64]bool),
	}
}

func (b *Bot) SetHub(h *api.Hub) { b.hub = h }

// persist saves the room and broadcasts its state to Mini App WebSocket clients.
func (b *Bot) persist(room *game.Room) {
	if err := b.store.Save(room); err != nil {
		b.log.Printf("persist room %s: %v", room.Code, err)
	}
	b.notifyHub(room)
}

// notifyHub broadcasts the current room state to all connected Mini App clients.
func (b *Bot) notifyHub(room *game.Room) {
	if b.hub == nil {
		return
	}
	ids := make([]int64, 0, len(b.sethAnswers[room.Code]))
	for id := range b.sethAnswers[room.Code] {
		ids = append(ids, id)
	}
	b.hub.BroadcastRoom(room, ids)
}

// currentState returns the current sanitised RoomState for a room code.
func (b *Bot) currentState(code string) (*api.RoomState, error) {
	room, err := b.store.Get(code)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(b.sethAnswers[code]))
	for id := range b.sethAnswers[code] {
		ids = append(ids, id)
	}
	s := api.RoomStateFrom(room, ids)
	return &s, nil
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
	case "старт":
		b.cmdStartGame(msg, userID)
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
	b.persist(room)

	b.notifyLobby(room)
}

func (b *Bot) cmdStartGame(msg *tgbotapi.Message, userID int64) {
	room, err := b.store.FindByPlayer(userID)
	if err != nil {
		b.sendText(msg.Chat.ID, "Вы не состоите ни в одной комнате. /new — создать, /join КОД — войти.")
		return
	}
	if room.Phase != game.PhaseLobby {
		b.sendText(msg.Chat.ID, "Игра уже идёт.")
		return
	}
	if len(room.Players) < 2 {
		b.sendText(msg.Chat.ID, "Нужно хотя бы 2 игрока.")
		return
	}

	room.Round = 0
	room.ActiveIdx = 0
	if err := room.Transition(game.PhaseQuestion); err != nil {
		b.log.Printf("transition: %v", err)
		return
	}
	b.persist(room)

	b.broadcast(room, "⚖️ <b>Суд начинается!</b>")
	b.startQuestion(room)
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
	b.persist(room)

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
	b.persist(room)

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
	b.persist(room)
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
	b.persist(room)

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
	b.persist(room)

	// check if all answered
	if len(b.sethAnswers[room.Code]) >= len(room.Players) {
		b.resolveSeth(room)
	}
}

// ─── Phase transitions ────────────────────────────────────────────────────────

func (b *Bot) startQuestion(room *game.Room) {
	room.DrawCard(b.bank)
	room.PhaseDeadline = time.Now().Add(questionDuration)
	b.persist(room)

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
	b.persist(room)

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
	b.persist(room)

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
	b.persist(room)
	b.notifyScoreboard(room)
	b.startQuestion(room)
}

func (b *Bot) endGame(room *game.Room, winner *game.Player) {
	room.StopTimer()
	_ = room.Transition(game.PhaseFinished)
	b.persist(room)
	b.notifyWinner(room, winner)
}

// ─── Public API actions (called by HTTP API server) ───────────────────────────

func (b *Bot) APIGetRoom(code string) (*api.RoomState, error) {
	return b.currentState(code)
}

func (b *Bot) APICreateRoom(userID int64, username string) (*api.RoomState, error) {
	if existing, err := b.store.FindByPlayer(userID); err == nil {
		s := api.RoomStateFrom(existing, nil)
		return &s, nil
	}
	code := generateCode()
	room := game.NewRoom(code, userID, username)
	if err := b.store.Create(room); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	b.notifyHub(room)
	return b.currentState(code)
}

func (b *Bot) APIJoinRoom(code string, userID int64, username string) (*api.RoomState, error) {
	room, err := b.store.Get(code)
	if err != nil {
		return nil, errors.New("комната не найдена")
	}
	if room.Phase != game.PhaseLobby {
		return nil, errors.New("игра уже идёт")
	}
	room.AddPlayer(userID, username) // returns false if already in, that's fine
	b.persist(room)
	b.notifyLobby(room)
	return b.currentState(code)
}

func (b *Bot) APIStartGame(code string, userID int64) (*api.RoomState, error) {
	room, err := b.store.Get(code)
	if err != nil {
		return nil, errors.New("комната не найдена")
	}
	if room.AdminID != userID {
		return nil, errors.New("только организатор может начать игру")
	}
	if room.Phase != game.PhaseLobby {
		return nil, errors.New("игра уже идёт")
	}
	if len(room.Players) < 2 {
		return nil, errors.New("нужно минимум 2 игрока")
	}
	room.Round = 0
	room.ActiveIdx = 0
	if err := room.Transition(game.PhaseQuestion); err != nil {
		return nil, err
	}
	b.persist(room)
	b.broadcast(room, "⚖️ <b>Суд начинается!</b>")
	b.startQuestion(room)
	return b.currentState(code)
}

func (b *Bot) APIAnswer(code string, userID int64, text string) (*api.RoomState, error) {
	room, err := b.store.Get(code)
	if err != nil {
		return nil, errors.New("комната не найдена")
	}
	if room.Phase != game.PhaseQuestion {
		return nil, errors.New("сейчас не фаза вопроса")
	}
	active := room.ActivePlayer()
	if active == nil || active.UserID != userID {
		return nil, errors.New("не твой ход")
	}
	if text == "" {
		return nil, errors.New("ответ не может быть пустым")
	}
	room.CurrentAnswer = &game.AnswerRecord{Type: "text", Text: text}
	b.persist(room)
	b.startVoting(room)
	return b.currentState(code)
}

func (b *Bot) APIVoice(code string, userID int64) (*api.RoomState, error) {
	room, err := b.store.Get(code)
	if err != nil {
		return nil, errors.New("комната не найдена")
	}
	if room.Phase != game.PhaseQuestion {
		return nil, errors.New("сейчас не фаза вопроса")
	}
	active := room.ActivePlayer()
	if active == nil || active.UserID != userID {
		return nil, errors.New("не твой ход")
	}
	room.CurrentAnswer = &game.AnswerRecord{Type: "voice"}
	b.persist(room)
	b.startVoting(room)
	return b.currentState(code)
}

func (b *Bot) APIVote(code string, userID int64, trust bool) (*api.RoomState, error) {
	room, err := b.store.Get(code)
	if err != nil {
		return nil, errors.New("комната не найдена")
	}
	if room.Phase != game.PhaseVoting {
		return nil, errors.New("сейчас не фаза голосования")
	}
	active := room.ActivePlayer()
	if active != nil && active.UserID == userID {
		return nil, errors.New("активный игрок не голосует")
	}
	if !room.AddVote(userID, trust) {
		return nil, errors.New("ты уже проголосовал")
	}
	b.persist(room)
	if room.AllVoted() {
		b.resolveVoting(room)
	}
	return b.currentState(code)
}

func (b *Bot) APISeth(code string, userID int64, optIdx int) (*api.RoomState, error) {
	room, err := b.store.Get(code)
	if err != nil {
		return nil, errors.New("комната не найдена")
	}
	if room.Phase != game.PhaseSeth {
		return nil, errors.New("сейчас не фаза Сета")
	}
	q := room.CurrentQuestion
	if q == nil || q.Type != "seth" {
		return nil, errors.New("нет вопроса Сета")
	}
	if b.sethAnswers[code] == nil {
		b.sethAnswers[code] = make(map[int64]bool)
	}
	if _, already := b.sethAnswers[code][userID]; already {
		return nil, errors.New("ты уже ответил")
	}
	b.sethAnswers[code][userID] = optIdx == q.CorrectOptIdx
	b.persist(room)
	if len(b.sethAnswers[code]) >= len(room.Players) {
		b.resolveSeth(room)
	}
	return b.currentState(code)
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
