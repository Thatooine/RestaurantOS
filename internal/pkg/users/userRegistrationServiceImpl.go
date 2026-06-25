package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserRegistrationServiceImpl struct {
	accessTokenCreator authentication.AccessTokenCreator
	userRepository     users.UserRepository
}

func NewUserRegistrationServiceImpl(
	accessTokenCreator authentication.AccessTokenCreator,
	userRepository users.UserRepository,
) *UserRegistrationServiceImpl {
	return &UserRegistrationServiceImpl{
		accessTokenCreator: accessTokenCreator,
		userRepository:     userRepository,
	}
}

// RegisterWithEmailAndPassword creates a new user with a bcrypt-hashed password
// and issues a JWT. It fails if a user with the same email already exists.
func (s *UserRegistrationServiceImpl) RegisterWithEmailAndPassword(ctx context.Context, request users.RegisterWithEmailAndPasswordRequest) (*users.RegisterResponse, error) {
	if err := validateRegistration(request); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for RegisterWithEmailAndPassword: %w", err)
	}

	// Reject registration if a user already exists for this email.
	_, err := s.userRepository.GetUser(ctx, users.GetUserRequest{Email: request.Email})
	switch {
	case err == nil:
		return nil, fmt.Errorf("a user with this email already exists: %w", errs.ErrConflict)
	case errors.Is(err, errs.ErrNotFound):
		// expected — continue with registration
	default:
		log.Ctx(ctx).Error().Err(err).Msg("failed to check existing user")
		return nil, fmt.Errorf("RegisterWithEmailAndPassword failed: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to hash password")
		return nil, fmt.Errorf("RegisterWithEmailAndPassword failed: %w", err)
	}

	createResp, err := s.userRepository.CreateUser(ctx, users.CreateUserRequest{
		Name:         request.Name,
		Email:        request.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create user")
		return nil, fmt.Errorf("RegisterWithEmailAndPassword failed: %w", err)
	}

	return s.issueToken(ctx, createResp.User)
}

func validateRegistration(request users.RegisterWithEmailAndPasswordRequest) error {
	var reasons []string

	if request.Name == "" {
		reasons = append(reasons, "Name is required")
	}

	if request.Email == "" {
		reasons = append(reasons, "Email is required")
	}

	if request.Password == "" {
		reasons = append(reasons, "Password is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (s *UserRegistrationServiceImpl) issueToken(ctx context.Context, user users.User) (*users.RegisterResponse, error) {
	loginClaim := authentication.LoginClaim{
		UserID:         user.ID,
		Email:          user.Email,
		ExpirationTime: time.Now().Add(1 * time.Hour).Unix(),
	}

	tokenResp, err := s.accessTokenCreator.CreateAccessToken(ctx, authentication.CreateAccessTokenRequest{
		LoginClaim: loginClaim,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	return &users.RegisterResponse{
		Token:  tokenResp.AccessToken,
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}
