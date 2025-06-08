package repository

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/axyut/niyamAPI/internal/types" // Adjust import path
)

// UserRepository defines the interface for user data operations.
// This abstraction allows swapping out database implementations easily.
type UserRepository interface {
	CreateUser(ctx context.Context, user *types.User) (*types.User, error)
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*types.User, error)
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	// Add other CRUD operations as needed:
	// UpdateUser(ctx context.Context, id primitive.ObjectID, updates bson.M) error
	// DeleteUser(ctx context.Context, id primitive.ObjectID) error
}

// mongoUserRepository implements UserRepository for MongoDB.
type mongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoDB user repository.
func NewMongoUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepository{
		collection: db.Collection("users"), // Assuming your users collection is named "users"
	}
}

// CreateUser inserts a new user into MongoDB.
func (r *mongoUserRepository) CreateUser(ctx context.Context, user *types.User) (*types.User, error) {
	// Ensure ID is generated if not already set (e.g., if you don't generate it in service layer)
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		// Handle duplicate key error specifically (e.g., duplicate email)
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("user with this email already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Verify the inserted ID matches the user's ID
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok || insertedID != user.ID {
		log.Printf("WARNING: Mismatch in inserted ID: expected %v, got %v", user.ID, insertedID)
	}

	log.Printf("INFO: User created with ID: %s", user.ID.Hex())
	return user, nil
}

// GetUserByID retrieves a user by their MongoDB ObjectID.
func (r *mongoUserRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*types.User, error) {
	var user types.User
	filter := bson.M{"_id": id} // Filter by MongoDB's _id field

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email address.
func (r *mongoUserRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	filter := bson.M{"email": email} // Filter by the email field

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}
