package users

import (
	"context"
	"fmt"
	"testing"

	authMock "github.com/bash/the-dancing-pony-v2-rnyfbr/internal/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"golang.org/x/crypto/bcrypt"
)

// --- RegisterWithEmailAndPassword tests ---

func TestRegisterWithEmailAndPassword_Success(t *testing.T) {
	var createReq users.CreateUserRequest
	newUser := users.User{ID: "user-1", Name: "Frodo", Email: "frodo@shire.com", Roles: []users.Role{users.RoleCustomer}}

	repo := &UserRepositoryMock{
		GetUserFn: func(_ context.Context, _ users.GetUserRequest) (*users.GetUserResponse, error) {
			return nil, fmt.Errorf("not found: %w", errs.ErrNotFound)
		},
		CreateUserFn: func(_ context.Context, req users.CreateUserRequest) (*users.CreateUserResponse, error) {
			createReq = req
			return &users.CreateUserResponse{User: newUser}, nil
		},
	}

	tokenCreator := &authMock.AccessTokenCreatorServiceMock{
		CreateAccessTokenFn: func(_ context.Context, _ authentication.CreateAccessTokenRequest) (*authentication.CreateAccessTokenResponse, error) {
			return &authentication.CreateAccessTokenResponse{AccessToken: "jwt-123"}, nil
		},
	}

	svc := &UserRegistrationServiceImpl{
		accessTokenCreator: tokenCreator,
		userRepository:     repo,
	}

	resp, err := svc.RegisterWithEmailAndPassword(context.Background(), users.RegisterWithEmailAndPasswordRequest{
		Name:     "Frodo",
		Email:    "frodo@shire.com",
		Password: "one-ring",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "jwt-123" {
		t.Errorf("expected token jwt-123, got %s", resp.Token)
	}

	// The stored hash must verify against the original password and must not be the plaintext.
	if createReq.PasswordHash == "" {
		t.Fatal("expected a password hash to be stored")
	}
	if createReq.PasswordHash == "one-ring" {
		t.Fatal("password must be hashed, not stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(createReq.PasswordHash), []byte("one-ring")); err != nil {
		t.Errorf("stored hash does not match original password: %v", err)
	}
}

func TestRegisterWithEmailAndPassword_AlreadyExists(t *testing.T) {
	repo := &UserRepositoryMock{
		GetUserFn: func(_ context.Context, _ users.GetUserRequest) (*users.GetUserResponse, error) {
			return &users.GetUserResponse{User: users.User{ID: "user-1", Email: "frodo@shire.com"}}, nil
		},
		CreateUserFn: func(_ context.Context, _ users.CreateUserRequest) (*users.CreateUserResponse, error) {
			t.Fatal("CreateUser should not be called when the email already exists")
			return nil, nil
		},
	}

	svc := &UserRegistrationServiceImpl{userRepository: repo}

	_, err := svc.RegisterWithEmailAndPassword(context.Background(), users.RegisterWithEmailAndPasswordRequest{
		Name:     "Frodo",
		Email:    "frodo@shire.com",
		Password: "one-ring",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRegisterWithEmailAndPassword_ValidationFails(t *testing.T) {
	svc := &UserRegistrationServiceImpl{}

	// Missing password
	_, err := svc.RegisterWithEmailAndPassword(context.Background(), users.RegisterWithEmailAndPasswordRequest{
		Name:  "Frodo",
		Email: "frodo@shire.com",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// --- issueToken tests ---

func TestIssueToken_Success(t *testing.T) {
	tokenCreator := &authMock.AccessTokenCreatorServiceMock{
		CreateAccessTokenFn: func(_ context.Context, req authentication.CreateAccessTokenRequest) (*authentication.CreateAccessTokenResponse, error) {
			if req.LoginClaim.UserID != "user-1" {
				t.Fatalf("expected user ID user-1, got %s", req.LoginClaim.UserID)
			}
			if req.LoginClaim.Email != "frodo@shire.com" {
				t.Fatalf("expected email frodo@shire.com, got %s", req.LoginClaim.Email)
			}
			return &authentication.CreateAccessTokenResponse{AccessToken: "jwt-token-123"}, nil
		},
	}

	svc := &UserRegistrationServiceImpl{
		accessTokenCreator: tokenCreator,
	}

	user := users.User{ID: "user-1", Email: "frodo@shire.com"}
	resp, err := svc.issueToken(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "jwt-token-123" {
		t.Errorf("expected token jwt-token-123, got %s", resp.Token)
	}
	if resp.UserID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", resp.UserID)
	}
	if resp.Email != "frodo@shire.com" {
		t.Errorf("expected email frodo@shire.com, got %s", resp.Email)
	}
}

func TestIssueToken_TokenCreationFails(t *testing.T) {
	tokenCreator := &authMock.AccessTokenCreatorServiceMock{
		CreateAccessTokenFn: func(_ context.Context, _ authentication.CreateAccessTokenRequest) (*authentication.CreateAccessTokenResponse, error) {
			return nil, fmt.Errorf("signing error")
		},
	}

	svc := &UserRegistrationServiceImpl{
		accessTokenCreator: tokenCreator,
	}

	_, err := svc.issueToken(context.Background(), users.User{ID: "user-1", Email: "test@test.com"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
