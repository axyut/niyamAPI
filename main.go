package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	// "github.com/shirou/gopsutil/cpu" // Commented out as per previous request to use simulated load

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

// Global variable to store the server's start time for uptime calculation.
var startTime = time.Now()

// MetadataOutput represents the response structure for the root endpoint.
type MetadataOutput struct {
	Body struct {
		Service       string          `json:"service" example:"My API"`
		Version       string          `json:"version" example:"v1"`
		Description   string          `json:"description" example:"API description"`
		Status        string          `json:"status" example:"operational"`
		Uptime        string          `json:"uptime" example:"8d 19h 16m"`
		Health        HealthStatus    `json:"health"`
		Documentation string          `json:"documentation" example:"/docs"`
		Links         MetadataLinks   `json:"links"`
		Contact       MetadataContact `json:"contact"`
		Environment   string          `json:"environment" example:"development"`
	}
}

// HealthStatus holds health check details.
type HealthStatus struct {
	Database string  `json:"database" example:"ok"`
	Server   string  `json:"server" example:"ok"`
	Load     float64 `json:"load" example:"11.35"` // Simulated load value
}

// MetadataLinks holds related links.
type MetadataLinks struct {
	Self          string `json:"self" example:"/"`
	PrivacyPolicy string `json:"privacyPolicy" example:"/api/terms_and_condition"`
	// Add more links as needed
}

// MetadataContact holds contact information.
type MetadataContact struct {
	Name  string `json:"name" example:"API Support"`
	Email string `json:"email" example:"mail@achyutkoirala.com.np"`
	URL   string `json:"url" example:"/contact"`
}

// HealthCheckOutput represents the health check response (already defined, kept for consistency).
type HealthCheckOutput struct {
	Body struct {
		Status string `json:"status" example:"healthy" doc:"API health status"`
	}
}

func main() {
	// --- 1. Load environment variables from .env file ---
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or error loading .env: %v. Proceeding without it.", err)
	} else {
		log.Println(".env file loaded successfully.")
	}

	// --- 2. Configuration: Get port from environment variable with a default ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified in .env or actual environment
	}
	listenAddr := ":" + port

	// --- 3. Create a new router & API ---
	router := chi.NewMux()

	// Add chi middleware for request logging and panic recovery
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // Recovers from panics, prevents server crash

	// Store the Huma config in a variable to access its fields
	apiConfig := huma.DefaultConfig("Niyam API", "1.0.0")
	apiConfig.Info.Description = "API/Backend service for the Niyam application."
	api := humachi.New(router, apiConfig)

	// --- 4. Register API Handlers ---

	// Root endpoint "/"
	huma.Get(api, "/", func(ctx context.Context, input *struct{}) (*MetadataOutput, error) {
		// Calculate uptime
		uptimeDuration := time.Since(startTime)
		uptime := fmt.Sprintf("%dd %dh %dm",
			int(uptimeDuration.Hours()/24),
			int(uptimeDuration.Hours())%24,
			int(uptimeDuration.Minutes())%60,
		)

		// Simulated load
		simulatedLoad := rand.Float64() * 20.0 // Random float between 0.0 and 20.0

		resp := &MetadataOutput{}
		resp.Body.Service = apiConfig.Info.Title
		resp.Body.Version = apiConfig.Info.Version
		resp.Body.Description = apiConfig.Info.Description
		resp.Body.Status = "operational"
		resp.Body.Uptime = uptime
		resp.Body.Health = HealthStatus{
			Database: "ok",          // Placeholder: In a real app, check DB connection
			Server:   "ok",          // Placeholder: Always "ok" if server is responding
			Load:     simulatedLoad, // Using the simulated load value
		}
		resp.Body.Documentation = "/docs"

		resp.Body.Links = MetadataLinks{
			Self:          "/",
			PrivacyPolicy: "/api/terms_and_condition",
		}

		resp.Body.Contact = MetadataContact{
			Name:  "API Support",
			Email: "mail@achyutkoirala.com.np",
			URL:   "/contact",
		}

		// Get environment from an ENV var, default to "development"
		appEnv := os.Getenv("APP_ENV")
		if appEnv == "" {
			appEnv = "development"
		}
		resp.Body.Environment = appEnv

		return resp, nil
	})

	// Health Check Endpoint
	huma.Get(api, "/healthz", func(ctx context.Context, input *struct{}) (*HealthCheckOutput, error) {
		// log.Println("Health check requested.") // Logger middleware will cover this.
		resp := &HealthCheckOutput{}
		resp.Body.Status = "healthy"
		return resp, nil
	})

	// --- 5. Dynamically Get the Actual Host and Port (for logging) ---
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Server failed to listen on %s: %v", listenAddr, err)
	}
	defer listener.Close()

	// Extract just the port number from the listener's address
	_, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		log.Fatalf("Failed to parse listener address: %v", err)
	}

	// --- Construct the Base URL for logging without explicit environment checks ---
	var publicBaseURL string
	// Attempt to get a public URL from an environment variable first
	if publicURL := os.Getenv("API_PUBLIC_URL"); publicURL != "" {
		publicBaseURL = publicURL
	} else {
		// Default to localhost for local development if no public URL is provided
		publicBaseURL = fmt.Sprintf("http://localhost:%s", portStr)
	}

	// Use the constructed publicBaseURL in the log message
	log.Printf("Server listening on %s (Access docs at %s/docs)\n", listener.Addr().String(), publicBaseURL)

	// --- 6. Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         listener.Addr().String(),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting HTTP server on %s...", srv.Addr)
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
		log.Println("HTTP server stopped.")
	}()

	sig := <-quit
	log.Printf("Received signal '%s'. Shutting down server...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully shut down.")
}
