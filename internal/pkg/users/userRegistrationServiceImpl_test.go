package users

import (
	"context"
	"fmt"
	"testing"

	authMock "github.com/bash/the-dancing-pony-v2-rnyfbr/internal/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
)

// --- getOrCreateUser tests ---

func TestGetOrCreateUser_ReturnsExistingUser(t *testing.T) {
	existingUser := users.User{ID: "user-1", Name: "Frodo", Email: "frodo@shire.com", Roles: []users.Role{users.RoleCustomer}}

	repo := &UserRepositoryMock{
		GetUserFn: func(_ context.Context, req users.GetUserRequest) (*users.GetUserResponse, error) {
			if req.Email != "frodo@shire.com" {
				t.Fatalf("expected email frodo@shire.com, got %s", req.Email)
			}
			return &users.GetUserResponse{User: existingUser}, nil
		},
		CreateUserFn: func(_ context.Context, _ users.CreateUserRequest) (*users.CreateUserResponse, error) {
			t.Fatal("CreateUser should not be called when user exists")
			return nil, nil
		},
	}

	svc := &UserRegistrationServiceImpl{
		userRepository: repo,
	}

	user, err := svc.getOrCreateUser(context.Background(), "Frodo", "frodo@shire.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", user.ID)
	}
}

func TestGetOrCreateUser_CreatesNewUser(t *testing.T) {
	newUser := users.User{ID: "user-2", Name: "Sam", Email: "sam@shire.com", Roles: []users.Role{users.RoleCustomer}}

	repo := &UserRepositoryMock{
		GetUserFn: func(_ context.Context, _ users.GetUserRequest) (*users.GetUserResponse, error) {
			return nil, fmt.Errorf("user not found")
		},
		CreateUserFn: func(_ context.Context, req users.CreateUserRequest) (*users.CreateUserResponse, error) {
			if req.Name != "Sam" || req.Email != "sam@shire.com" {
				t.Fatalf("unexpected create request: %+v", req)
			}
			return &users.CreateUserResponse{User: newUser}, nil
		},
	}

	svc := &UserRegistrationServiceImpl{
		userRepository: repo,
	}

	user, err := svc.getOrCreateUser(context.Background(), "Sam", "sam@shire.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-2" {
		t.Errorf("expected user ID user-2, got %s", user.ID)
	}
}

func TestGetOrCreateUser_CreateFails(t *testing.T) {
	repo := &UserRepositoryMock{
		GetUserFn: func(_ context.Context, _ users.GetUserRequest) (*users.GetUserResponse, error) {
			return nil, fmt.Errorf("user not found")
		},
		CreateUserFn: func(_ context.Context, _ users.CreateUserRequest) (*users.CreateUserResponse, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	svc := &UserRegistrationServiceImpl{
		userRepository: repo,
	}

	_, err := svc.getOrCreateUser(context.Background(), "Gandalf", "gandalf@istari.com")
	if err == nil {
		t.Fatal("expected error, got nil")
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

// --- Full registration flow tests (getOrCreateUser + issueToken) ---

func TestRegistrationFlow_NewUser(t *testing.T) {
	newUser := users.User{ID: "user-1", Name: "Frodo", Email: "frodo@shire.com", Roles: []users.Role{users.RoleCustomer}}

	repo := &UserRepositoryMock{
		GetUserFn: func(_ context.Context, _ users.GetUserRequest) (*users.GetUserResponse, error) {
			return nil, fmt.Errorf("not found")
		},
		CreateUserFn: func(_ context.Context, _ users.CreateUserRequest) (*users.CreateUserResponse, error) {
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

	user, err := svc.getOrCreateUser(context.Background(), "Frodo", "frodo@shire.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := svc.issueToken(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "jwt-123" {
		t.Errorf("expected token jwt-123, got %s", resp.Token)
	}
	if resp.UserID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", resp.UserID)
	}
}

func TestRegistrationFlow_RecoveryFromPartialRegistration(t *testing.T) {
	// User already exists in DB (Firebase succeeded before, DB creation succeeded, but token failed)
	existingUser := users.User{ID: "user-1", Name: "Frodo", Email: "frodo@shire.com", Roles: []users.Role{users.RoleCustomer}}

	createCalled := false
	repo := &UserRepositoryMock{
		GetUserFn: func(_ context.Context, req users.GetUserRequest) (*users.GetUserResponse, error) {
			if req.Email == "frodo@shire.com" {
				return &users.GetUserResponse{User: existingUser}, nil
			}
			return nil, fmt.Errorf("not found")
		},
		CreateUserFn: func(_ context.Context, _ users.CreateUserRequest) (*users.CreateUserResponse, error) {
			createCalled = true
			t.Fatal("CreateUser should not be called when user already exists")
			return nil, nil
		},
	}

	tokenCreator := &authMock.AccessTokenCreatorServiceMock{
		CreateAccessTokenFn: func(_ context.Context, _ authentication.CreateAccessTokenRequest) (*authentication.CreateAccessTokenResponse, error) {
			return &authentication.CreateAccessTokenResponse{AccessToken: "recovered-jwt"}, nil
		},
	}

	svc := &UserRegistrationServiceImpl{
		accessTokenCreator: tokenCreator,
		userRepository:     repo,
	}

	// Simulate retry: user exists, should skip creation
	user, err := svc.getOrCreateUser(context.Background(), "Frodo", "frodo@shire.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected existing user ID, got %s", user.ID)
	}

	resp, err := svc.issueToken(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "recovered-jwt" {
		t.Errorf("expected recovered-jwt, got %s", resp.Token)
	}
	if createCalled {
		t.Error("CreateUser should not have been called")
	}
}
