package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig holds all application-wide configurations.
type AppConfig struct {
	Port        int
	PublicURL   string
	Environment string
	MongoURI    string
	JWTSecret   string // <--- NEW: Secret key for JWT signing
	// Add other configurations like API keys etc.
}

// LoadConfig loads application configuration from environment variables and .env file.
func LoadConfig() (*AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or error loading .env: %v. Proceeding with system environment variables.", err)
	} else {
		log.Println(".env file loaded successfully.")
	}

	cfg := &AppConfig{}

	// Port configuration
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	cfg.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT environment variable: %w", err)
	}

	// Public URL for the API
	cfg.PublicURL = os.Getenv("API_PUBLIC_URL")
	if cfg.PublicURL == "" {
		cfg.PublicURL = fmt.Sprintf("http://localhost:%d", cfg.Port)
	}

	// Application environment
	cfg.Environment = os.Getenv("APP_ENV")
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	// MongoDB Configuration
	cfg.MongoURI = os.Getenv("MONGO_URI")
	if cfg.MongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI environment variable is not set")
	}

	// --- NEW: JWT Secret Configuration ---
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	return cfg, nil
}
