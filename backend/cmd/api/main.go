package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mograby3500/mini-discord/websocket"
	"golang.org/x/crypto/bcrypt"

	"github.com/mograby3500/mini-discord/cmd/api/auth"
)

type App struct {
	DB     *sqlx.DB
	Router *mux.Router
	Hub    *websocket.Hub
}

type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

type CreateServerRequest struct {
	Name string `json:"name"` // The 'name' field in the JSON body will be mapped to this struct field
}

func (a *App) Initialize() error {
	// Load environment variables from .env file
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// Construct PostgreSQL connection string
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		dbUser, dbPassword, dbName, dbHost, dbPort, dbSSLMode)

	// Connect to PostgreSQL
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	a.DB = db

	// Initialize router
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/signup", a.handleSignup).Methods("POST")
	a.Router.HandleFunc("/login", a.handleLogin).Methods("POST")
	a.Router.HandleFunc("/user", a.getUser).Methods("GET")
	a.Router.HandleFunc("/servers", a.handleCreateServer).Methods("POST")
	a.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")
	a.Router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(a.DB, a.Hub, w, r)
	}).Methods("GET")

	a.Hub = websocket.NewHub()
	go a.Hub.Run()

	return nil
}

func (a *App) handleSignup(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Insert user into database
	_, err = a.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		user.Username, user.Email, hashedPassword,
	)
	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	var dbUser User
	err = a.DB.Get(&dbUser, "SELECT id FROM users WHERE email=$1", user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": dbUser.ID,
	})
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Find user by email
	var dbUser User
	err := a.DB.Get(&dbUser, "SELECT id, username, email, password FROM users WHERE email=$1", user.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": dbUser.ID,
	})
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")

	userID, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var user User
	err = a.DB.Get(&user, "SELECT id, username, email FROM users WHERE id=$1", userID)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (a *App) handleCreateServer(w http.ResponseWriter, r *http.Request) {
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

	// Insert user into database
	var server_id int64
	err = a.DB.QueryRow(
		"INSERT INTO servers (name, owner_id) VALUES ($1, $2) RETURNING id",
		request.Name, userID,
	).Scan(&server_id)
	if err != nil {
		http.Error(w, "Could not create server", http.StatusInternalServerError)
		return
	}

	//Create default channel for the server
	_, err = a.DB.Exec(
		"INSERT INTO channels (server_id, name, type) VALUES ($1, $2, $3)",
		server_id, "text", "text",
	)
	if err != nil {
		http.Error(w, "Could not create default channel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "server created with default channel", "server_id": fmt.Sprintf("%d", server_id)})
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
