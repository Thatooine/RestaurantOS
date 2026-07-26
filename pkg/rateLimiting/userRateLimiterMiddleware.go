package rateLimiting

import (
	"errors"
	"net/http"
	"time"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/gin-gonic/gin"
	"github.com/mennanov/limiters"
	"github.com/rs/zerolog/log"
)

type userRateLimiterMiddleware struct {
	rateLimiter RedisTokenBucketRateLimiter
	refillRate  time.Duration
	capacity    int64
	ttl         time.Duration
}

// NewUserRateLimiterMiddleware rate limits authenticated requests by the
// validated claim's UserID using a Redis-backed token bucket.
func NewUserRateLimiterMiddleware(
	rateLimiter RedisTokenBucketRateLimiter,
	capacity int64,
	refillRate time.Duration,
) gin.HandlerFunc {
	middleware := &userRateLimiterMiddleware{
		rateLimiter: rateLimiter,
		refillRate:  refillRate,
		capacity:    capacity,
		ttl:         time.Duration(int64(refillRate) * capacity),
	}
	return middleware.Handle
}

func (m *userRateLimiterMiddleware) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	claim, ok := authentication.LoginClaimFromGinContext(c)
	if !ok {
		// No login claim — let the auth middleware handle rejection.
		c.Next()
		return
	}

	key := "token_bucket:" + claim.UserID

	stateBackend, err := m.rateLimiter.TokenStateBackend(ctx, key, m.ttl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create token state backend")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	tokenBucket := m.rateLimiter.TokenBucket(
		ctx,
		TokenBucketRequest{
			RefillRate: m.refillRate,
			Capacity:   m.capacity,
		}, stateBackend.StateBackend)

	if _, err := m.rateLimiter.Limit(ctx, tokenBucket.TokenBucket); err != nil {
		if errors.Is(err, limiters.ErrLimitExhausted) {
			log.Ctx(ctx).Warn().Str("userID", claim.UserID).Msg("rate limit exceeded")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		log.Ctx(ctx).Error().Err(err).Msg("rate limiting error")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Next()
}
