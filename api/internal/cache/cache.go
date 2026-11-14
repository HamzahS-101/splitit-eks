package cache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/HamzahS-101/splitit-eks/api/internal/metrics"
	"github.com/redis/go-redis/v9"
)

const balancesTTL = 5 * time.Minute

type Cache struct {
	rdb *redis.Client
	log *slog.Logger
}

func New(addr string, log *slog.Logger) *Cache {
	if addr == "" {
		log.Warn("REDIS_ADDR not set, balance caching disabled")
		return &Cache{log: log}
	}
	return &Cache{
		log: log,
		rdb: redis.NewClient(&redis.Options{
			Addr:         addr,
			DialTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
			MaxRetries:   1,
		}),
	}
}

func (c *Cache) Close() error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

func (c *Cache) GetBalances(ctx context.Context, groupID string) ([]byte, bool) {
	if c.rdb == nil {
		return nil, false
	}
	b, err := c.rdb.Get(ctx, balancesKey(groupID)).Bytes()
	switch {
	case err == nil:
		metrics.CacheResults.WithLabelValues("hit").Inc()
		return b, true
	case errors.Is(err, redis.Nil):
		metrics.CacheResults.WithLabelValues("miss").Inc()
	default:
		metrics.CacheResults.WithLabelValues("error").Inc()
		c.log.Warn("cache read failed", "error", err)
	}
	return nil, false
}

func (c *Cache) SetBalances(ctx context.Context, groupID string, b []byte) {
	if c.rdb == nil {
		return
	}
	if err := c.rdb.Set(ctx, balancesKey(groupID), b, balancesTTL).Err(); err != nil {
		c.log.Warn("cache write failed", "error", err)
	}
}

func (c *Cache) Invalidate(ctx context.Context, groupID string) {
	if c.rdb == nil {
		return
	}
	if err := c.rdb.Del(ctx, balancesKey(groupID)).Err(); err != nil {
		c.log.Warn("cache invalidate failed", "error", err)
	}
}

func balancesKey(groupID string) string {
	return "balances:" + groupID
}
