package authentication

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/go-jose/go-jose/v4"
	"github.com/rs/zerolog/log"
)

// AccessTokenValidatorImpl validates signed JWT access tokens using go-jose.
type AccessTokenValidatorImpl struct {
	publicKey *rsa.PublicKey
}

// NewAccessTokenValidatorImpl returns a new AccessTokenValidatorImpl
// with the provided RSA public key for verifying token signatures.
func NewAccessTokenValidatorImpl(publicKey *rsa.PublicKey) *AccessTokenValidatorImpl {
	return &AccessTokenValidatorImpl{
		publicKey: publicKey,
	}
}

// ValidateAccessToken parses the compact-serialized JWT, verifies its signature,
// unmarshals the login claims, and checks that the token has not expired.
func (a *AccessTokenValidatorImpl) ValidateAccessToken(ctx context.Context, request authentication.ValidateAccessTokenRequest) (*authentication.ValidateAccessTokenResponse, error) {
	// parse the compact-serialized JWS
	signed, err := jose.ParseSigned(request.AccessToken, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("could not parse access token")
		return nil, fmt.Errorf("ValidateAccessToken failed: could not parse token: %w", err)
	}

	// verify the signature and extract the payload
	payload, err := signed.Verify(a.publicKey)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("could not verify access token signature")
		return nil, fmt.Errorf("ValidateAccessToken failed: could not verify token signature: %w", err)
	}

	// unmarshal the payload into LoginClaim
	var claim authentication.LoginClaim
	if err := json.Unmarshal(payload, &claim); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("could not unmarshal token claims")
		return nil, fmt.Errorf("ValidateAccessToken failed: could not unmarshal claims: %w", err)
	}

	// check token expiration
	if time.Now().Unix() > claim.ExpirationTime {
		log.Ctx(ctx).Warn().Str("userID", claim.UserID).Msg("access token has expired")
		return nil, fmt.Errorf("ValidateAccessToken failed: token has expired")
	}

	return &authentication.ValidateAccessTokenResponse{
		LoginClaim: claim,
	}, nil
}
