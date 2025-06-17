package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/mograby3500/mini-discord/cmd/api/auth"
)

// Message represents a chat message stored in the database
type Message struct {
	ID        int    `db:"id" json:"id"`
	ServerId  int    `db:"server_id" json:"server_id"`
	ChannelID int    `db:"channel_id" json:"channel_id"`
	UserID    int    `db:"user_id" json:"user_id"`
	Content   string `db:"content" json:"content"`
	Type      string `db:"type" json:"type"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

// Hub manages WebSocket connections
type Hub struct {
	clients    map[int]map[int]*Client // serverID -> userID -> *Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

// Client represents a WebSocket client
type Client struct {
	conn    *websocket.Conn
	userID  int
	send    chan Message
	servers []int
}

type WebsocketHandler struct {
	DB  *sqlx.DB
	Hub *Hub
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
		clients:    make(map[int]map[int]*Client),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *WebsocketHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(h.DB, h.Hub, w, r)
	}).Methods("GET")
}

func (h *Hub) Run(db *sqlx.DB) {
	for {
		select {
		case client := <-h.register:
			var serverIDs []int
			err := db.Select(&serverIDs, `
				SELECT server_id FROM user_servers WHERE user_id = $1
			`, client.userID)
			if err != nil {
				log.Println("Database error:", err)
				continue
			}

			client.servers = serverIDs

			h.mutex.Lock()
			for _, serverID := range serverIDs {
				if h.clients[serverID] == nil {
					h.clients[serverID] = make(map[int]*Client)
				}
				h.clients[serverID][client.userID] = client
			}
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			for _, serverID := range client.servers {
				if userMap, ok := h.clients[serverID]; ok {
					delete(userMap, client.userID)
					if len(userMap) == 0 {
						delete(h.clients, serverID)
					}
				}
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.Lock()
			if userMap, ok := h.clients[message.ServerId]; ok {
				for userID, client := range userMap {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(userMap, userID)
					}
				}
				if len(userMap) == 0 {
					delete(h.clients, message.ServerId)
				}
			}
			h.mutex.Unlock()
		}
	}
}

func handleWebSocket(db *sqlx.DB, hub *Hub, w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	userID, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	client := &Client{
		conn:    conn,
		userID:  int(userID),
		send:    make(chan Message),
		servers: []int{},
	}
	hub.register <- client
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
			Content   string `json:"content"`
			ChannelID int    `json:"channel_id"`
			ServerID  int    `json:"server_id"`
		}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		message := Message{
			ChannelID: msg.ChannelID,
			UserID:    c.userID,
			Content:   msg.Content,
			Type:      "text",
			ServerId:  msg.ServerID,
		}

		var count int
		err = db.Get(&count, `
			SELECT
				COUNT(*)
			FROM
				channels
			JOIN
				user_servers
			ON
				channels.server_id = user_servers.server_id
			WHERE
				channels.id = $1 AND user_servers.user_id = $2
		`, message.ChannelID, c.userID)

		if err != nil {
			log.Println("Database error:", err)
			continue
		}
		if count == 0 {
			log.Println("User not authorized to send message to channel")
			continue
		}

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
