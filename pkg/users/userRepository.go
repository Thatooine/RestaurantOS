package users

import "context"

// UserRepository defines the persistence port for users. It combines creation
// and read operations over the users store. This is an internal port and is
// NOT exposed over REST — see UserService for the exposed read operations.
type UserRepository interface {
	CreateUser(ctx context.Context, request CreateUserRequest) (*CreateUserResponse, error)
	GetUser(ctx context.Context, request GetUserRequest) (*GetUserResponse, error)
	GetUserByID(ctx context.Context, request GetUserByIDRequest) (*GetUserByIDResponse, error)
	ListUsers(ctx context.Context, request ListUsersRequest) (*ListUsersResponse, error)
	SearchUsers(ctx context.Context, request SearchUsersRequest) (*SearchUsersResponse, error)
	UpdateUserRoles(ctx context.Context, request UpdateUserRolesRequest) (*UpdateUserRolesResponse, error)
}

// CreateUser

type CreateUserRequest struct {
	Name         string
	Email        string
	PasswordHash string
}

type CreateUserResponse struct {
	User User
}

// GetUser

type GetUserRequest struct {
	Email string
}

type GetUserResponse struct {
	User User
}

// GetUserByID

type GetUserByIDRequest struct {
	ID string
}

type GetUserByIDResponse struct {
	User User
}

// UpdateUserRoles

type UpdateUserRolesRequest struct {
	UserID string
	Roles  []Role
}

type UpdateUserRolesResponse struct {
	User User
}

// ListUsers

type ListUsersRequest struct {
	Offset int
	Limit  int
}

type ListUsersResponse struct {
	Users []User
	Total int64
}

// SearchUsers

type SearchUsersRequest struct {
	Query  string
	Offset int
	Limit  int
}

type SearchUsersResponse struct {
	Users []User
	Total int64
}
