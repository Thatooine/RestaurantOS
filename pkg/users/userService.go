package users

import "context"

// UserService defines the exposed application service for viewing, listing, and
// searching users. It is the REST-facing port (see UserServiceRESTAdaptor) and
// delegates to the internal UserRepository. The request/response DTOs are shared
// with UserRepository (declared in userRepository.go).
type UserService interface {
	GetUser(ctx context.Context, request GetUserRequest) (*GetUserResponse, error)
	ListUsers(ctx context.Context, request ListUsersRequest) (*ListUsersResponse, error)
	SearchUsers(ctx context.Context, request SearchUsersRequest) (*SearchUsersResponse, error)
}
