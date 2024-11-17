package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sshlykov/shortener/pkg/logger"
)

// Config represents application configuration
type Config struct {
	HTTPPort    int
	MetricsPort int
	DatabaseURL string
	BrokerURL   string
}

// App represents the main application structure
type App struct {
	cfg      *Config
	logger   *logger.Logger
	db       *sql.DB
	registry []Runner
	metrics  *Metrics
	broker   *Broker
	services *Services
	wg       sync.WaitGroup
}

// Metrics represents application metrics
type Metrics struct {
	httpRequestsTotal *prometheus.CounterVec
	server            *http.Server
}

// Services contains all business services
type Services struct {
	cropper *CropperService
	zoomer  *ZoomerService
	rotator *RotatorService
	user    *UserService
}

// Broker handles message queue operations
type Broker struct {
	producer Producer
	consumer Consumer
}

// New creates and initializes a new application instance
func New(cfg *Config) (*App, error) {
	// Initialize logger
	log, err := logger.Setup(logger.LevelInfo, logger.ModePretty)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize metrics
	metrics := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint"},
		),
	}
	prometheus.MustRegister(metrics.httpRequestsTotal)

	// Initialize database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize broker
	broker := &Broker{
		producer: NewProducer(cfg.BrokerURL),
		consumer: NewConsumer(cfg.BrokerURL),
	}

	// Initialize services
	services := &Services{
		cropper: NewCropperService(),
		zoomer:  NewZoomerService(),
		rotator: NewRotatorService(),
		user:    NewUserService(db),
	}

	// Create runners
	runners := []Runner{
		NewBusinessAPI(cfg, services, log, metrics),
		NewMetricsServer(metrics, cfg.MetricsPort),
		NewBrokerConsumer(broker.consumer),
		NewReadinessServer(cfg.HTTPPort + 1),
		NewHealthServer(cfg.HTTPPort + 2),
	}

	return &App{
		cfg:      cfg,
		logger:   log,
		db:       db,
		metrics:  metrics,
		broker:   broker,
		services: services,
		registry: runners,
	}, nil
}

// Run starts all application components
func (a *App) Run(ctx context.Context) error {
	// Start all runners
	for _, runner := range a.registry {
		a.wg.Add(1)
		go func(r Runner) {
			defer a.wg.Done()
			if err := r.Run(ctx); err != nil {
				a.logger.Error("runner failed")
			}
		}(runner)
	}

	// Setup graceful shutdown
	<-ctx.Done()
	a.Stop()
	a.wg.Wait()

	return nil
}

// Stop stops all application components
func (a *App) Stop() {
	for _, runner := range a.registry {
		runner.Stop()
	}
	a.db.Close()
	a.logger.Sync()
}

// Checker interface implementation example
type DBHealthChecker struct {
	db      *sql.DB
	metrics *Metrics
}

func (h *DBHealthChecker) Check(ctx context.Context) error {
	// Check database connection
	if err := h.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database check failed: %w", err)
	}
	return nil
}

// Runner interface implementation example
type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer(metrics *Metrics, port int) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return &MetricsServer{server: server}
}

func (m *MetricsServer) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		m.Stop()
	}()

	if err := m.server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("metrics server failed: %w", err)
	}
	return nil
}

func (m *MetricsServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m.server.Shutdown(ctx)
}

// Builder interface implementation example
type MetricsServerBuilder struct{}

func (b *MetricsServerBuilder) Build(cfg *Config) Runner {
	metrics := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint"},
		),
	}
	return NewMetricsServer(metrics, cfg.MetricsPort)
}

// BusinessAPI handles HTTP API requests
type BusinessAPI struct {
	logger   *logger.Logger
	services *Services
	server   *http.Server
	metrics  *Metrics
}

// Image represents the image data and operations
type Image struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Rotation int    `json:"rotation"`
}

// ImageProcessor provides image processing operations
type ImageProcessor interface {
	Crop(ctx context.Context, img *Image, width, height int) error
	Zoom(ctx context.Context, img *Image, factor float64) error
	Rotate(ctx context.Context, img *Image, degrees int) error
}

// Each service implements both business logic and health checking
type CropperService struct {
	status Status
	mu     sync.RWMutex
}

type Status struct {
	Healthy   bool
	LastCheck time.Time
	Error     error
}

func NewCropperService() *CropperService {
	return &CropperService{
		status: Status{Healthy: true},
	}
}

func (s *CropperService) Crop(ctx context.Context, img *Image, width, height int) error {
	if !s.status.Healthy {
		return fmt.Errorf("service is unhealthy")
	}
	// Implement cropping logic
	return nil
}

func (s *CropperService) performHealthCheck(ctx context.Context) error {
	// Check dependencies, resource availability, etc.
	return nil
}

// Checker interface implementation
func (s *CropperService) Check(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Perform health check
	err := s.performHealthCheck(ctx)
	s.status = Status{
		Healthy:   err == nil,
		LastCheck: time.Now(),
		Error:     err,
	}
	return err
}

func NewBusinessAPI(cfg *Config, services *Services, log logger.Logger, metrics *Metrics) *BusinessAPI {
	api := &BusinessAPI{
		logger:   log,
		services: services,
		metrics:  metrics,
	}

	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/v1/images/{id}/crop", api.CropImage).Methods("POST")
	router.HandleFunc("/api/v1/images/{id}/zoom", api.ZoomImage).Methods("POST")
	router.HandleFunc("/api/v1/images/{id}/rotate", api.RotateImage).Methods("POST")

	// Health check routes
	router.HandleFunc("/health", api.HealthCheck).Methods("GET")
	router.HandleFunc("/ready", api.ReadinessCheck).Methods("GET")

	api.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}

	return api
}

func (api *BusinessAPI) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		api.Stop()
	}()

	api.logger.Info("starting business API server", zap.String("addr", api.server.Addr))
	if err := api.server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("business API server failed: %w", err)
	}
	return nil
}

func (api *BusinessAPI) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	api.server.Shutdown(ctx)
}

type HealthRegistry struct {
	checkers []Checker
	logger   *zap.Logger
}

func NewHealthRegistry(logger *zap.Logger) *HealthRegistry {
	return &HealthRegistry{
		logger: logger,
	}
}

func (r *HealthRegistry) RegisterChecker(c Checker) {
	r.checkers = append(r.checkers, c)
}

func (r *HealthRegistry) CheckAll(ctx context.Context) map[string]Status {
	results := make(map[string]Status)
	for _, checker := range r.checkers {
		name := fmt.Sprintf("%T", checker)
		err := checker.Check(ctx)
		results[name] = Status{
			Healthy:   err == nil,
			LastCheck: time.Now(),
			Error:     err,
		}
	}
	return results
}

func (api *BusinessAPI) CropImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Record metric
	api.metrics.httpRequestsTotal.WithLabelValues("POST", "/api/v1/images/crop").Inc()

	img := &Image{ID: id}
	if err := api.services.cropper.Crop(r.Context(), img, req.Width, req.Height); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(img)
}

func (api *BusinessAPI) HealthCheck(w http.ResponseWriter, r *http.Request) {
	registry := NewHealthRegistry(api.logger)

	// Register all components that implement Checker interface
	registry.RegisterChecker(api.services.cropper)
	registry.RegisterChecker(api.services.zoomer)
	registry.RegisterChecker(api.services.rotator)

	// Perform health checks
	results := registry.CheckAll(r.Context())

	// If any service is unhealthy, return 503
	for _, status := range results {
		if !status.Healthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(results)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}
