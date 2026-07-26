package authentication

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ginContextKey string

const loginClaimContextKey ginContextKey = "loginClaim"

// NewAuthMiddleware checks for an access token in the Authorization header
// ("Bearer <token>") or in the "access_token" cookie. Valid claims are stored
// in Gin context for authenticated handlers.
func NewAuthMiddleware(validator AccessTokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		token := extractToken(c.Request)
		if token == "" {
			log.Ctx(ctx).Warn().Msg("no access token found in request")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := validator.ValidateAccessToken(ctx, ValidateAccessTokenRequest{
			AccessToken: token,
		})
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("access token validation failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(loginClaimContextKey, resp.LoginClaim)
		c.Next()
	}
}

// LoginClaimFromGinContext retrieves the LoginClaim stored in Gin context.
// Returns the claim and true if present, or a zero value and false otherwise.
func LoginClaimFromGinContext(c *gin.Context) (LoginClaim, bool) {
	value, exists := c.Get(loginClaimContextKey)
	if !exists {
		return LoginClaim{}, false
	}
	claim, ok := value.(LoginClaim)
	return claim, ok
}

// extractToken looks for an access token first in the Authorization header
// (expecting "Bearer <token>"), then in a cookie named "access_token".
func extractToken(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if token, found := strings.CutPrefix(authHeader, "Bearer "); found {
			return token
		}
	}

	if cookie, err := r.Cookie("access_token"); err == nil {
		return cookie.Value
	}

	return ""
}
