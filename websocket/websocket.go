package websocket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/mograby3500/mini-discord/cmd/api/auth"
)

// Message represents a chat message stored in the database
type Message struct {
	ID        int    `db:"id"`
	ChannelID int    `db:"channel_id"`
	UserID    int    `db:"user_id"`
	Content   string `db:"content"`
	CreatedAt string `db:"created_at"`
}

// Hub manages WebSocket connections
type Hub struct {
	clients    map[int]map[*Client]bool // channel_id -> clients
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

// Client represents a WebSocket client
type Client struct {
	conn      *websocket.Conn
	channelID int
	userID    int
	send      chan Message
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust for production
	},
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the Hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if h.clients[client.channelID] == nil {
				h.clients[client.channelID] = make(map[*Client]bool)
			}
			h.clients[client.channelID][client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if clients, ok := h.clients[client.channelID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, client.channelID)
				}
				close(client.send)
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients[message.ChannelID] {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients[message.ChannelID], client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// HandleWebSocket upgrades an HTTP request to a WebSocket connection
func HandleWebSocket(db *sqlx.DB, hub *Hub, w http.ResponseWriter, r *http.Request) {
	// extract token from Authorization header
	tokenStr := r.Header.Get("Authorization")
	userID, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	channelIDStr := r.URL.Query().Get("channel_id")
	// Parse channel ID
	channelID := 0
	fmt.Sscanf(channelIDStr, "%d", &channelID)
	if channelID == 0 {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Create client
	client := &Client{
		conn:      conn,
		channelID: channelID,
		userID:    int(userID),
		send:      make(chan Message),
	}

	// Register client
	hub.register <- client

	// Start read/write goroutines
	go client.writeMessages(db, hub)
	client.readMessages(db, hub)
}

// writeMessages sends messages to the client
func (c *Client) writeMessages(db *sqlx.DB, hub *Hub) {
	defer func() {
		c.conn.Close()
		hub.unregister <- c
	}()

	for message := range c.send {
		err := c.conn.WriteJSON(message)
		if err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

// readMessages receives messages from the client
func (c *Client) readMessages(db *sqlx.DB, hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg struct {
			Content string `json:"content"`
		}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		message := Message{
			ChannelID: c.channelID,
			UserID:    c.userID,
			Content:   msg.Content,
		}

		// Store message in database
		err = db.QueryRow(
			"INSERT INTO messages (channel_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at",
			message.ChannelID, message.UserID, message.Content,
		).Scan(&message.ID, &message.CreatedAt)
		if err != nil {
			log.Println("Database error:", err)
			continue
		}

		// Broadcast message
		hub.broadcast <- message
	}
}
