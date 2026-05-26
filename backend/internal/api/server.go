package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// GameActions is the interface the API server uses to drive game state.
// Implemented by *bot.Bot to keep logic in one place.
type GameActions interface {
	APICreateRoom(userID int64, username string) (*RoomState, error)
	APIJoinRoom(code string, userID int64, username string) (*RoomState, error)
	APIStartGame(code string, userID int64) (*RoomState, error)
	APIAnswer(code string, userID int64, text string) (*RoomState, error)
	APIVoice(code string, userID int64) (*RoomState, error)
	APIVote(code string, userID int64, trust bool) (*RoomState, error)
	APISeth(code string, userID int64, optIdx int) (*RoomState, error)
	APIGetRoom(code string) (*RoomState, error)
	APIFinishGame(code string, userID int64) (*RoomState, error)
	APILeaveGame(code string, userID int64) error
}

// Server is the HTTP API + WebSocket server for the Mini App.
type Server struct {
	game      GameActions
	hub       *Hub
	botToken  string
	webAppURL string // allowed CORS origin
	devMode   bool   // if true, accepts "dev:{json}" initData without HMAC (local dev only)
	log       *log.Logger
	upgrader  websocket.Upgrader
}

func NewServer(game GameActions, hub *Hub, botToken, webAppURL string, devMode bool, logger *log.Logger) *Server {
	s := &Server{
		game:      game,
		hub:       hub,
		botToken:  botToken,
		webAppURL: webAppURL,
		devMode:   devMode,
		log:       logger,
	}
	s.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if webAppURL == "" {
				return true // dev mode: allow all
			}
			origin := r.Header.Get("Origin")
			return origin == webAppURL
		},
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ws/room/{code}", s.handleWS)
	mux.HandleFunc("GET /api/room/{code}", s.handleGetRoom)
	mux.HandleFunc("POST /api/room", s.handleCreateRoom)
	mux.HandleFunc("POST /api/room/{code}/join", s.handleJoin)
	mux.HandleFunc("POST /api/room/{code}/start", s.handleStart)
	mux.HandleFunc("POST /api/room/{code}/answer", s.handleAnswer)
	mux.HandleFunc("POST /api/room/{code}/voice", s.handleVoice)
	mux.HandleFunc("POST /api/room/{code}/vote", s.handleVote)
	mux.HandleFunc("POST /api/room/{code}/seth", s.handleSeth)
	mux.HandleFunc("POST /api/room/{code}/finish", s.handleFinish)
	mux.HandleFunc("POST /api/room/{code}/leave", s.handleLeave)

	return s.cors(mux)
}

// ─── Middleware ────────────────────────────────────────────────────────────────

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := s.webAppURL
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Init-Data")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// auth extracts and validates the Telegram user from the request.
// initData can be in the X-Telegram-Init-Data header or the init_data query param (for WS).
//
// Dev mode: when DEV_MODE=true the server also accepts initData in the form
//
//	dev:{"id":111111,"username":"player1"}
//
// This lets you test locally without a real Telegram session. NEVER enable in production.
func (s *Server) auth(r *http.Request) (*TelegramUser, error) {
	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData == "" {
		initData = r.URL.Query().Get("init_data")
	}
	if initData == "" {
		return nil, errors.New("missing init data")
	}

	// Dev-mode shortcut: skip HMAC, parse user directly from JSON payload.
	if s.devMode && strings.HasPrefix(initData, "dev:") {
		payload := strings.TrimPrefix(initData, "dev:")
		var user TelegramUser
		if err := json.Unmarshal([]byte(payload), &user); err != nil {
			return nil, errors.New("dev mode: invalid user JSON")
		}
		if user.ID == 0 {
			return nil, errors.New("dev mode: user ID must be non-zero")
		}
		s.log.Printf("[DEV] auth bypass for user %d (%s)", user.ID, usernameOf(&user))
		return &user, nil
	}

	return ValidateInitData(initData, s.botToken)
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_ = user

	code := r.PathValue("code")
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Printf("ws upgrade: %v", err)
		return
	}

	// Send current state immediately on connect.
	if state, err := s.game.APIGetRoom(code); err == nil {
		if data, err := json.Marshal(WSMessage{Type: "state", State: *state}); err == nil {
			_ = ws.WriteMessage(websocket.TextMessage, data)
		}
	}

	s.hub.ServeWS(ws, code)
}

func (s *Server) handleGetRoom(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	state, err := s.game.APIGetRoom(code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	state, err := s.game.APICreateRoom(user.ID, usernameOf(user))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	code := r.PathValue("code")
	state, err := s.game.APIJoinRoom(code, user.ID, usernameOf(user))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	code := r.PathValue("code")
	state, err := s.game.APIStartGame(code, user.ID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleAnswer(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	code := r.PathValue("code")
	state, err := s.game.APIAnswer(code, user.ID, body.Text)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleVoice(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	code := r.PathValue("code")
	state, err := s.game.APIVoice(code, user.ID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Trust bool `json:"trust"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	code := r.PathValue("code")
	state, err := s.game.APIVote(code, user.ID, body.Trust)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleSeth(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Option int `json:"option"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	code := r.PathValue("code")
	state, err := s.game.APISeth(code, user.ID, body.Option)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleFinish(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	code := r.PathValue("code")
	state, err := s.game.APIFinishGame(code, user.ID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, state)
}

func (s *Server) handleLeave(w http.ResponseWriter, r *http.Request) {
	user, err := s.auth(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	code := r.PathValue("code")
	if err := s.game.APILeaveGame(code, user.ID); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
