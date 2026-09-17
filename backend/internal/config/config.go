// Package config centralizes all environment-driven configuration in one
// validated struct, loaded once at startup in cmd/api/main.go.
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
	Env string
	Server
	Postgres
	Redis
	JWT
	Storage
	CORS
	RateLimit
	SMTP
	Translate
	FrontendURL string
}

type Server struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	LogLevel        string
}

type Postgres struct {
	DSN         string
	MaxOpenConn int
	MaxIdleConn int
	ConnMaxLife time.Duration
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

type JWT struct {
	AccessSecret   string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	Issuer         string
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite string
}

type Storage struct {
	Provider     string // "local" | "s3"
	LocalBaseDir string
	LocalPublic  string
	S3Bucket     string
	S3Region     string
	S3Endpoint   string
	S3AccessKey  string
	S3SecretKey  string
	S3PathStyle  bool
	S3PublicURL  string
}

type CORS struct {
	AllowedOrigins []string
}

type RateLimit struct {
	RequestsPerMinute     int
	AuthRequestsPerMinute int
}

// SMTP is optional: when Host is empty, the app falls back to a console mailer
// that logs the reset link instead of emailing it (fine for local dev).
type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Translate configures the admin "translate a whole book from an uploaded
// English document" feature: a Python subprocess (google-docx parsing + the
// Gemini API) turns the document into a CSV in the same shape the regular
// Excel/CSV importer already understands, so it can flow through that
// existing, battle-tested pipeline unchanged.
type Translate struct {
	GeminiAPIKey   string
	GeminiModel    string
	RequestDelay   time.Duration
	PythonBin      string
	ScriptPath     string
	CommandTimeout time.Duration
}

// Load reads a .env file if present (local dev convenience; ignored in
// containers where env vars are injected directly) and builds the Config,
// applying sane defaults so the service is runnable with a minimal .env.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env: getEnv("APP_ENV", "development"),
		Server: Server{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
		},
		Postgres: Postgres{
			DSN:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/bookreader?sslmode=disable"),
			MaxOpenConn: getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConn: getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLife: getDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		},
		Redis: Redis{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
		JWT: JWT{
			AccessSecret:   getEnv("JWT_ACCESS_SECRET", ""),
			AccessTTL:      getDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:     getDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
			Issuer:         getEnv("JWT_ISSUER", "bookreader-api"),
			CookieDomain:   getEnv("COOKIE_DOMAIN", "localhost"),
			CookieSecure:   getBool("COOKIE_SECURE", false),
			CookieSameSite: getEnv("COOKIE_SAMESITE", "Strict"),
		},
		Storage: Storage{
			Provider:     getEnv("STORAGE_PROVIDER", "local"),
			LocalBaseDir: getEnv("STORAGE_LOCAL_DIR", "./storage/uploads"),
			LocalPublic:  getEnv("STORAGE_LOCAL_PUBLIC_URL", "http://localhost:8080/uploads"),
			S3Bucket:     getEnv("S3_BUCKET", ""),
			S3Region:     getEnv("S3_REGION", "us-east-1"),
			S3Endpoint:   getEnv("S3_ENDPOINT", ""),
			S3AccessKey:  getEnv("S3_ACCESS_KEY", ""),
			S3SecretKey:  getEnv("S3_SECRET_KEY", ""),
			S3PathStyle:  getBool("S3_USE_PATH_STYLE", true),
			S3PublicURL:  getEnv("S3_PUBLIC_URL", ""),
		},
		CORS: CORS{
			AllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		},
		RateLimit: RateLimit{
			RequestsPerMinute:     getInt("RATE_LIMIT_RPM", 120),
			AuthRequestsPerMinute: getInt("RATE_LIMIT_AUTH_RPM", 10),
		},
		SMTP: SMTP{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "no-reply@bookreader.local"),
		},
		Translate: Translate{
			GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
			GeminiModel:    getEnv("GEMINI_MODEL", "gemini-3.6-flash"),
			RequestDelay:   getDuration("GEMINI_REQUEST_DELAY", 4500*time.Millisecond),
			PythonBin:      getEnv("PYTHON_BIN", "python3"),
			ScriptPath:     getEnv("TRANSLATE_SCRIPT_PATH", "scripts/translate_book.py"),
			CommandTimeout: getDuration("TRANSLATE_COMMAND_TIMEOUT", 30*time.Minute),
		},
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	if cfg.JWT.AccessSecret == "" {
		if cfg.Env == "production" {
			return nil, fmt.Errorf("JWT_ACCESS_SECRET must be set in production")
		}
		cfg.JWT.AccessSecret = "dev-only-insecure-secret-change-me"
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
