package users

import (
	"context"

	pkgUsers "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
)

var _ pkgUsers.UserRepository = &UserRepositoryMock{}

// UserRepositoryMock is a hand-written mock of users.UserRepository.
type UserRepositoryMock struct {
	CreateUserFn      func(ctx context.Context, request pkgUsers.CreateUserRequest) (*pkgUsers.CreateUserResponse, error)
	GetUserFn         func(ctx context.Context, request pkgUsers.GetUserRequest) (*pkgUsers.GetUserResponse, error)
	GetUserByIDFn     func(ctx context.Context, request pkgUsers.GetUserByIDRequest) (*pkgUsers.GetUserByIDResponse, error)
	ListUsersFn       func(ctx context.Context, request pkgUsers.ListUsersRequest) (*pkgUsers.ListUsersResponse, error)
	SearchUsersFn     func(ctx context.Context, request pkgUsers.SearchUsersRequest) (*pkgUsers.SearchUsersResponse, error)
	UpdateUserRolesFn func(ctx context.Context, request pkgUsers.UpdateUserRolesRequest) (*pkgUsers.UpdateUserRolesResponse, error)
}

func (m *UserRepositoryMock) CreateUser(ctx context.Context, request pkgUsers.CreateUserRequest) (*pkgUsers.CreateUserResponse, error) {
	return m.CreateUserFn(ctx, request)
}

func (m *UserRepositoryMock) GetUser(ctx context.Context, request pkgUsers.GetUserRequest) (*pkgUsers.GetUserResponse, error) {
	return m.GetUserFn(ctx, request)
}

func (m *UserRepositoryMock) GetUserByID(ctx context.Context, request pkgUsers.GetUserByIDRequest) (*pkgUsers.GetUserByIDResponse, error) {
	return m.GetUserByIDFn(ctx, request)
}

func (m *UserRepositoryMock) ListUsers(ctx context.Context, request pkgUsers.ListUsersRequest) (*pkgUsers.ListUsersResponse, error) {
	return m.ListUsersFn(ctx, request)
}

func (m *UserRepositoryMock) SearchUsers(ctx context.Context, request pkgUsers.SearchUsersRequest) (*pkgUsers.SearchUsersResponse, error) {
	return m.SearchUsersFn(ctx, request)
}

func (m *UserRepositoryMock) UpdateUserRoles(ctx context.Context, request pkgUsers.UpdateUserRolesRequest) (*pkgUsers.UpdateUserRolesResponse, error) {
	return m.UpdateUserRolesFn(ctx, request)
}
