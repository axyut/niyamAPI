package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive" // For MongoDB's ObjectID
)

// User represents the User model in your database.
// `bson` tags are for MongoDB, `json` tags for API responses, `huma` tags for OpenAPI.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" huma:"example:654a93c7e0f2f3f4c5d6e7f8"`
	Email     string             `bson:"email" json:"email" huma:"example:test@example.com,minLength:5,maxLength:100"`
	Password  string             `bson:"password" json:"-" huma:"readOnly:true,example:hashedpassword" doc:"Hashed password, not sent to client"` // `json:"-"` prevents marshaling, `huma:"readOnly:true"` for docs
	Role      string             `bson:"role" json:"role" huma:"example:user,enum:user|admin" doc:"User role (e.g., user, admin)"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt" huma:"example:2024-01-01T12:00:00Z"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt" huma:"example:2024-01-01T12:00:00Z"`
}

// CreateUserInput is the input structure for creating a new user.
type CreateUserInput struct {
	Body struct {
		Email    string `json:"email" huma:"minLength:5,maxLength:100,example:newuser@example.com" doc:"User's email address"`
		Password string `json:"password" huma:"minLength:8,maxLength:50,example:SecurePass123!" doc:"User's password (min 8 chars)"`
	}
}

// UserOutput is the output structure for returning a user.
// It omits sensitive information like the password hash.
type UserOutput struct {
	Body struct {
		ID        string    `json:"id" huma:"example:654a93c7e0f2f3f4c5d6e7f8"`
		Email     string    `json:"email" huma:"example:test@example.com"`
		Role      string    `json:"role" huma:"example:user"`
		CreatedAt time.Time `json:"createdAt" huma:"example:2024-01-01T12:00:00Z"`
		UpdatedAt time.Time `json:"updatedAt" huma:"example:2024-01-01T12:00:00Z"`
	}
}

// GetUserByIDInput is the input structure for getting a user by ID.
type GetUserByIDInput struct {
	ID string `path:"id" huma:"example:654a93c7e0f2f3f4c5d6e7f8" doc:"User ID"`
}
