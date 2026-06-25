package users

import "github.com/google/uuid"

// NewID generates a UUID v4 string.
func NewID() string {
	return uuid.New().String()
}

type Role string

const (
	RoleAdmin           Role = "Admin"
	RoleCustomer        Role = "Customer"
	RoleRestaurantOwner Role = "RestaurantOwner"
)

// User represents a registered user of the application.
type User struct {
	// Unique identifier for this user
	ID string `json:"id" bson:"id"`
	// Full name of the user
	Name string `json:"name" bson:"name"`
	// Email address used for login
	Email string `json:"email" bson:"email"`
	// Roles assigned to the user
	Roles []Role `json:"roles" bson:"roles"`
	// Bcrypt hash of the user's password. Never serialized to JSON.
	PasswordHash string `json:"-" bson:"passwordHash"`
}
