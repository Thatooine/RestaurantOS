package rateLimiting

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mennanov/limiters"
	"github.com/rs/zerolog/log"
)

type ipRateLimiterMiddleware struct {
	rateLimiter RedisTokenBucketRateLimiter
	refillRate  time.Duration
	capacity    int64
	ttl         time.Duration
}

// NewIpRateLimiterMiddleware rate limits requests by Request.RemoteAddr using a
// Redis-backed token bucket.
func NewIpRateLimiterMiddleware(
	rateLimiter RedisTokenBucketRateLimiter,
	capacity int64,
	refillRate time.Duration,
) gin.HandlerFunc {
	middleware := &ipRateLimiterMiddleware{
		rateLimiter: rateLimiter,
		refillRate:  refillRate,
		capacity:    capacity,
		ttl:         time.Duration(int64(refillRate) * capacity),
	}
	return middleware.Handle
}

func (m *ipRateLimiterMiddleware) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		ip = c.Request.RemoteAddr
	}

	key := "ip_token_bucket:" + ip

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
			log.Ctx(ctx).Warn().Str("ip", ip).Msg("ip rate limit exceeded")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		log.Ctx(ctx).Error().Err(err).Msg("rate limiting error")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Next()
}
