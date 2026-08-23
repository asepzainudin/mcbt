package config

import (
	"bufio"
	"fmt"
	"os"
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
	if err := loadEnvFile(".env"); err != nil {
		return nil, err
	}

	cfg := &Config{
		AppName:  getEnv("APP_NAME", "mcbt"),
		AppEnv:   getEnv("APP_ENV", "development"),
		AppHost:  getEnv("APP_HOST", "localhost"),
		AppPort:  getEnv("APP_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
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
	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func (c *Config) ConnMaxLifetime() time.Duration {
	return time.Duration(c.DB.ConnMaxLifetimeMins) * time.Minute
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
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
