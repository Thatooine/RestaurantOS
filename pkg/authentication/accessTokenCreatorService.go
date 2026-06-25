package authentication

import "context"

type AccessTokenCreator interface {
	CreateAccessToken(ctx context.Context, request CreateAccessTokenRequest) (*CreateAccessTokenResponse, error)
}

type CreateAccessTokenRequest struct {
	LoginClaim LoginClaim
}

type CreateAccessTokenResponse struct {
	AccessToken string
}
