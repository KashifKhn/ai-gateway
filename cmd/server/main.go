package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kashifkhan/ai-gateway/internal/adapters"
	"github.com/kashifkhan/ai-gateway/internal/adapters/opencode"
	"github.com/kashifkhan/ai-gateway/internal/api"
	"github.com/kashifkhan/ai-gateway/internal/auth"
	"github.com/kashifkhan/ai-gateway/internal/config"
)

const Version = "1.0.0"

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("Warning: Could not load config file: %v. Using defaults.", err)
		cfg = config.DefaultConfig()
	}

	printBanner()

	registry := adapters.NewRegistry(cfg.DefaultBackend)

	if backendCfg, ok := cfg.Backends["opencode"]; ok && backendCfg.Enabled {
		opencodeAdapter := opencode.New(backendCfg)
		if err := opencodeAdapter.Initialize(nil); err != nil {
			log.Printf("Warning: Failed to initialize OpenCode adapter: %v", err)
		} else {
			registry.Register(opencodeAdapter)
			log.Printf("✓ OpenCode adapter initialized (http://%s:%d)", backendCfg.Host, backendCfg.Port)
		}
	}

	authenticator := auth.NewAuthenticator(cfg.Auth.Keys, cfg.Auth.Enabled)
	if cfg.Auth.Enabled {
		log.Printf("✓ Authentication enabled")
		log.Printf("  Default API Key: %s", auth.GetDefaultKey())
	} else {
		log.Printf("⚠ Authentication disabled")
	}

	rateLimiter := auth.NewRateLimiter(cfg.RateLimit.Enabled, cfg.RateLimit.RequestsPerMinute)
	if cfg.RateLimit.Enabled {
		log.Printf("✓ Rate limiting enabled (%d req/min)", cfg.RateLimit.RequestsPerMinute)
	}

	router := api.SetupRouter(registry, authenticator, rateLimiter, Version)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("🚀 AI Gateway started on http://%s", addr)
		log.Printf("")
		log.Printf("Endpoints:")
		log.Printf("  GET  /health              - Health check")
		log.Printf("  GET  /v1/models           - List available models")
		log.Printf("  GET  /v1/backends         - List available backends")
		log.Printf("  POST /v1/chat/completions - Chat completion")
		log.Printf("")
		log.Printf("Example usage:")
		log.Printf("  curl -X POST http://%s/v1/chat/completions \\", addr)
		log.Printf("    -H 'Authorization: Bearer %s' \\", auth.GetDefaultKey())
		log.Printf("    -H 'Content-Type: application/json' \\")
		log.Printf("    -d '{\"model\": \"big-pickle\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello!\"}]}'")
		log.Printf("")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	registry.Shutdown()

	log.Println("Server exited")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════╗
║          KASHIF AI GATEWAY v%s            ║
║     Your Personal OpenAI Alternative          ║
╚═══════════════════════════════════════════════╝
`
	fmt.Printf(banner, Version)
	fmt.Println()
}
