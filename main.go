package main

import (
	"context"
	"fmt"
	"log"
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

	// IMPORTANT: Adjust these import paths to match your Go module path
	// (e.g., github.com/your-username/your-repo-name/internal/...)
	"github.com/axyut/niyamAPI/internal/config"
	"github.com/axyut/niyamAPI/internal/db"
	"github.com/axyut/niyamAPI/internal/handler"
	"github.com/axyut/niyamAPI/internal/service"
	"github.com/axyut/niyamAPI/internal/utils"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor" // Huma format package for CBOR
)

func main() {
	// 1. Initialize Uptime Tracking
	// Call this as early as possible in the main function to accurately track
	// the application's uptime from its true starting point.
	utils.InitUptime()

	// 2. Load Application Configuration
	// This step reads environment variables (including those from .env file via godotenv)
	// and parses them into a structured AppConfig. Fatal exit if essential config is missing.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Failed to load application configuration: %v", err)
	}
	log.Printf("INFO: Application environment: %s", cfg.Environment)
	log.Printf("INFO: Configured API Public URL: %s", cfg.PublicURL)

	// 3. Initialize Database Client
	// Establish connection to MongoDB using the loaded configuration.
	// This is a critical dependency, so a fatal exit occurs on failure.
	dbClient, err := db.NewDBClient(cfg) // Pass configuration to DB client for connection URI
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize database client: %v", err)
	}
	// Defer closing the database connection. This ensures the connection is
	// gracefully closed when the main function exits (e.g., on shutdown signal).
	defer func() {
		if err := dbClient.Close(); err != nil {
			log.Printf("ERROR: Error closing database client: %v", err)
		}
	}()

	// 4. Initialize Services (Business Logic Layer)
	// Services encapsulate the core business logic of your application. They depend on
	// external resources like the database client.
	svc := service.NewServices(dbClient, cfg) // Pass dbClient to the service layer

	// 5. Create new Chi router
	// Chi is a lightweight, idiomatic router for building HTTP services in Go.
	router := chi.NewMux()

	// Apply Chi middleware for request logging and panic recovery.
	// `middleware.Logger` provides structured logging for each incoming HTTP request,
	// showing method, path, status, and duration.
	router.Use(middleware.Logger)
	// `middleware.Recoverer` catches any panics that occur within handlers,
	// logs the stack trace, and prevents the entire server from crashing,
	// returning a 500 Internal Server Error to the client.
	router.Use(middleware.Recoverer)

	// 6. Create Huma API instance
	// Huma builds on top of Chi (via humachi adapter) to provide powerful features
	// like automatic OpenAPI 3.0 documentation generation and request/response validation.
	apiConfig := huma.DefaultConfig("Niyam API", "1.0.0")
	apiConfig.Info.Description = "API/Backend service for the Niyam application."

	api := humachi.New(router, apiConfig)

	// 7. Register API Handlers
	// Instantiate the Handlers struct, passing it all its essential dependencies:
	// application configuration, business services, and the database client.
	// Then, call its method to register all defined API endpoints with the Huma API.
	hndlrs := handler.NewHandlers(cfg, svc, dbClient) // Pass config, services, and dbClient to handlers
	hndlrs.RegisterHandlers(api)

	// // 8. Start HTTP Server
	// // The server listens on the port specified in the configuration.
	listenAddr := fmt.Sprintf(":%d", cfg.Port)
	// listener, err := net.Listen("tcp", listenAddr)
	// if err != nil {
	// 	log.Fatalf("FATAL: Server failed to listen on %s: %v", listenAddr, err)
	// }
	// defer listener.Close() // Ensure the listener is closed when main function exits.

	// // Log the actual server listening address (e.g., ":7860" or "[::]:7860")
	// // and the publicly accessible documentation URL, which uses the PublicURL from config.
	// log.Printf("INFO: Server listening on %s (Access docs at %s/docs)\n", listener.Addr().String(), cfg.PublicURL)

	// Dynamically Get the Actual Host and Port (for logging)
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

	publicBaseURL := fmt.Sprintf("http://localhost:%s", portStr)
	appEnv := os.Getenv("APP_ENV")

	if publicURL := os.Getenv("API_PUBLIC_URL"); publicURL != "" && appEnv != "development" {
		publicBaseURL = publicURL
	}

	// Use the constructed publicBaseURL in the log message
	log.Printf("Server listening on %s (Access docs at %s/docs\n", listener.Addr().String(), publicBaseURL)

	// Configure the HTTP server with explicit timeouts for read, write, and idle operations.
	// This helps prevent resource exhaustion and improves server robustness.
	srv := &http.Server{
		Addr:         listener.Addr().String(), // Server address determined by listener
		Handler:      router,                   // The Chi router handles all incoming requests
		ReadTimeout:  5 * time.Second,          // Maximum duration for reading the entire request
		WriteTimeout: 10 * time.Second,         // Maximum duration before timing out writes of the response
		IdleTimeout:  120 * time.Second,        // Maximum amount of time to wait for the next request when keep-alives are enabled
	}

	// Start the HTTP server in a separate goroutine.
	// This allows the main goroutine to proceed to set up graceful shutdown.
	go func() {
		log.Printf("INFO: Starting HTTP server on %s...", srv.Addr)
		// srv.Serve() blocks until the server closes or an error occurs.
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			// Log a fatal error if the server fails to start or serve requests unexpectedly,
			// ensuring the application exits.
			log.Fatalf("FATAL: HTTP server failed: %v", err)
		}
		log.Println("INFO: HTTP server stopped.")
	}()

	// 9. Graceful Shutdown
	// Set up a channel to receive OS signals. This allows the application to respond
	// to termination requests (e.g., Ctrl+C, `docker stop`).
	quit := make(chan os.Signal, 1)
	// Notify the 'quit' channel upon receiving SIGINT (Ctrl+C) or SIGTERM (termination request).
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block the main goroutine until an OS signal is received on the 'quit' channel.
	sig := <-quit
	log.Printf("INFO: Received signal '%s'. Shutting down server gracefully...", sig)

	// Create a context with a timeout for the graceful shutdown process.
	// If the server doesn't shut down within this duration, it will be forcefully closed.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // Ensure the context's cancel function is called to release resources.

	// Attempt to gracefully shut down the HTTP server.
	// Existing connections are given time to complete before the server is stopped.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server shutdown failed: %v", err)
	}

	log.Println("INFO: Server gracefully shut down.")
}
