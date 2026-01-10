package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server         ServerConfig             `yaml:"server"`
	Logging        LoggingConfig            `yaml:"logging"`
	Auth           AuthConfig               `yaml:"auth"`
	RateLimit      RateLimitConfig          `yaml:"rate_limit"`
	DefaultBackend string                   `yaml:"default_backend"`
	Backends       map[string]BackendConfig `yaml:"backends"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type AuthConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Keys     []string `yaml:"keys"`
	KeysFile string   `yaml:"keys_file"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	Burst             int  `yaml:"burst"`
}

type BackendConfig struct {
	Enabled bool          `yaml:"enabled"`
	Type    string        `yaml:"type"`
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
	Models  []ModelConfig `yaml:"models"`
}

type ModelConfig struct {
	ID      string   `yaml:"id"`
	Aliases []string `yaml:"aliases"`
	Free    bool     `yaml:"free"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 120 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Auth: AuthConfig{
			Enabled: true,
			Keys:    []string{},
		},
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 60,
			Burst:             10,
		},
		DefaultBackend: "opencode",
		Backends: map[string]BackendConfig{
			"opencode": {
				Enabled: true,
				Type:    "opencode",
				Host:    "localhost",
				Port:    3001,
				Timeout: 60 * time.Second,
				Models: []ModelConfig{
					{ID: "big-pickle", Aliases: []string{"pickle", "bp"}, Free: true},
					{ID: "grok-code-fast-1", Aliases: []string{"grok", "grok-fast"}, Free: true},
					{ID: "glm-4.7", Aliases: []string{"glm", "glm4"}, Free: true},
					{ID: "minimax-m2.1", Aliases: []string{"minimax", "mm"}, Free: true},
				},
			},
		},
	}
}

func Load(path string) (*Config, error) {
	godotenv.Load()

	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			yaml.Unmarshal(data, cfg)
		}
	}

	applyEnvOverrides(cfg)

	if cfg.Auth.KeysFile != "" {
		keys, err := loadKeysFile(cfg.Auth.KeysFile)
		if err != nil {
			return nil, err
		}
		cfg.Auth.Keys = append(cfg.Auth.Keys, keys...)
	}

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if host := os.Getenv("HOST"); host != "" {
		cfg.Server.Host = host
	}

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}

	if enabled := os.Getenv("AUTH_ENABLED"); enabled != "" {
		cfg.Auth.Enabled = enabled == "true" || enabled == "1"
	}

	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		cfg.Auth.Keys = append(cfg.Auth.Keys, apiKey)
	}

	if apiKey := os.Getenv("AI_GATEWAY_API_KEY"); apiKey != "" {
		cfg.Auth.Keys = append(cfg.Auth.Keys, apiKey)
	}

	if host := os.Getenv("OPENCODE_HOST"); host != "" {
		if backend, ok := cfg.Backends["opencode"]; ok {
			backend.Host = host
			cfg.Backends["opencode"] = backend
		}
	}

	if port := os.Getenv("OPENCODE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			if backend, ok := cfg.Backends["opencode"]; ok {
				backend.Port = p
				cfg.Backends["opencode"] = backend
			}
		}
	}

	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		cfg.RateLimit.Enabled = enabled == "true" || enabled == "1"
	}

	if rpm := os.Getenv("RATE_LIMIT_RPM"); rpm != "" {
		if r, err := strconv.Atoi(rpm); err == nil {
			cfg.RateLimit.RequestsPerMinute = r
		}
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}
}

type KeysFileConfig struct {
	Keys []KeyConfig `yaml:"keys"`
}

type KeyConfig struct {
	Key     string `yaml:"key"`
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
}

func loadKeysFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var keysFile KeysFileConfig
	if err := yaml.Unmarshal(data, &keysFile); err != nil {
		return nil, err
	}

	var keys []string
	for _, k := range keysFile.Keys {
		if k.Enabled {
			keys = append(keys, k.Key)
		}
	}

	return keys, nil
}
