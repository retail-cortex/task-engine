package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserContextMiddleware extracts the user identity (either via X-User-ID header or a Bearer token)
// and binds it to both Gin context and Go standard context.
func UserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the user ID from "X-User-ID" header or mock JWT authorization header
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				// Simple mock token extractor for dev/test environment
				userID = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Fallback to a default system user if none is provided (specifically for internal mock scenarios)
		if userID == "" {
			userID = "00000000-0000-0000-0000-000000000000"
		}

		// Set in Gin Context (for handlers)
		c.Set("userID", userID)

		// Bind it to the http.Request Go Context so standard library context tools read it (essential for GORM transactional hooks)
		ctx := c.Request.Context()
		ctxWithUser := context.WithValue(ctx, "userID", userID)
		c.Request = c.Request.WithContext(ctxWithUser)

		c.Next()
	}
}

// CORSConfigMiddleware sets up permissive development headers to support modular local ports.
func CORSConfigMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
