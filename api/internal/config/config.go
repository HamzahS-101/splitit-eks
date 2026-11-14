package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	MetricsPort    string
	DBHost         string
	DBPort         string
	DBUser         string
	DBName         string
	DBSSLMode      string
	DBPasswordFile string
	RedisAddr      string
	CleanupDays    int
	ShutdownDelay  time.Duration
}

func Load() (Config, error) {
	c := Config{
		Port:           env("PORT", "8080"),
		MetricsPort:    env("METRICS_PORT", "9090"),
		DBHost:         env("DB_HOST", "localhost"),
		DBPort:         env("DB_PORT", "5432"),
		DBUser:         env("DB_USER", "splitit"),
		DBName:         env("DB_NAME", "splitit"),
		DBSSLMode:      env("DB_SSLMODE", "disable"),
		DBPasswordFile: os.Getenv("DB_PASSWORD_FILE"),
		RedisAddr:      os.Getenv("REDIS_ADDR"),
	}

	var err error
	if c.CleanupDays, err = envInt("CLEANUP_DAYS", 30); err != nil {
		return c, err
	}
	if c.ShutdownDelay, err = time.ParseDuration(env("SHUTDOWN_DELAY", "5s")); err != nil {
		return c, fmt.Errorf("SHUTDOWN_DELAY: %w", err)
	}
	if c.DBPasswordFile == "" {
		return c, errors.New("DB_PASSWORD_FILE is required")
	}
	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return n, nil
}
