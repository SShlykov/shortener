package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sshlykov/shortener/internal/config"
)

/*
	type DatabaseChecker struct {
		db Database
	}

	func NewDatabaseChecker(db Database) *DatabaseChecker {
		return &DatabaseChecker{db: db}
	}

	func (c *DatabaseChecker) Check(ctx context.Context) error {
		return c.db.Ping(ctx)
	}

	func (c *DatabaseChecker) Name() string {
		return "database"
	}

	type CacheChecker struct {
		cache Cache
	}

	func NewCacheChecker(cache Cache) *CacheChecker {
		return &CacheChecker{cache: cache}
	}

	func (c *CacheChecker) Check(ctx context.Context) error {
		return c.cache.Ping(ctx)
	}

	func (c *CacheChecker) Name() string {
		return "cache"
	}
*/
func NewHealthCheck(cfg config.HealthConfig, metrics MetricsCollector) (*HealthCheck, error) {
	if cfg.Port == 0 {
		return nil, fmt.Errorf("health check port is required")
	}
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 30 * time.Second
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 5 * time.Second
	}

	h := &HealthCheck{
		cfg:     cfg,
		metrics: metrics,
	}
	h.status.Store(ServiceStatus{
		Status:    "starting",
		Details:   make(map[string]CheckerStatus),
		Timestamp: time.Now(),
	})

	// Настраиваем HTTP сервер
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/ready", h.handleReady)

	h.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: mux,
	}

	return h, nil
}

func (h *HealthCheck) Start(ctx context.Context) error {
	// Запускаем периодическую проверку в фоне
	go h.runChecks(ctx)

	// Запускаем HTTP сервер
	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("health check server error: %w", err)
	}

	return nil
}

func (h *HealthCheck) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, h.cfg.ShutdownTimeout)
	defer cancel()

	return h.server.Shutdown(ctx)
}

func (h *HealthCheck) AddChecker(checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

func (h *HealthCheck) runChecks(ctx context.Context) {
	ticker := time.NewTicker(h.cfg.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.performChecks(ctx)
		}
	}
}

func (h *HealthCheck) performChecks(ctx context.Context) {
	h.mu.RLock()
	checkers := h.checkers
	h.mu.RUnlock()

	status := ServiceStatus{
		Status:    "healthy",
		Details:   make(map[string]CheckerStatus),
		Timestamp: time.Now(),
	}

	for _, checker := range checkers {
		checkerStatus := CheckerStatus{Status: "healthy"}

		if err := checker.Check(ctx); err != nil {
			checkerStatus.Status = "unhealthy"
			checkerStatus.Error = err.Error()
			status.Status = "unhealthy"
		}

		status.Details[checker.Name()] = checkerStatus
		h.metrics.RecordHealthCheck(checker.Name(), checkerStatus.Status == "healthy")
	}

	h.status.Store(status)
}

func (h *HealthCheck) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := h.status.Load().(ServiceStatus)

	w.Header().Set("Content-Type", "application/json")
	if status.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

func (h *HealthCheck) handleReady(w http.ResponseWriter, r *http.Request) {
	status := h.status.Load().(ServiceStatus)

	w.Header().Set("Content-Type", "application/json")
	if status.Status == "starting" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}
