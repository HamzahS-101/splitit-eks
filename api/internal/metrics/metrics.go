package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "splitit_http_requests_total",
		Help: "HTTP requests by route and status code.",
	}, []string{"route", "status"})

	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "splitit_http_request_duration_seconds",
		Help:    "HTTP request latency by route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route"})

	CacheResults = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "splitit_balance_cache_total",
		Help: "Balance cache lookups by result (hit, miss, error).",
	}, []string{"result"})
)
