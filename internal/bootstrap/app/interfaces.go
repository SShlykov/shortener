package app

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sshlykov/shortener/internal/config"
)

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	// другие необходимые методы логгера
}

type MetricsCollector interface {
	RecordHealthCheck(name string, success bool)
	//IncRequestCounter(method string)
	//ObserveRequestDuration(method string, duration float64)
}

type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

type HealthCheck struct {
	checkers []HealthChecker
	server   *http.Server
	metrics  MetricsCollector
	status   atomic.Value // ServiceStatus
	mu       sync.RWMutex
	cfg      config.HealthConfig
}

type ServiceStatus struct {
	Status    string                   `json:"status"`
	Details   map[string]CheckerStatus `json:"details"`
	Timestamp time.Time                `json:"timestamp"`
}

type CheckerStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
