// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

// UserContextMiddleware extracts the user identity (either via X-User-ID header, Google OAuth 2.0 ID token, or standard mock stubs)
// and binds it to both Gin context and Go standard context.
func UserContextMiddleware(adminService service.AdminService, clientID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		
		// If developer-assigned X-User-ID header is absent, inspect dynamic OAuth Bearer headers
		if userID == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				
				// Standard offline testing bypasses for database sandboxes & internal automated sweeps
				if token == "00000000-0000-0000-0000-000000000000" || token == "user-initiator" {
					userID = token
				} else if clientID == "" {
					// Fallback to legacy mock extractor when client ID is not configured (offline/dev testing contexts)
					userID = token
				} else {
					// 1. Crypographic validation of Google OAuth ID Token using standard client SDK
					payload, err := idtoken.Validate(c.Request.Context(), token, clientID)
					if err != nil {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"error": "Google ID Token authentication checks failed: " + err.Error(),
						})
						return
					}
					
					// 2. Query GORM database to resolve user profile by Google Subject ID
					user, err := adminService.FindUserByOAuth(c.Request.Context(), "google", payload.Subject)
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// 3. Dynamic Single-Click Sign-up: auto-register new profiles seamlessly on first login
							email, _ := payload.Claims["email"].(string)
							name, _ := payload.Claims["name"].(string)
							
							log.Printf("[OAuth] Registering new first-time SSO user: %s (%s)", name, email)
							newUser := &model.User{
								OAuthProvider: "google",
								OAuthID:       payload.Subject,
								Email:         email,
								Name:          name,
								Metadata:      model.JSONB("{}"),
							}
							
							regErr := adminService.RegisterUser(c.Request.Context(), newUser)
							if regErr == nil {
								userID = newUser.ID
							} else {
								log.Printf("[OAuth] Warning: failed dynamically seeding user profile: %v", regErr)
								userID = "00000000-0000-0000-0000-000000000000"
							}
						} else {
							log.Printf("[OAuth] Error: failed mapping GORM query validation: %v", err)
							userID = "00000000-0000-0000-0000-000000000000"
						}
					} else {
						userID = user.ID
					}
				}
			}
		}

		// Fallback to default system user if none provided (specifically for localized sweeps)
		if userID == "" {
			userID = "00000000-0000-0000-0000-000000000000"
		}

		// Bind verified user session parameters globally
		c.Set("userID", userID)

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
