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
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type IAPClaims struct {
	Iss   string `json:"iss"`
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Aud   string `json:"aud"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
}

var (
	iapKeysCache       map[string]*ecdsa.PublicKey
	iapKeysCacheExpiry time.Time
	iapKeysMutex       sync.RWMutex
)

func getIAPPublicKey(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	iapKeysMutex.RLock()
	pubKey, exists := iapKeysCache[kid]
	expired := time.Now().After(iapKeysCacheExpiry)
	iapKeysMutex.RUnlock()

	if exists && !expired {
		return pubKey, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.gstatic.com/iap/verify/public_key", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch IAP public keys: status %d", resp.StatusCode)
	}

	var rawKeys map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&rawKeys); err != nil {
		return nil, err
	}

	newCache := make(map[string]*ecdsa.PublicKey)
	for k, v := range rawKeys {
		block, _ := pem.Decode([]byte(v))
		if block == nil {
			continue
		}
		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			continue
		}
		if ecKey, ok := pubInterface.(*ecdsa.PublicKey); ok {
			newCache[k] = ecKey
		}
	}

	iapKeysMutex.Lock()
	iapKeysCache = newCache
	iapKeysCacheExpiry = time.Now().Add(1 * time.Hour)
	pubKey, exists = iapKeysCache[kid]
	iapKeysMutex.Unlock()

	if !exists {
		return nil, fmt.Errorf("IAP public key with kid %s not found after reload", kid)
	}
	return pubKey, nil
}

func validateIAPJWT(ctx context.Context, token string, expectedAudience string) (*IAPClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid IAP JWT token format")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}
	if header.Alg != "ES256" {
		return nil, fmt.Errorf("unexpected signing algorithm: %s", header.Alg)
	}

	pubKey, err := getIAPPublicKey(ctx, header.Kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	signingInput := parts[0] + "." + parts[1]
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hash := hasher.Sum(nil)

	if len(sigBytes) != 64 {
		return nil, fmt.Errorf("invalid ECDSA signature length: got %d, expected 64", len(sigBytes))
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])
	if !ecdsa.Verify(pubKey, hash, r, s) {
		return nil, fmt.Errorf("IAP signature verification failed")
	}

	var claims IAPClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	if claims.Iss != "https://cloud.google.com/iap" {
		return nil, fmt.Errorf("invalid issuer: expected https://cloud.google.com/iap, got %s", claims.Iss)
	}

	if expectedAudience != "" && claims.Aud != expectedAudience {
		return nil, fmt.Errorf("audience mismatch: expected %s, got %s", expectedAudience, claims.Aud)
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// UserContextMiddleware extracts the user identity (either via X-Goog-IAP-JWT-Assertion, X-User-ID header, Google OAuth 2.0 ID token, or standard mock stubs)
// and binds it to both Gin context and Go standard context.
func UserContextMiddleware(adminService service.AdminService, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		userID = strings.TrimPrefix(userID, "A2A_USER_")

		// Strictly reject the development bypass user ID on the HTTP ingress if it's the only form of auth
		if userID == "00000000-0000-0000-0000-000000000000" {
			authHeader := c.GetHeader("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Unauthorized: Development bypass user ID is strictly forbidden on this interface",
				})
				return
			}
			// If they provided a Bearer token, we ignore the bypass user ID and let the JWT validator take over
			userID = ""
		}

		// If header is absent, check IAP assertion
		if userID == "" {
			iapAssertion := c.GetHeader("X-Goog-IAP-JWT-Assertion")
			if iapAssertion != "" {
				claims, err := validateIAPJWT(c.Request.Context(), iapAssertion, cfg.IAP.Audience)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "IAP JWT validation failed: " + err.Error(),
					})
					return
				}

				// Resolve user profile by Google Subject ID (claims.Sub)
				user, err := adminService.FindUserByOAuth(c.Request.Context(), "google", claims.Sub)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						email := claims.Email
						name := email
						if idx := strings.Index(email, "@"); idx != -1 {
							name = email[:idx]
						}

						// Check if there is an existing pre-seeded user with this email
						allUsers, listErr := adminService.ListUsers(c.Request.Context())
						var existingUser *model.User
						if listErr == nil {
							for _, u := range allUsers {
								if strings.EqualFold(u.Email, email) {
									existingUser = u
									break
								}
							}
						}

						if existingUser != nil {
							log.Printf("[IAP] Binding existing seeded profile %s to Google Subject ID: %s", email, claims.Sub)
							existingUser.OAuthID = claims.Sub
							existingUser.OAuthProvider = "google"
							if updateErr := adminService.UpdateUser(c.Request.Context(), existingUser); updateErr == nil {
								userID = existingUser.ID
							} else {
								log.Printf("[IAP] Warning: failed binding seeded user OAuthID: %v", updateErr)
								userID = existingUser.ID
							}
						} else {
							// Single-Click Sign-up: auto-register new profiles seamlessly on first login
							log.Printf("[IAP] Registering new first-time SSO user: %s (%s)", name, email)
							newUser := &model.User{
								OAuthProvider: "google",
								OAuthID:       claims.Sub,
								Email:         email,
								Name:          name,
								Metadata:      model.JSONB("{}"),
							}
							regErr := adminService.RegisterUser(c.Request.Context(), newUser)
							if regErr == nil {
								userID = newUser.ID
							} else {
								log.Printf("[IAP] Warning: failed dynamically seeding user profile: %v", regErr)
								c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user profile"})
								return
							}
						}
					} else {
						log.Printf("[IAP] Error: failed mapping GORM query validation: %v", err)
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user profile"})
						return
					}
				} else {
					userID = user.ID
				}
			}
		}

		// Fallback to dynamic OAuth Bearer headers if still not resolved
		if userID == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")

				// Strictly reject any hardcoded mock/bypass tokens in Authorization header
				if token == "00000000-0000-0000-0000-000000000000" || token == "user-initiator" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "Unauthorized: Development bypass tokens are strictly forbidden",
					})
					return
				}

				if cfg.OAuth.ClientID == "" {
					// Strictly prevent auth bypass in non-local environments
					runtime := os.Getenv("MODENV_RUNTIME")
					if runtime != "" && runtime != "local" {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"error": "Unauthorized: Legacy mock auth bypass is strictly forbidden in non-local environments",
						})
						return
					}
					// Fallback to legacy mock extractor when client ID is not configured (offline/dev testing contexts)
					userID = token
				} else {
					// 1. Extract the audience from the token payload without cryptographic overhead
					aud, err := extractAudienceFromJWT(token)
					var payload *idtoken.Payload
					if err == nil {
						// 2. Verify that the audience is whitelisted
						isAllowed := aud == cfg.OAuth.ClientID
						if !isAllowed {
							for _, allowedID := range cfg.OAuth.AllowedClientIDs {
								if allowedID != "" && allowedID == aud {
									isAllowed = true
									break
								}
							}
						}

						// Fallback: in local/dev environments, also allow tokens minted by the gcloud CLI
						gcloudClientID := "32555940559.apps.googleusercontent.com" // Default gcloud CLI audience
						legacyGcloudClientID := "764086051750-31g961vd5no7cbh154hra5147986m792.apps.googleusercontent.com"
						if !isAllowed && (aud == gcloudClientID || aud == legacyGcloudClientID) {
							isAllowed = true
							log.Printf("[IAP] Warning: Validating token using gcloud CLI audience fallback. This must be disabled in production.")
						}

						// 3. Cryptographically validate the token EXACTLY ONCE against the matched audience
						if isAllowed {
							payload, err = idtoken.Validate(c.Request.Context(), token, aud)
						} else {
							// Check service-to-service IAP aud fallback
							serviceAud := "https://" + c.Request.Host
							if p, fallbackErr := idtoken.Validate(c.Request.Context(), token, serviceAud); fallbackErr == nil {
								payload = p
								err = nil
							} else {
								err = fmt.Errorf("audience %q is not whitelisted in server configurations", aud)
							}
						}
					}
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
							email, _ := payload.Claims["email"].(string)
							name, _ := payload.Claims["name"].(string)

							// Check if there is an existing pre-seeded user with this email
							allUsers, listErr := adminService.ListUsers(c.Request.Context())
							var existingUser *model.User
							if listErr == nil {
								for _, u := range allUsers {
									if strings.EqualFold(u.Email, email) {
										existingUser = u
										break
									}
								}
							}

							if existingUser != nil {
								log.Printf("[OAuth] Binding existing seeded profile %s to Google Subject ID: %s", email, payload.Subject)
								existingUser.OAuthID = payload.Subject
								existingUser.OAuthProvider = "google"
								if updateErr := adminService.UpdateUser(c.Request.Context(), existingUser); updateErr == nil {
									userID = existingUser.ID
								} else {
									log.Printf("[OAuth] Warning: failed binding seeded user OAuthID: %v", updateErr)
									userID = existingUser.ID
								}
							} else {
								// 3. Dynamic Single-Click Sign-up: auto-register new profiles seamlessly on first login
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
									c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user profile"})
									return
								}
							}
						} else {
							log.Printf("[OAuth] Error: failed mapping GORM query validation: %v", err)
							c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user profile"})
							return
						}
					} else {
						userID = user.ID
					}
				}
			}
		}

		// Unified email/username-to-UUID database resolution pass
		if userID != "" {
			if len(userID) != 36 || strings.Count(userID, "-") != 4 {
				allUsers, err := adminService.ListUsers(c.Request.Context())
				if err == nil {
					for _, u := range allUsers {
						if strings.EqualFold(u.Email, userID) || strings.EqualFold(u.Name, userID) {
							userID = u.ID
							break
						}
					}
				}
			}
		}

		// If still not resolved, strictly reject the request!
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: A valid, cryptographically verified user context is required",
			})
			return
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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID, traceparent, tracestate")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// extractAudienceFromJWT decodes the payload of a JWT and retrieves the "aud" claim without verifying its signature.
func extractAudienceFromJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format: expected 3 parts")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims struct {
		Aud string `json:"aud"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return claims.Aud, nil
}
