package config

import (
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

const Version = "0.1.0"

type Config struct {
	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	DatabaseURL string `env:"DATABASE_URL,required"`

	JWTSecret  string `env:"JWT_SECRET,required"`
	APIKeySalt string `env:"API_KEY_SALT,required"`

	OpenRouterAPIKey  string `env:"OPENROUTER_API_KEY"`
	OpenRouterBaseURL string `env:"OPENROUTER_BASE_URL" envDefault:"https://openrouter.ai/api/v1"`
	OpenRouterModelFast  string `env:"OPENROUTER_MODEL_FAST" envDefault:"google/gemini-flash-1.5-8b"`
	OpenRouterModelSmart string `env:"OPENROUTER_MODEL_SMART" envDefault:"anthropic/claude-sonnet-4-5"`
	OpenRouterModelEmbed string `env:"OPENROUTER_MODEL_EMBED" envDefault:"openai/text-embedding-3-small"`

	AutomationSavingsRate float64 `env:"AUTOMATION_SAVINGS_RATE" envDefault:"0.7"`

	WorkerStrategyConcurrency int `env:"WORKER_STRATEGY_CONCURRENCY" envDefault:"3"`
	WorkerProposalConcurrency int `env:"WORKER_PROPOSAL_CONCURRENCY" envDefault:"3"`

	StorageBackend string `env:"STORAGE_BACKEND" envDefault:"local"`
	Domain         string `env:"DOMAIN" envDefault:"erman.ai"`

	startedAt time.Time
}

func Load() (*Config, error) {
	cfg := &Config{startedAt: time.Now()}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = buildDatabaseURL()
	}
	return cfg, nil
}

func buildDatabaseURL() string {
	host := getEnv("POSTGRES_HOST", "postgres")
	port := getEnv("POSTGRES_PORT", "5432")
	db := getEnv("POSTGRES_DB", "erman_ai")
	user := getEnv("POSTGRES_USER", "erman_ai")
	pass := os.Getenv("POSTGRES_PASSWORD")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (c *Config) UptimeSeconds() int64 {
	return int64(time.Since(c.startedAt).Seconds())
}

func (c *Config) PublicBaseURL() string {
	if c.Environment == "development" {
		return "http://localhost"
	}
	return "https://" + c.Domain
}
