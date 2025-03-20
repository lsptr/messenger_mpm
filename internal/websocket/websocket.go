package websocket

import (
	"MessengerMPM/internal/auth"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешить все источники
	},
}

// Client представляет подключенного клиента.
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID int    // ID пользователя
	chatID string // ID чата
}

// Hub управляет всеми подключенными клиентами.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

// NewHub создает новый Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run запускает Hub.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:

			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// HandleWebSocket обрабатывает WebSocket-соединения.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	log.Println("WebSocket: входящее соединение")
	// Получаем токен из запроса
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Token is required", http.StatusUnauthorized)
		return
	}

	// Проверяем токен и извлекаем user_id
	userID, err := auth.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Получаем chat_id из запроса
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	// Устанавливаем WebSocket-соединение
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	// Создаем клиента
	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		chatID: chatID,
	}

	h.register <- client
	go client.writePump()
	go client.readPump(h)
}

// writePump отправляет сообщения клиенту.
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Failed to write message:", err)
			break
		}
	}
}

// readPump читает сообщения от клиента.
func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message:", err)
			break
		}

		h.broadcast <- message
	}
}

// отправляет сообщение
func (h *Hub) SendToUser(message []byte, chatID string, userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if client.chatID == chatID && client.userID == userID {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}
