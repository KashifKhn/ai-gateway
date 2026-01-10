package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kashifkhan/ai-gateway/internal/adapters"
	"github.com/kashifkhan/ai-gateway/internal/auth"
)

func SetupRouter(
	registry *adapters.Registry,
	authenticator *auth.Authenticator,
	rateLimiter *auth.RateLimiter,
	version string,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(rateLimiter.Middleware())
	router.Use(authenticator.Middleware())

	handler := NewHandler(registry, version)

	router.GET("/health", handler.Health)

	v1 := router.Group("/v1")
	{
		v1.GET("/models", handler.ListModels)
		v1.GET("/backends", handler.ListBackends)
		v1.POST("/chat/completions", handler.ChatCompletions)
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
