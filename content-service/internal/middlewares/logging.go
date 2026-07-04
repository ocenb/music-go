package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/logger"
)

func Logging(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		ctx := logger.IntoContext(c.Request.Context(), log.With(
			slog.String("http.method", c.Request.Method),
			slog.String("http.path", c.Request.URL.Path),
		))
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()

		reqLog := logger.FromContext(ctx)
		reqLog.Info("HTTP request",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
		)
	}
}
