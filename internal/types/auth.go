package types

import (
	"github.com/golang-jwt/jwt/v5" // Required for JWT claims structure
	// Required for MongoDB's ObjectID (if user ID is ObjectID)
)

// LoginInput is the input structure for the /auth/login endpoint.
// It defines the fields expected in the request body for a login attempt.
type LoginInput struct {
	Body struct {
		Email    string `json:"email" huma:"minLength:5,maxLength:100,example:user@example.com" doc:"User's email address"`
		Password string `json:"password" huma:"minLength:8,maxLength:50,example:SecurePass123!" doc:"User's password"`
	}
}

// AuthClaims defines the custom claims that will be embedded within your JWT.
// It includes standard JWT claims (via embedding `jwt.RegisteredClaims`)
// and custom application-specific claims like UserID, Email, and Role.
type AuthClaims struct {
	UserID               string `json:"userId"` // The unique identifier of the user
	Email                string `json:"email"`  // The user's email address
	Role                 string `json:"role"`   // The user's assigned role (e.g., "user", "admin")
	jwt.RegisteredClaims        // Standard JWT claims (e.g., ExpiresAt, IssuedAt)
}

// AuthOutputBody defines the structure for the JSON response body
// returned upon successful login or signup. This is a named struct
// to resolve the anonymous struct mismatch error.
type AuthOutputBody struct {
	Token string     `json:"token" huma:"example:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." doc:"Authentication token"`
	User  UserOutput `json:"user" doc:"Basic public user information"`
}

// AuthOutput is the output structure for a successful login or signup operation.
// It uses the named AuthOutputBody struct for its body.
type AuthOutput struct {
	Body AuthOutputBody // <--- Changed from anonymous struct to named AuthOutputBody
}
