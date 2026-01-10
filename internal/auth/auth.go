package auth

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kashifkhan/ai-gateway/internal/models"
)

const DefaultAPIKey = "sk-kashif-ai-gateway-secret-key-2024"

type Authenticator struct {
	keys    map[string]bool
	mu      sync.RWMutex
	enabled bool
}

func NewAuthenticator(keys []string, enabled bool) *Authenticator {
	auth := &Authenticator{
		keys:    make(map[string]bool),
		enabled: enabled,
	}

	for _, key := range keys {
		auth.keys[key] = true
	}

	if len(auth.keys) == 0 && enabled {
		auth.keys[DefaultAPIKey] = true
	}

	return auth
}

func (a *Authenticator) AddKey(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.keys[key] = true
}

func (a *Authenticator) RemoveKey(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.keys, key)
}

func (a *Authenticator) ValidateKey(key string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.keys[key]
}

func (a *Authenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.enabled {
			c.Next()
			return
		}

		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apiErr := models.ErrMissingAPIKey()
			c.JSON(apiErr.GetStatus(), apiErr)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			apiErr := models.ErrInvalidAPIKey()
			c.JSON(apiErr.GetStatus(), apiErr)
			c.Abort()
			return
		}

		apiKey := parts[1]

		if !a.ValidateKey(apiKey) {
			apiErr := models.ErrInvalidAPIKey()
			c.JSON(apiErr.GetStatus(), apiErr)
			c.Abort()
			return
		}

		c.Set("api_key", apiKey)
		c.Next()
	}
}

func GetDefaultKey() string {
	return DefaultAPIKey
}

type RateLimiter struct {
	enabled           bool
	requestsPerMinute int
}

func NewRateLimiter(enabled bool, requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		enabled:           enabled,
		requestsPerMinute: requestsPerMinute,
	}
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.enabled {
			c.Next()
			return
		}

		// TODO: Implement proper rate limiting
		c.Next()
	}
}
