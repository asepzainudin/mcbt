package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName string
	AppEnv  string
	AppHost string
	AppPort string

	DB DBCfg

	LogLevel string
	EnvFile  string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	CookieSecure   bool
	CookieSameSite string
}

type DBCfg struct {
	Host                string
	Port                string
	User                string
	Password            string
	Name                string
	SSLMode             string
	MaxOpenConns        int
	MaxIdleConns        int
	ConnMaxLifetimeMins int
}

func (d DBCfg) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func Load() (*Config, error) {
	envPath, found := findEnvFile(".env")
	if found {
		if err := loadEnvFile(envPath); err != nil {
			return nil, err
		}
	} else if !requiredEnvPresent() {
		return nil, errors.New(
			"no .env file found and required environment variables are not set: " +
				"place .env in the project root or run from within the project directory",
		)
	}

	cfg := &Config{
		AppName:        getEnv("APP_NAME", "mcbt"),
		AppEnv:         getEnv("APP_ENV", "development"),
		AppHost:        getEnv("APP_HOST", "localhost"),
		AppPort:        getEnv("APP_PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		EnvFile:        envPath,
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTAccessTTL:   getEnvDuration("JWT_ACCESS_TTL_MINUTES", 15, time.Minute),
		JWTRefreshTTL:  getEnvDuration("JWT_REFRESH_TTL_DAYS", 7, 24*time.Hour),
		CookieSecure:   getEnvBool("COOKIE_SECURE", false),
		CookieSameSite: getEnv("COOKIE_SAMESITE", "strict"),
		DB: DBCfg{
			Host:                getEnv("DB_HOST", "localhost"),
			Port:                getEnv("DB_PORT", "5432"),
			User:                getEnv("DB_USER", "postgres"),
			Password:            getEnv("DB_PASSWORD", ""),
			Name:                getEnv("DB_NAME", "mcbt"),
			SSLMode:             getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:        getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:        getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetimeMins: getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5),
		},
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.JWTSecret) == "" {
		return errors.New("JWT_SECRET is required: generate one with `openssl rand -hex 32`")
	}
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}
	if c.IsProduction() && !c.CookieSecure {
		return errors.New("COOKIE_SECURE must be true when APP_ENV=production")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func (c *Config) ConnMaxLifetime() time.Duration {
	return time.Duration(c.DB.ConnMaxLifetimeMins) * time.Minute
}

func findEnvFile(name string) (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for {
		candidate := filepath.Join(dir, name)
		if fileExists(candidate) {
			return candidate, true
		}

		if fileExists(filepath.Join(dir, "go.mod")) {
			return "", false
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func requiredEnvPresent() bool {
	required := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return false
		}
	}
	return true
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}
	return sc.Err()
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func getEnvDuration(key string, fallback int64, unit time.Duration) time.Duration {
	n := int64(getEnvInt(key, int(fallback)))
	if n <= 0 {
		n = fallback
	}
	return time.Duration(n) * unit
}
