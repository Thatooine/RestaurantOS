package users

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/rs/zerolog/log"
)

type UserRegistrationServiceImpl struct {
	firebaseApp        *firebase.App
	accessTokenCreator authentication.AccessTokenCreator
	userRepository     users.UserRepository
	firebaseAPIKey     string
}

func NewUserRegistrationServiceImpl(
	firebaseApp *firebase.App,
	accessTokenCreator authentication.AccessTokenCreator,
	userRepository users.UserRepository,
	firebaseAPIKey string,
) *UserRegistrationServiceImpl {
	return &UserRegistrationServiceImpl{
		firebaseApp:        firebaseApp,
		accessTokenCreator: accessTokenCreator,
		userRepository:     userRepository,
		firebaseAPIKey:     firebaseAPIKey,
	}
}

// RegisterWithEmailAndPassword creates the user on Firebase via the Identity Toolkit
// signUp API, then creates the user in our database and issues a JWT.
func (s *UserRegistrationServiceImpl) RegisterWithEmailAndPassword(ctx context.Context, request users.RegisterWithEmailAndPasswordRequest) (*users.RegisterResponse, error) {
	firebaseResp, err := s.signUpFirebase(ctx, request.Email, request.Password)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("firebase sign up failed")
		return nil, fmt.Errorf("RegisterWithEmailAndPassword failed: %w", err)
	}

	user, err := s.getOrCreateUser(ctx, request.Name, firebaseResp.Email)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get or create user")
		return nil, fmt.Errorf("RegisterWithEmailAndPassword failed: %w", err)
	}

	return s.issueToken(ctx, user)
}

// RegisterWithFirebaseToken verifies the Firebase ID token (the user already
// exists on Firebase from Google sign-up), creates the user in our database,
// and issues a JWT.
func (s *UserRegistrationServiceImpl) RegisterWithFirebaseToken(ctx context.Context, request users.RegisterWithFirebaseTokenRequest) (*users.RegisterResponse, error) {
	authClient, err := s.firebaseApp.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("RegisterWithFirebaseToken failed: %w", err)
	}

	verifiedToken, err := authClient.VerifyIDToken(ctx, request.FirebaseToken)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("token verification failed")
		return nil, fmt.Errorf("RegisterWithFirebaseToken failed: %w", err)
	}

	userRecord, err := authClient.GetUser(ctx, verifiedToken.UID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get firebase user")
		return nil, fmt.Errorf("RegisterWithFirebaseToken failed: %w", err)
	}

	name := request.Name
	if name == "" {
		name = userRecord.DisplayName
	}

	user, err := s.getOrCreateUser(ctx, name, userRecord.Email)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get or create user")
		return nil, fmt.Errorf("RegisterWithFirebaseToken failed: %w", err)
	}

	return s.issueToken(ctx, user)
}

// getOrCreateUser checks if a user with the given email already exists (recovering
// from a previous partial registration). If not, it creates a new user.
func (s *UserRegistrationServiceImpl) getOrCreateUser(ctx context.Context, name string, email string) (users.User, error) {
	existingResp, err := s.userRepository.GetUser(ctx, users.GetUserRequest{Email: email})
	if err == nil {
		return existingResp.User, nil
	}

	createResp, err := s.userRepository.CreateUser(
		ctx,
		users.CreateUserRequest{
			Name:  name,
			Email: email,
		})
	if err != nil {
		return users.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return createResp.User, nil
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

// Firebase Identity Toolkit sign-up types

type firebaseSignUpRequestBody struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

type firebaseSignUpResponseBody struct {
	IDToken      string `json:"idToken"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
}

func (s *UserRegistrationServiceImpl) signUpFirebase(ctx context.Context, email string, password string) (*firebaseSignUpResponseBody, error) {
	bodyData, err := json.Marshal(firebaseSignUpRequestBody{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshalling sign up request body: %w", err)
	}

	url := fmt.Sprintf(
		"https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=%s",
		s.firebaseAPIKey,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, fmt.Errorf("error constructing sign up request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing sign up request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("firebase sign up failed with status %d", resp.StatusCode)
	}

	var signUpResp firebaseSignUpResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&signUpResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling sign up response: %w", err)
	}

	return &signUpResp, nil
}
