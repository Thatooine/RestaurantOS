package authentication

import (
	"context"
	"fmt"
	"time"

	pkgAuth "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type EmailAndPasswordAuthenticatorService struct {
	accessTokenCreator pkgAuth.AccessTokenCreator
	userRepository     users.UserRepository
}

func NewEmailAndPasswordAuthenticatorService(
	accessTokenCreator pkgAuth.AccessTokenCreator,
	userRepository users.UserRepository,
) *EmailAndPasswordAuthenticatorService {
	return &EmailAndPasswordAuthenticatorService{
		accessTokenCreator: accessTokenCreator,
		userRepository:     userRepository,
	}
}

// AuthenticateWithEmailAndPassword looks up the user by email and verifies the
// supplied password against the stored bcrypt hash, then issues a signed access
// token. Invalid email or password both surface as errs.ErrForbidden so the
// response never reveals which one was wrong.
func (s *EmailAndPasswordAuthenticatorService) AuthenticateWithEmailAndPassword(ctx context.Context, request pkgAuth.EmailAndPasswordAuthRequest) (*pkgAuth.EmailAndPasswordAuthResponse, error) {
	userResp, err := s.userRepository.GetUser(ctx, users.GetUserRequest{Email: request.Email})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to retrieve user by email")
		return nil, fmt.Errorf("invalid credentials: %w", errs.ErrForbidden)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userResp.User.PasswordHash), []byte(request.Password)); err != nil {
		log.Ctx(ctx).Warn().Msg("password mismatch")
		return nil, fmt.Errorf("invalid credentials: %w", errs.ErrForbidden)
	}

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
		return nil, fmt.Errorf("AuthenticateWithEmailAndPassword failed: %w", err)
	}

	return &pkgAuth.EmailAndPasswordAuthResponse{
		Token:  tokenResp.AccessToken,
		UserID: userResp.User.ID,
		Email:  userResp.User.Email,
	}, nil
}
