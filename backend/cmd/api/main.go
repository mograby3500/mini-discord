package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mograby3500/mini-discord/cmd/api/auth"
	"github.com/mograby3500/mini-discord/cmd/api/servers"
	"github.com/mograby3500/mini-discord/db"
	"github.com/mograby3500/mini-discord/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	SQLDB   *sqlx.DB
	MongoDB *mongo.Client
	Router  *mux.Router
	Hub     *websocket.Hub
}

func (a *App) Initialize() error {
	pgDB, err := db.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("PostgreSQL connection failed: %w", err)
	}
	a.SQLDB = pgDB

	mongoClient, err := db.ConnectMongo()
	if err != nil {
		return fmt.Errorf("MongoDB connection failed: %w", err)
	}
	a.MongoDB = mongoClient

	a.Hub = websocket.NewHub()
	a.Router = mux.NewRouter()

	authHandler := &auth.Handler{DB: pgDB}
	authHandler.RegisterRoutes(a.Router)

	serverHandler := &servers.ServerHandler{DB: pgDB, MongoDB: mongoClient}
	serverHandler.RegisterRoutes(a.Router)

	websocketHandler := &websocket.WebsocketHandler{MongoDB: mongoClient, Hub: a.Hub}
	websocketHandler.RegisterRoutes(a.Router)

	a.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	go a.Hub.Run(pgDB)
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
