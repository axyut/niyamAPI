package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref" // For pinging the DB

	"github.com/axyut/niyamAPI/internal/config" // Adjust import path to your module
)

// Client represents your database client.
type Client struct {
	Mongo *mongo.Client
}

// NewDBClient initializes and returns a new MongoDB client.
func NewDBClient(cfg *config.AppConfig) (*Client, error) {
	log.Println("INFO: Initializing MongoDB client...")

	// Use context with a timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB using the URI from config
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the primary to verify connection and credentials
	// This ensures the database is reachable and credentials are correct.
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		// If ping fails, attempt to disconnect cleanly.
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		if dcErr := client.Disconnect(disconnectCtx); dcErr != nil {
			log.Printf("WARNING: Failed to disconnect MongoDB client after ping failure: %v", dcErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("INFO: MongoDB client connected and pinged successfully.")
	return &Client{Mongo: client}, nil
}

// Close closes the MongoDB connection.
func (c *Client) Close() error {
	log.Println("INFO: Closing MongoDB client...")
	if c.Mongo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.Mongo.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect MongoDB client: %w", err)
		}
	}
	log.Println("INFO: MongoDB client disconnected.")
	return nil
}

// IsConnected checks if the MongoDB client is currently connected and reachable.
// This is used for the health check endpoint.
func (c *Client) IsConnected() bool {
	if c.Mongo == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := c.Mongo.Ping(ctx, readpref.Primary())
	return err == nil
}
