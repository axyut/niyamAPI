package handler

import (
	"context"
	"log"

	"github.com/danielgtaylor/huma/v2"

	"github.com/axyut/niyamAPI/internal/types" // Adjust import path to your module
)

// RegisterUserHandlers registers API endpoints related to user management.
// It's a method on the Handlers struct, giving it access to AppConfig, Services, and DBClient.
func (h *Handlers) RegisterUserHandlers(api huma.API) {
	// POST /users: Endpoint to create a new user (signup).
	// Expects CreateUserInput and now returns AuthOutput on success (which includes the JWT and user info).
	huma.Post(api, "/users", func(ctx context.Context, input *types.CreateUserInput) (*types.AuthOutput, error) { // <--- Changed return type to *types.AuthOutput
		log.Printf("INFO: Received request to create user: %s", input.Body.Email)

		// Call the UserService to handle the business logic of user creation.
		// The service will now return the AuthOutput, which includes the JWT token.
		authOutput, err := h.Services.UserService.CreateUser(ctx, input.Body.Email, input.Body.Password)
		if err != nil {
			// Huma automatically converts errors into Problem JSON based on `huma.Error` types.
			// Log the error and return it; Huma will handle the HTTP status code.
			log.Printf("ERROR: Failed to create user %s: %v", input.Body.Email, err)
			return nil, err // Return the error directly; Huma handles the Problem JSON conversion.
		}

		log.Printf("INFO: User created and authenticated successfully: %s (ID: %s)", authOutput.Body.User.Body.Email, authOutput.Body.User.Body.ID)
		return authOutput, nil // Return the AuthOutput directly
	})

	// GET /users/{id}: Endpoint to retrieve a user by their ID.
	// Expects GetUserByIDInput (from path parameter) and returns UserOutput.
	huma.Get(api, "/users/{id}", func(ctx context.Context, input *types.GetUserByIDInput) (*types.UserOutput, error) {
		log.Printf("INFO: Received request to get user by ID: %s", input.ID)

		// Call the UserService to retrieve the user.
		userOutput, err := h.Services.UserService.GetUserByID(ctx, input.ID)
		if err != nil {
			// Huma will automatically convert huma.Error400BadRequest and huma.Error404NotFound
			// errors returned by the service layer.
			log.Printf("ERROR: Failed to get user %s: %v", input.ID, err)
			return nil, err // Return the error directly; Huma handles the Problem JSON conversion.
		}

		log.Printf("INFO: User found: %s (ID: %s)", userOutput.Body.Email, userOutput.Body.ID)
		return userOutput, nil // Return the found user's public data
	})
}
