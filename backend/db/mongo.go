package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() (*mongo.Client, error) {
	mongoHost := os.Getenv("MONGO_HOST")
	mongoDB := os.Getenv("MONGO_DB")
	mongoUser := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoAuthSource := os.Getenv("MONGO_AUTH_SOURCE")

	if mongoHost == "" || mongoDB == "" || mongoUser == "" || mongoPassword == "" || mongoAuthSource == "" {
		return nil, fmt.Errorf("MongoDB environment variables are not set properly")
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
		mongoUser, mongoPassword, mongoHost, mongoDB, mongoAuthSource)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("✅ Connected to MongoDB")
	return client, nil
}
