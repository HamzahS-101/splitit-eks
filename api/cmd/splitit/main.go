package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HamzahS-101/splitit-eks/api/internal/cache"
	"github.com/HamzahS-101/splitit-eks/api/internal/config"
	"github.com/HamzahS-101/splitit-eks/api/internal/db"
	"github.com/HamzahS-101/splitit-eks/api/internal/httpapi"
	"github.com/HamzahS-101/splitit-eks/api/internal/store"
	"github.com/HamzahS-101/splitit-eks/api/migrations"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: splitit <serve|migrate|cleanup>")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch os.Args[1] {
	case "serve":
		err = serve(ctx, cfg, log)
	case "migrate":
		err = migrate(ctx, cfg, log)
	case "cleanup":
		err = cleanup(ctx, cfg, log)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(2)
	}

	if err != nil {
		log.Error(os.Args[1]+" failed", "error", err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	c := cache.New(cfg.RedisAddr, log)
	defer c.Close()

	srv := httpapi.New(store.New(pool), c, log)

	api := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	metricsMux := http.NewServeMux()
	metricsMux.Handle("GET /metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:              ":" + cfg.MetricsPort,
		Handler:           metricsMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 2)
	for _, s := range []*http.Server{api, metricsSrv} {
		go func() {
			if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}
		}()
	}

	srv.SetReady(true)
	log.Info("listening", "port", cfg.Port, "metrics_port", cfg.MetricsPort)

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	log.Info("shutdown started", "delay", cfg.ShutdownDelay.String())
	srv.SetReady(false)
	time.Sleep(cfg.ShutdownDelay)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := api.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Info("shutdown complete")
	return nil
}

func migrate(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.UpContext(ctx, sqlDB, "."); err != nil {
		return err
	}

	version, err := goose.GetDBVersionContext(ctx, sqlDB)
	if err != nil {
		return err
	}
	log.Info("migrations applied", "version", version)
	return nil
}

func cleanup(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	deleted, err := store.New(pool).DeleteInactiveGroups(ctx, cfg.CleanupDays)
	if err != nil {
		return err
	}
	log.Info("cleanup complete", "groups_deleted", deleted, "inactive_days", cfg.CleanupDays)
	return nil
}
