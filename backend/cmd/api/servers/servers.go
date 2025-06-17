package servers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/mograby3500/mini-discord/cmd/api/auth"
)

type ServerHandler struct {
	DB *sqlx.DB
}

type CreateServerRequest struct {
	Name string `json:"name"`
}

type Channel struct {
	ID        int64     `db:"id" json:"id"`
	ServerID  int64     `db:"server_id" json:"server_id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type ServerWithChannels struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Channels []Channel `json:"channels"`
}

func (h *ServerHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/servers", h.handleCreateServer).Methods("POST")
	router.HandleFunc("/servers", h.handleGetUserServers).Methods("GET")
	router.HandleFunc("/channels", h.handleCreateChannel).Methods("POST")
}

func (h *ServerHandler) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	userID, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTxx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Could not start transaction", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var serverID int64
	err = tx.QueryRow(
		"INSERT INTO servers (name, owner_id) VALUES ($1, $2) RETURNING id",
		request.Name, userID,
	).Scan(&serverID)
	if err != nil {
		http.Error(w, "Failed to create server", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO channels (server_id, name, type) VALUES ($1, $2, $3)", serverID, "text", "text")
	if err != nil {
		http.Error(w, "Failed to create default channel", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO user_servers (user_id, server_id, role) VALUES ($1, $2, $3)", userID, serverID, "owner")
	if err != nil {
		http.Error(w, "Failed to link user to server", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Database commit failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "server created with default channel",
		"server_id": fmt.Sprintf("%d", serverID),
	})
}

func (h *ServerHandler) handleGetUserServers(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	userID, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	var raw []struct {
		ID         int64     `db:"id"`
		ServerID   int64     `db:"server_id"`
		ServerName string    `db:"server_name"`
		Name       string    `db:"name"`
		Type       string    `db:"type"`
		CreatedAt  time.Time `db:"created_at"`
	}
	err = h.DB.Select(&raw, `
		SELECT 
			c.id,
			c.server_id,
			s.name AS server_name,
			c.name,
			c.type,
			c.created_at
		FROM 
			channels c
		JOIN 
			user_servers us ON us.server_id = c.server_id
		JOIN 
			servers s ON s.id = c.server_id
		WHERE 
			us.user_id = $1
	`, userID)

	if err != nil {
		log.Printf("Error fetching channels: %v", err)
		http.Error(w, "Failed to fetch channels", http.StatusInternalServerError)
		return
	}

	// Group by server
	serverMap := make(map[int64]*ServerWithChannels)
	for _, row := range raw {
		if _, exists := serverMap[row.ServerID]; !exists {
			serverMap[row.ServerID] = &ServerWithChannels{
				ID:       row.ServerID,
				Name:     row.ServerName,
				Channels: []Channel{},
			}
		}
		serverMap[row.ServerID].Channels = append(serverMap[row.ServerID].Channels, Channel{
			ID:        row.ID,
			ServerID:  row.ServerID,
			Name:      row.Name,
			Type:      row.Type,
			CreatedAt: row.CreatedAt,
		})
	}

	// Convert map to slice
	result := make([]ServerWithChannels, 0, len(serverMap))
	for _, server := range serverMap {
		result = append(result, *server)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type CreateChannelRequest struct {
	ServerID int64  `json:"server_id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // 'text' or 'voice'
}

func (h *ServerHandler) handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	userID, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var request CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if request.Type != "text" && request.Type != "voice" {
		http.Error(w, "Invalid channel type: must be 'text' or 'voice'", http.StatusBadRequest)
		return
	}
	var exists bool
	err = h.DB.Get(&exists, `
		SELECT EXISTS (
			SELECT 1 FROM user_servers 
			WHERE user_id = $1 AND server_id = $2 AND role IN ('owner', 'admin')
		)
	`, userID, request.ServerID)
	if err != nil {
		http.Error(w, "Failed to verify user permissions", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Forbidden: You are not a member of this server", http.StatusForbidden)
		return
	}

	var channelID int64
	err = h.DB.QueryRow(`
		INSERT INTO channels (server_id, name, type) 
		VALUES ($1, $2, $3) 
		RETURNING id
	`, request.ServerID, request.Name, request.Type).Scan(&channelID)
	if err != nil {
		http.Error(w, "Failed to create channel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "channel created successfully",
		"channel_id": fmt.Sprintf("%d", channelID),
	})
}
