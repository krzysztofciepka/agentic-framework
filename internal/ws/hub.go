package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]map[string]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]map[string]bool),
	}
}

func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		http.Error(w, "conversation_id query param required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	if _, ok := h.clients[conn]; !ok {
		h.clients[conn] = make(map[string]bool)
	}
	h.clients[conn][conversationID] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) SendToConversation(conversationID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn, subs := range h.clients {
		if subs[conversationID] {
			conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (h *Hub) SendEvent(conversationID string, event any) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.SendToConversation(conversationID, data)
}
