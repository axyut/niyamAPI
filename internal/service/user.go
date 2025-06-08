package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5" // Required for JWT operations
	"golang.org/x/crypto/bcrypt"   // For password hashing

	"github.com/danielgtaylor/huma/v2" // For Huma-specific error types

	"github.com/axyut/niyamAPI/internal/repository" // Adjust import path to your module
	"github.com/axyut/niyamAPI/internal/types"      // Adjust import path to your module
	"go.mongodb.org/mongo-driver/bson/primitive"    // For MongoDB ObjectID
)

// UserService defines the interface for user-related business logic operations,
// including user creation, retrieval, and authentication.
type UserService interface {
	// CreateUser handles new user registration. It hashes the password,
	// saves the user to the database, and then generates an authentication token.
	// Returns `*types.AuthOutput` which includes the token and user's public info.
	CreateUser(ctx context.Context, email, password string) (*types.AuthOutput, error)

	// GetUserByID retrieves a user by their unique ID.
	// It takes a string ID and returns a `*types.UserOutput`.
	GetUserByID(ctx context.Context, id string) (*types.UserOutput, error)

	// AuthenticateUser handles user login. It verifies the provided credentials,
	// and if valid, issues a new authentication token.
	// Returns `*types.AuthOutput` which includes the token and user's public info.
	AuthenticateUser(ctx context.Context, email, password string) (*types.AuthOutput, error)

	// GenerateToken creates a JSON Web Token (JWT) for a given user.
	// This is a utility method used internally by CreateUser and AuthenticateUser.
	GenerateToken(user *types.User) (string, error)

	// Add other user-related business methods here as your application grows,
	// e.g., UpdateUser, DeleteUser, ChangePassword, ResetPassword.
}

// userService is the concrete implementation of the UserService interface.
// It holds a reference to a `UserRepository`, which handles data persistence,
// and the JWT secret key for signing tokens.
type userService struct {
	userRepo  repository.UserRepository
	jwtSecret string // JWT secret from application configuration
}

// NewUserService creates and returns a new instance of UserService.
// It accepts a `UserRepository` and the JWT secret as dependencies.
func NewUserService(userRepo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// GenerateToken creates a JWT for the given user.
// This function constructs the JWT claims, signs the token using HMAC SHA256,
// and returns the signed token string.
func (s *userService) GenerateToken(user *types.User) (string, error) {
	// Define JWT claims. These include standard registered claims (like expiration)
	// and custom claims specific to your application (UserID, Email, Role).
	claims := types.AuthClaims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token valid for 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // Time when the token was issued
			NotBefore: jwt.NewNumericDate(time.Now()),                     // Token is not valid before this time
			Issuer:    "niyam-api",                                        // Identifier for the issuer of the token
			Subject:   user.ID.Hex(),                                      // Subject of the token (typically user ID)
			Audience:  jwt.ClaimStrings{"users"},                          // Audience for which the token is intended
		},
	}

	// Create a new token object with the defined claims and the signing method.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key. The secret key is converted to a byte slice.
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		log.Printf("ERROR: Failed to sign token for user %s: %v", user.Email, err)
		return "", fmt.Errorf("failed to sign token")
	}

	return tokenString, nil
}

// CreateUser handles the business logic for creating a new user.
// It first checks for existing users, hashes the password, persists the user,
// and then generates an authentication token for the new user.
func (s *userService) CreateUser(ctx context.Context, email, password string) (*types.AuthOutput, error) {
	// 1. Check if a user with the provided email already exists.
	// This prevents duplicate user registrations.
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		// If no error, a user with this email was found, indicating a conflict.
		log.Printf("INFO: Attempted to create user with existing email: %s", email)
		// FIXED: Pass `nil` or an actual `error` as the second argument.
		return nil, huma.Error409Conflict("user with this email already exists", nil)
	} else if err.Error() != fmt.Sprintf("user with email '%s' not found", email) {
		log.Printf("ERROR: Unexpected error when checking for existing user by email %s: %v", email, err)
		return nil, fmt.Errorf("failed to check existing user")
	}

	// 2. Hash the plain-text password using bcrypt.
	// bcrypt is a strong, adaptive hashing algorithm suitable for passwords.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash password for email %s: %v", email, err)
		return nil, fmt.Errorf("failed to hash password")
	}

	// 3. Prepare the new User model with all necessary fields.
	now := time.Now()
	user := &types.User{
		Email:     email,
		Password:  string(hashedPassword), // Store the hashed password
		Role:      "user",                 // Assign a default role
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 4. Call the UserRepository to persist the new user to the database.
	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		// Log and return a generic error if user creation fails at the repository level.
		log.Printf("ERROR: Service failed to create user %s in repository: %v", email, err)
		return nil, fmt.Errorf("failed to create user")
	}

	// 5. Generate a JWT for the newly created user.
	// This token can be used by the client for subsequent authenticated requests.
	token, err := s.GenerateToken(createdUser)
	if err != nil {
		log.Printf("ERROR: Failed to generate token for new user %s (ID: %s): %v", createdUser.Email, createdUser.ID.Hex(), err)
		return nil, fmt.Errorf("failed to generate authentication token")
	}

	// 6. Convert the internal User model to the public UserOutput type.
	// This transformation hides sensitive data like the password hash.
	userOutput := types.UserOutput{}
	userOutput.Body.ID = createdUser.ID.Hex() // Convert MongoDB ObjectID to string
	userOutput.Body.Email = createdUser.Email
	userOutput.Body.Role = createdUser.Role
	userOutput.Body.CreatedAt = createdUser.CreatedAt
	userOutput.Body.UpdatedAt = createdUser.UpdatedAt

	// 7. Return the authentication output, containing the token and user's public info.
	return &types.AuthOutput{Body: types.AuthOutputBody{
		Token: token,
		User:  userOutput,
	}}, nil
}

// GetUserByID retrieves a user by their ID.
// It parses the string ID, calls the repository, and transforms the result.
func (s *userService) GetUserByID(ctx context.Context, id string) (*types.UserOutput, error) {
	// 1. Validate and convert the string ID from the API request to a MongoDB ObjectID.
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Return a 400 Bad Request error if the ID format is invalid.
		// FIXED: Pass `nil` or an actual `error` as the second argument.
		return nil, huma.Error400BadRequest("invalid user ID format", nil)
	}

	// 2. Call the UserRepository to fetch the user by ObjectID.
	user, err := s.userRepo.GetUserByID(ctx, objID)
	if err != nil {
		// Handle specific "not found" error.
		if err.Error() == "user not found" {
			// FIXED: Pass `nil` or an actual `error` as the second argument.
			return nil, huma.Error404NotFound("user not found", nil)
		}
		// Log and return a generic internal server error for other unexpected issues.
		log.Printf("ERROR: Service failed to get user by ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve user")
	}

	// 3. Convert the internal User model to the public UserOutput type.
	userOutput := &types.UserOutput{}
	userOutput.Body.ID = user.ID.Hex()
	userOutput.Body.Email = user.Email
	userOutput.Body.Role = user.Role
	userOutput.Body.CreatedAt = user.CreatedAt
	userOutput.Body.UpdatedAt = user.UpdatedAt

	return userOutput, nil
}

// AuthenticateUser handles user login.
// It retrieves the user by email, compares the provided password with the stored hash,
// and if valid, generates and returns a new JWT.
func (s *userService) AuthenticateUser(ctx context.Context, email, password string) (*types.AuthOutput, error) {
	// 1. Retrieve the user by email from the repository.
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// For security, it's best practice to return a generic "authentication failed"
		// message whether the email was not found or the password was incorrect.
		log.Printf("INFO: Authentication attempt for email %s failed (user lookup error: %v)", email, err)
		// FIXED: Pass `nil` or an actual `error` as the second argument.
		return nil, huma.Error401Unauthorized("authentication failed", nil)
	}

	// 2. Compare the provided plain-text password with the stored hashed password.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// If passwords don't match, authentication fails.
		log.Printf("INFO: Authentication attempt for email %s failed (password mismatch)", email)
		// FIXED: Pass `nil` or an actual `error` as the second argument.
		return nil, huma.Error401Unauthorized("authentication failed", nil)
	}

	// 3. If credentials are valid, generate a JWT for the authenticated user.
	token, err := s.GenerateToken(user)
	if err != nil {
		log.Printf("ERROR: Failed to generate token for authenticated user %s (ID: %s): %v", user.Email, user.ID.Hex(), err)
		return nil, fmt.Errorf("failed to generate authentication token")
	}

	// 4. Prepare the AuthOutput, which includes the generated token and public user information.
	userOutput := types.UserOutput{}
	userOutput.Body.ID = user.ID.Hex()
	userOutput.Body.Email = user.Email
	userOutput.Body.Role = user.Role
	userOutput.Body.CreatedAt = user.CreatedAt
	userOutput.Body.UpdatedAt = user.UpdatedAt

	// 5. Return the authentication output, using the new named struct.
	return &types.AuthOutput{Body: types.AuthOutputBody{
		Token: token,
		User:  userOutput,
	}}, nil
}
