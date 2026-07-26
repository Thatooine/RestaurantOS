package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Middleware injects the global logger into Request.Context so contextual
// logging continues to work throughout Gin handlers and application services.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := log.Logger.WithContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
