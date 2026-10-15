package handler

import (
	"log"

	"github.com/gin-gonic/gin"
)

func LogRequests(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.Default()
	}

	return func(c *gin.Context) {
		logger.Printf("incoming request method=%s path=%s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}
