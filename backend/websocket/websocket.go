package websocket

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"slices"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/mograby3500/mini-discord/cmd/api/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Message represents a chat message stored in the database
type Message struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	ChannelID int       `bson:"channel_id" json:"channel_id"`
	UserID    int       `bson:"user_id" json:"user_id"`
	UserName  string    `bson:"user_name,omitempty" json:"user_name,omitempty"`
	Content   string    `bson:"content" json:"content"`
	Type      string    `bson:"type" json:"type"`
	ServerId  int       `bson:"server_id" json:"server_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Hub struct {
	clients    map[int]map[int]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

type Client struct {
	conn     *websocket.Conn
	userID   int
	UserName string
	send     chan Message
	servers  []int
	channels []int
}

type WebsocketHandler struct {
	MongoDB *mongo.Client
	Hub     *Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust for production
	},
}

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
		handleWebSocket(h.MongoDB, h.Hub, w, r)
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

			var channelIDs []int
			err = db.Select(&channelIDs, `
				SELECT c.id
				FROM   channels c
				JOIN   user_servers us ON c.server_id = us.server_id
				WHERE  us.user_id = $1
			`, client.userID)
			if err != nil {
				log.Println("Database error (channels):", err)
				return
			}
			client.channels = channelIDs

			var userName string
			err = db.Get(&userName, `
				SELECT username FROM users WHERE id = $1
			`, client.userID)
			if err != nil {
				log.Println("Database error:", err)
				continue
			}
			client.UserName = userName

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

func handleWebSocket(mongDB *mongo.Client, hub *Hub, w http.ResponseWriter, r *http.Request) {
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
	go client.writeMessages(hub)
	client.readMessages(mongDB, hub)
}

// writeMessages sends messages to the client
func (c *Client) writeMessages(hub *Hub) {
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
func (c *Client) readMessages(mongoDB *mongo.Client, hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	collection := mongoDB.Database(os.Getenv("MONGO_DB")).Collection("messages")

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
			UserName:  c.UserName,
			Content:   msg.Content,
			Type:      "text",
			ServerId:  msg.ServerID,
			CreatedAt: time.Now(),
		}

		authorized := slices.Contains(c.channels, message.ChannelID)
		if !authorized {
			log.Println("User not authorized to send message to channel")
			continue
		}

		res, err := collection.InsertOne(context.Background(), message)
		if err != nil {
			log.Println("MongoDB insert error:", err)
			continue
		}
		if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
			message.ID = oid.Hex()
		}

		hub.broadcast <- message
	}
}

func (hub *Hub) DeleteServer(serverID int, chanelIDs []int) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	if _, exists := hub.clients[serverID]; !exists {
		return
	}

	for _, client := range hub.clients[serverID] {
		client.servers = slices.DeleteFunc(client.servers, func(s int) bool {
			return s == serverID
		})
		client.channels = slices.DeleteFunc(client.channels, func(c int) bool {
			return slices.Contains(chanelIDs, c)
		})
		delete(hub.clients[serverID], client.userID)
	}
	delete(hub.clients, serverID)
}
