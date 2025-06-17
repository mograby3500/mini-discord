package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mograby3500/mini-discord/cmd/api/auth"
	"github.com/mograby3500/mini-discord/cmd/api/servers"
	"github.com/mograby3500/mini-discord/websocket"
)

type App struct {
	DB     *sqlx.DB
	Router *mux.Router
	Hub    *websocket.Hub
}

type CreateServerRequest struct {
	Name string `json:"name"`
}

func (a *App) Initialize() error {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		dbUser, dbPassword, dbName, dbHost, dbPort, dbSSLMode)

	var db *sqlx.DB
	var err error
	maxRetries := 10
	for i := range maxRetries {
		db, err = sqlx.Connect("postgres", connStr)
		if err == nil {
			break
		}
		log.Printf("Database not ready, retrying in 3 seconds... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
	}
	a.DB = db
	a.Hub = websocket.NewHub()
	a.Router = mux.NewRouter()

	authHandler := &auth.Handler{DB: db}
	authHandler.RegisterRoutes(a.Router)

	serverHandler := &servers.ServerHandler{DB: db}
	serverHandler.RegisterRoutes(a.Router)

	websocketHandler := &websocket.WebsocketHandler{DB: db, Hub: a.Hub}
	websocketHandler.RegisterRoutes(a.Router)

	// Register other routes
	a.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	go a.Hub.Run(db)
	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	app := &App{}
	if err := app.Initialize(); err != nil {
		log.Fatal(err)
	}
	handler := corsMiddleware(app.Router)
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
