package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HamzahS-101/splitit-eks/api/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBSSLMode)

	pc, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	pc.BeforeConnect = func(_ context.Context, cc *pgx.ConnConfig) error {
		password, err := readPassword(cfg.DBPasswordFile)
		if err != nil {
			return err
		}
		cc.Password = password
		return nil
	}
	pc.MaxConns = 10
	pc.MaxConnLifetime = 30 * time.Minute

	return pgxpool.NewWithConfig(ctx, pc)
}

func readPassword(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read db password file: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}
