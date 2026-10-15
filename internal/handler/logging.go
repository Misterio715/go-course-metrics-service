package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func LogRequests(logger *zerolog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = &log.Logger
	}

	return func(ctx *gin.Context) {
		startedAt := time.Now()
		ctx.Next()

		logger.Info().
			Str("uri", ctx.Request.RequestURI).
			Str("method", ctx.Request.Method).
			Dur("duration_ms", time.Since(startedAt)).
			Int("status", ctx.Writer.Status()).
			Int("size", max(0, ctx.Writer.Size())).
			Msg("HTTP request")
	}
}
