package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ConnectPostgres connects to PostgreSQL with retry logic.
func ConnectPostgres() (*sqlx.DB, error) {
	dbUser := os.Getenv("SQLDB_USER")
	dbPassword := os.Getenv("SQLDB_PASSWORD")
	dbName := os.Getenv("SQLDB_NAME")
	dbHost := os.Getenv("SQLDB_HOST")
	dbPort := os.Getenv("SQLDB_PORT")
	dbSSLMode := os.Getenv("SQLDB_SSLMODE")

	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		dbUser, dbPassword, dbName, dbHost, dbPort, dbSSLMode,
	)

	var db *sqlx.DB
	var err error
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err == nil {
			log.Println("✅ Connected to PostgreSQL")
			return db, nil
		}
		log.Printf("PostgreSQL not ready, retrying in 3 seconds... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to PostgreSQL after %d retries: %w", maxRetries, err)
}
