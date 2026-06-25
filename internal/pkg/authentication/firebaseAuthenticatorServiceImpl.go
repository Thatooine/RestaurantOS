package authentication

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	pkgAuth "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/rs/zerolog/log"
)

type FirebaseAuthenticatorService struct {
	firebaseApp        *firebase.App
	accessTokenCreator pkgAuth.AccessTokenCreator
	userRepository     users.UserRepository
}

func NewFirebaseAuthenticatorService(
	firebaseApp *firebase.App,
	accessTokenCreator pkgAuth.AccessTokenCreator,
	userRepository users.UserRepository,
) *FirebaseAuthenticatorService {
	return &FirebaseAuthenticatorService{
		firebaseApp:        firebaseApp,
		accessTokenCreator: accessTokenCreator,
		userRepository:     userRepository,
	}
}

func (s *FirebaseAuthenticatorService) AuthenticateWithFirebaseToken(ctx context.Context, request pkgAuth.FirebaseAuthRequest) (*pkgAuth.FirebaseAuthResponse, error) {
	// Verify the Firebase ID token
	verifiedToken, err := s.verifyFirebaseToken(ctx, request.FirebaseToken)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("token verification failed")
		return nil, fmt.Errorf("AuthenticateWithFirebaseToken failed: %w", err)
	}

	// Fetch the Firebase user using the verified token UID
	userRecord, err := s.fetchFirebaseUserViaToken(ctx, verifiedToken.UID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("user fetch failed")
		return nil, fmt.Errorf("AuthenticateWithFirebaseToken failed: %w", err)
	}

	// Retrieve the user from the database by email
	userResp, err := s.userRepository.GetUser(ctx, users.GetUserRequest{Email: userRecord.Email})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to retrieve user by email")
		return nil, fmt.Errorf("AuthenticateWithFirebaseToken failed: %w", err)
	}

	// Construct user claims and issue a signed token
	loginClaim := pkgAuth.LoginClaim{
		UserID:         userResp.User.ID,
		Email:          userResp.User.Email,
		ExpirationTime: time.Now().Add(1 * time.Hour).Unix(),
	}

	tokenResp, err := s.accessTokenCreator.CreateAccessToken(
		ctx,
		pkgAuth.CreateAccessTokenRequest{
			LoginClaim: loginClaim,
		})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create access token")
		return nil, fmt.Errorf("AuthenticateWithFirebaseToken failed: %w", err)
	}

	return &pkgAuth.FirebaseAuthResponse{
		Token:  tokenResp.AccessToken,
		UserID: userResp.User.ID,
		Email:  userResp.User.Email,
	}, nil
}

// verifyFirebaseToken verifies the Firebase ID token and returns the decoded token payload.
func (s *FirebaseAuthenticatorService) verifyFirebaseToken(ctx context.Context, idToken string) (*auth.Token, error) {
	authClient, err := s.firebaseApp.Auth(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error getting auth client instance")
		return nil, fmt.Errorf("error getting auth client instance: %w", err)
	}

	verifiedToken, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error verifying firebase ID token")
		return nil, fmt.Errorf("error verifying firebase ID token: %w", err)
	}

	return verifiedToken, nil
}

// fetchFirebaseUserViaToken retrieves the Firebase user record using the UID from a verified token.
func (s *FirebaseAuthenticatorService) fetchFirebaseUserViaToken(ctx context.Context, uid string) (*auth.UserRecord, error) {
	authClient, err := s.firebaseApp.Auth(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error getting auth client instance")
		return nil, fmt.Errorf("error getting auth client instance: %w", err)
	}

	userRecord, err := authClient.GetUser(ctx, uid)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error getting firebase user")
		return nil, fmt.Errorf("error getting firebase user: %w", err)
	}

	return userRecord, nil
}
