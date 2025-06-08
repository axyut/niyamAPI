package handler

import (
	"context"
	"math/rand" // For simulated load

	"github.com/danielgtaylor/huma/v2"

	"github.com/axyut/niyamAPI/internal/types" // Adjust import path to your module
	"github.com/axyut/niyamAPI/internal/utils" // Adjust import path to your module
)

// RegisterHomeHandlers registers API endpoints for the root and health checks.
// This method is part of the Handlers struct, giving it access to shared dependencies
// like AppConfig and DBClient.
func (h *Handlers) RegisterHomeHandlers(api huma.API) {
	// Root endpoint "/" provides comprehensive metadata about the API service.
	// It's a GET request that returns a MetadataOutput JSON response.
	huma.Get(api, "/", func(ctx context.Context, input *struct{}) (*types.MetadataOutput, error) {
		// Calculate the application's uptime using the utility function.
		uptime := utils.GetUptime()

		// Simulate the current server load with a random floating-point number
		// between 0.0 and 20.0. This is a placeholder for actual system metrics.
		simulatedLoad := rand.Float64() * 20.0

		// Check the actual database connection status using the DBClient.
		// If DBClient.IsConnected() returns true, the status is "ok"; otherwise, "down".
		dbStatus := "down"
		if h.DBClient.IsConnected() {
			dbStatus = "ok"
		}

		// Create and populate the MetadataOutput response struct.
		resp := &types.MetadataOutput{}
		resp.Body.Service = h.AppConfig.PublicURL              // Uses the public URL from config.
		resp.Body.Version = api.OpenAPI().Info.Version         // API version from Huma config.
		resp.Body.Description = api.OpenAPI().Info.Description // API description from Huma config.
		resp.Body.Status = "operational"                       // General service status.
		resp.Body.Uptime = uptime                              // Formatted uptime string.
		resp.Body.Health = types.HealthStatus{                 // Detailed health status.
			Database: dbStatus, // Reflects actual MongoDB connection status.
			Server:   "ok",     // The server is "ok" if it successfully responds to this request.
			Load:     simulatedLoad,
		}
		resp.Body.Documentation = "/docs" // Relative path to Huma's generated OpenAPI documentation.

		// Populate links related to the API, such as self-reference and policies.
		resp.Body.Links = types.MetadataLinks{
			Self:          "/",
			PrivacyPolicy: "/api/terms_and_condition", // Example: A link to privacy policy.
		}

		// Populate contact information for API support.
		resp.Body.Contact = types.MetadataContact{
			Name:  "API Support",
			Email: "mail@achyutkoirala.com.np",
			URL:   "/contact", // Example: A link to contact page.
		}

		// Get the application environment from the loaded configuration.
		resp.Body.Environment = h.AppConfig.Environment

		return resp, nil
	})

	// Health Check Endpoint "/healthz" provides a simple, quick status check.
	// It's a GET request that returns a HealthCheckOutput JSON response.
	huma.Get(api, "/healthz", func(ctx context.Context, input *struct{}) (*types.HealthCheckOutput, error) {
		resp := &types.HealthCheckOutput{}
		resp.Body.Status = "healthy" // Simple "healthy" status.
		return resp, nil
	})
}
