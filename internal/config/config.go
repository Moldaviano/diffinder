package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	JWT      JWTConfig
	Webhook  WebhookConfig
	CORS     CORSConfig
	Log      LogConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	MaxConns int32
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode, d.MaxConns,
	)
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type WebhookConfig struct {
	GitHubSecret string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type LogConfig struct {
	Level  string
	Format string
}

// Load legge la configurazione da variabili d'ambiente.
// Se presente .env nella working dir, viene caricato (best-effort).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:         getenv("SERVER_HOST", "0.0.0.0"),
			Port:         getenv("SERVER_PORT", "8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
		},
		DB: DBConfig{
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenv("DB_PORT", "5432"),
			Name:     getenv("DB_NAME", "diffinder"),
			User:     getenv("DB_USER", "diffinder"),
			Password: getenv("DB_PASSWORD", "diffinder"),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
			MaxConns: int32(getInt("DB_MAX_CONNS", 10)),
		},
		JWT: JWTConfig{
			Secret:     mustEnv("JWT_SECRET"),
			AccessTTL:  getDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: getDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
		Webhook: WebhookConfig{
			GitHubSecret: mustEnv("GITHUB_WEBHOOK_SECRET"),
		},
		CORS: CORSConfig{
			AllowedOrigins: splitCSV(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:4200")),
		},
		Log: LogConfig{
			Level:  getenv("LOG_LEVEL", "info"),
			Format: getenv("LOG_FORMAT", "json"),
		},
	}

	if len(cfg.JWT.Secret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 chars")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return v
}

func getDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
