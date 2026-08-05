package config

import (
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

const Version = "0.1.0"

type Config struct {
	ServerPort  string `env:"SERVER_PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	DatabaseURL string `env:"DATABASE_URL,required"`

	JWTSecret  string `env:"JWT_SECRET,required"`
	APIKeySalt string `env:"API_KEY_SALT,required"`

	OpenRouterAPIKey     string `env:"OPENROUTER_API_KEY"`
	OpenRouterBaseURL    string `env:"OPENROUTER_BASE_URL" envDefault:"https://openrouter.ai/api/v1"`
	OpenRouterModelFast  string `env:"OPENROUTER_MODEL_FAST" envDefault:"google/gemini-flash-1.5-8b"`
	OpenRouterModelSmart string `env:"OPENROUTER_MODEL_SMART" envDefault:"anthropic/claude-sonnet-4-5"`
	OpenRouterModelEmbed string `env:"OPENROUTER_MODEL_EMBED" envDefault:"openai/text-embedding-3-small"`

	DeepSeekAPIKey       string `env:"DEEPSEEK_API_KEY"`
	DeepSeekBaseURL      string `env:"DEEPSEEK_BASE_URL" envDefault:"https://api.deepseek.com/v1"`
	DeepSeekModelDefault string `env:"DEEPSEEK_MODEL_DEFAULT" envDefault:"deepseek-chat"`

	YandexAPIKey     string `env:"YANDEX_API_KEY"`
	YandexFolderID   string `env:"YANDEX_FOLDER_ID"`
	YandexBaseURL    string `env:"YANDEX_BASE_URL" envDefault:"https://llm.api.cloud.yandex.net/v1"`
	YandexModelSmart string `env:"YANDEX_MODEL_SMART" envDefault:"yandexgpt/latest"`
	YandexModelFast  string `env:"YANDEX_MODEL_FAST" envDefault:"yandexgpt-lite/latest"`

	YandexPriceInputRUBPer1K      float64 `env:"YANDEX_PRICE_INPUT_RUB_PER_1K" envDefault:"0.6"`
	YandexPriceOutputRUBPer1K     float64 `env:"YANDEX_PRICE_OUTPUT_RUB_PER_1K" envDefault:"1.8"`
	OpenRouterPriceInputUSDPer1K  float64 `env:"OPENROUTER_PRICE_INPUT_USD_PER_1K" envDefault:"0.003"`
	OpenRouterPriceOutputUSDPer1K float64 `env:"OPENROUTER_PRICE_OUTPUT_USD_PER_1K" envDefault:"0.015"`
	DeepSeekPriceInputUSDPer1K    float64 `env:"DEEPSEEK_PRICE_INPUT_USD_PER_1K" envDefault:"0.014"`
	DeepSeekPriceOutputUSDPer1K   float64 `env:"DEEPSEEK_PRICE_OUTPUT_USD_PER_1K" envDefault:"0.028"`

	AutomationSavingsRate float64 `env:"AUTOMATION_SAVINGS_RATE" envDefault:"0.7"`

	WorkerStrategyConcurrency int `env:"WORKER_STRATEGY_CONCURRENCY" envDefault:"3"`
	WorkerProposalConcurrency int `env:"WORKER_PROPOSAL_CONCURRENCY" envDefault:"3"`

	StorageBackend             string `env:"STORAGE_BACKEND" envDefault:"local"`
	TelegramAssetsDir          string `env:"TELEGRAM_ASSETS_DIR" envDefault:"/app/data/telegram"`
	GeologicalJournalAssetsDir    string `env:"GEOLOGICAL_JOURNAL_ASSETS_DIR" envDefault:"/app/data/geological-journal"`
	AudioTranscriptionAssetsDir   string `env:"AUDIO_TRANSCRIPTION_ASSETS_DIR" envDefault:"/app/data/audio-transcription"`
	JournalPreprocessorURL        string `env:"JOURNAL_PREPROCESSOR_URL"`
	Domain                     string `env:"DOMAIN" envDefault:"erman.ai"`

	WorkspaceOIDCClientID      string `env:"WORKSPACE_OIDC_CLIENT_ID" envDefault:"erman-affine"`
	WorkspaceOIDCClientSecret  string `env:"WORKSPACE_OIDC_CLIENT_SECRET"`
	WorkspaceOIDCPrivateKeyB64 string `env:"WORKSPACE_OIDC_PRIVATE_KEY_B64"`
	WorkspaceOIDCRedirectURI   string `env:"WORKSPACE_OIDC_REDIRECT_URI"`

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

func (c *Config) WorkspaceOIDCIssuer() string {
	return c.PublicBaseURL() + "/api/v1/workspace/oidc"
}

func (c *Config) WorkspaceOIDCCallbackURL() string {
	if c.WorkspaceOIDCRedirectURI != "" {
		return c.WorkspaceOIDCRedirectURI
	}
	return c.PublicBaseURL() + "/workspace-app/oauth/callback"
}
