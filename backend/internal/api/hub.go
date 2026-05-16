package api

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anterekhova/egypt-tma/internal/game"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 512
)

// Hub manages WebSocket connections grouped by room code.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*wsConn]struct{}
	log   *log.Logger
}

type wsConn struct {
	ws   *websocket.Conn
	send chan []byte
}

func NewHub(logger *log.Logger) *Hub {
	return &Hub{
		rooms: make(map[string]map[*wsConn]struct{}),
		log:   logger,
	}
}

func (h *Hub) register(roomCode string, ws *websocket.Conn) *wsConn {
	c := &wsConn{ws: ws, send: make(chan []byte, 64)}
	h.mu.Lock()
	if h.rooms[roomCode] == nil {
		h.rooms[roomCode] = make(map[*wsConn]struct{})
	}
	h.rooms[roomCode][c] = struct{}{}
	h.mu.Unlock()
	return c
}

func (h *Hub) unregister(roomCode string, c *wsConn) {
	h.mu.Lock()
	if room := h.rooms[roomCode]; room != nil {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, roomCode)
		}
	}
	h.mu.Unlock()
	close(c.send)
}

// BroadcastRoom serialises the room and pushes the state to all WS clients in the room.
func (h *Hub) BroadcastRoom(room *game.Room, sethAnsweredIDs []int64) {
	state := RoomStateFrom(room, sethAnsweredIDs)
	data, err := json.Marshal(WSMessage{Type: "state", State: state})
	if err != nil {
		h.log.Printf("hub marshal: %v", err)
		return
	}

	h.mu.RLock()
	conns := h.rooms[room.Code]
	h.mu.RUnlock()

	for c := range conns {
		select {
		case c.send <- data:
		default:
			// slow client — skip frame
		}
	}
}

// ServeWS upgrades an HTTP connection to WebSocket and manages its lifecycle.
func (h *Hub) ServeWS(ws *websocket.Conn, roomCode string) {
	c := h.register(roomCode, ws)
	defer h.unregister(roomCode, c)

	go c.writePump()
	c.readPump() // blocks until connection closes
}

func (c *wsConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *wsConn) readPump() {
	defer c.ws.Close()
	c.ws.SetReadLimit(maxMsgSize)
	_ = c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := c.ws.ReadMessage(); err != nil {
			return
		}
	}
}
