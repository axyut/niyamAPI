package handler

import (
	"github.com/danielgtaylor/huma/v2"

	"github.com/axyut/niyamAPI/internal/config"  // Adjust import path
	"github.com/axyut/niyamAPI/internal/service" // Adjust import path

	// Adjust import path
	"github.com/axyut/niyamAPI/internal/db" // Adjust import path to your module
)

// Handlers holds dependencies for all API handlers.
// This struct will be initialized once in main.go and passed to all handler
// registration functions. It centralizes access to configuration, business
// services, and the database client.
type Handlers struct {
	AppConfig *config.AppConfig // Application-wide configuration
	Services  *service.Services // Business logic services layer
	DBClient  *db.Client        // MongoDB client for database operations/health checks
}

// NewHandlers creates a new Handlers instance.
// It takes instances of your configuration, services, and database client
// as dependencies, setting them up for use across all API handlers.
func NewHandlers(cfg *config.AppConfig, svc *service.Services, dbClient *db.Client) *Handlers {
	return &Handlers{
		AppConfig: cfg,
		Services:  svc,
		DBClient:  dbClient, // Pass the initialized MongoDB client here
	}
}

// RegisterHandlers registers all API endpoints with the Huma API instance.
// This method orchestrates the registration of handlers from different logical
// groups (e.g., home, user, goods).
func (h *Handlers) RegisterHandlers(api huma.API) {
	// Register handlers for the root ("/") and health ("healthz") endpoints.
	h.RegisterHomeHandlers(api)
	h.RegisterUserHandlers(api)
	// --- Placeholder for registering other domain-specific handlers ---
	// As you implement more features (e.g., for users, goods, transactions),
	// you would call their respective registration methods here.
	// Example:
	// h.RegisterUserHandlers(api)
	// h.RegisterGoodsHandlers(api)
	// h.RegisterTransactionHandlers(api)
	// h.RegisterProductionHandlers(api)
	// h.RegisterReportsHandlers(api)
	// h.RegisterAudienceHandlers(api)
}
