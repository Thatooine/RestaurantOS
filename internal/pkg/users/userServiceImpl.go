package users

import (
	"context"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
)

// UserServiceImpl implements users.UserService by delegating read operations to
// the internal users.UserRepository.
type UserServiceImpl struct {
	repository users.UserRepository
}

var _ users.UserService = &UserServiceImpl{}

func NewUserServiceImpl(repository users.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		repository: repository,
	}
}

func (s *UserServiceImpl) GetUser(ctx context.Context, request users.GetUserRequest) (*users.GetUserResponse, error) {
	return s.repository.GetUser(ctx, request)
}

func (s *UserServiceImpl) ListUsers(ctx context.Context, request users.ListUsersRequest) (*users.ListUsersResponse, error) {
	return s.repository.ListUsers(ctx, request)
}

func (s *UserServiceImpl) SearchUsers(ctx context.Context, request users.SearchUsersRequest) (*users.SearchUsersResponse, error) {
	return s.repository.SearchUsers(ctx, request)
}
