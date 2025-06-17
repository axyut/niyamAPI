package handler

import (
	"github.com/axyut/niyamAPI/internal/config"  // Adjust this to your actual module path
	"github.com/axyut/niyamAPI/internal/db"      // Adjust this to your actual module path
	"github.com/axyut/niyamAPI/internal/service" // Adjust this to your actual module path

	// Adjust this to your actual module path
	"github.com/danielgtaylor/huma/v2"
)

// Handlers holds dependencies for all API handlers.
type Handlers struct {
	AppConfig *config.AppConfig
	Services  *service.Services
	DBClient  *db.Client
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(cfg *config.AppConfig, svc *service.Services, dbClient *db.Client) *Handlers {
	return &Handlers{
		AppConfig: cfg,
		Services:  svc,
		DBClient:  dbClient,
	}
}

// RegisterHandlers registers all API endpoints with the Huma API instance.
func (h *Handlers) RegisterHandlers(api huma.API) {
	// Register core handlers (home, health)
	h.RegisterHomeHandlers(api)

	// Register user-related handlers (signup, get user)
	h.RegisterUserHandlers(api)

	// Register authentication handlers (login)
	h.RegisterAuthHandlers(api)

	// --- THIS IS THE CRUCIAL LINE FOR /scan ROUTE ---
	h.RegisterScanHandlers(api) // Make absolutely sure this line is present and uncommented!

}
