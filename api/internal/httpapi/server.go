package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/HamzahS-101/splitit-eks/api/internal/cache"
	"github.com/HamzahS-101/splitit-eks/api/internal/store"
)

type Server struct {
	store *store.Store
	cache *cache.Cache
	log   *slog.Logger
	ready atomic.Bool
}

func New(st *store.Store, c *cache.Cache, log *slog.Logger) *Server {
	return &Server{store: st, cache: c, log: log}
}

func (s *Server) SetReady(ready bool) {
	s.ready.Store(ready)
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)

	mux.HandleFunc("POST /api/groups", s.createGroup)
	mux.HandleFunc("GET /api/groups/{slug}", s.getGroup)
	mux.HandleFunc("POST /api/groups/{slug}/members", s.addMember)
	mux.HandleFunc("GET /api/groups/{slug}/expenses", s.listExpenses)
	mux.HandleFunc("POST /api/groups/{slug}/expenses", s.createExpense)
	mux.HandleFunc("PUT /api/groups/{slug}/expenses/{id}", s.updateExpense)
	mux.HandleFunc("DELETE /api/groups/{slug}/expenses/{id}", s.deleteExpense)
	mux.HandleFunc("GET /api/groups/{slug}/balances", s.getBalances)
	mux.HandleFunc("GET /api/groups/{slug}/settlements", s.listSettlements)
	mux.HandleFunc("POST /api/groups/{slug}/settlements", s.createSettlement)
	mux.HandleFunc("DELETE /api/groups/{slug}/settlements/{id}", s.deleteSettlement)

	return s.recoverer(s.observe(mux))
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if !s.ready.Load() {
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		s.log.Warn("readiness check failed", "error", err)
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
