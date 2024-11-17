package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Core interfaces
type (
	Checker interface {
		Check(ctx context.Context) error
	}

	Runner interface {
		Run(ctx context.Context) error
		Stop()
	}

	// Component базовый интерфейс всех компонентов
	Component interface {
		Name() string
		Init(ctx context.Context) error // инициализация
		Checker                         // проверка состояния
	}

	// ServiceComponent расширяет базовый компонент для бизнес-сервисов
	ServiceComponent interface {
		Component // инициализация и проверка состояния
		Runner    // запуск и остановка
	}

	// ComponentBuilder определяет интерфейс для построения компонентов
	ComponentBuilder interface {
		Build(ctx context.Context, deps *Dependencies) (Component, error)
	}

	// Dependencies содержит все зависимости для компонентов
	Dependencies struct {
		Config  *Config
		Logger  *zap.Logger
		Metrics MetricsClient
		DB      Database
	}
)

// Factory для создания компонентов
type ComponentFactory struct {
	builders map[string]ComponentBuilder
	mu       sync.RWMutex
}

func NewComponentFactory() *ComponentFactory {
	return &ComponentFactory{
		builders: make(map[string]ComponentBuilder),
	}
}

func (f *ComponentFactory) Register(name string, builder ComponentBuilder) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.builders[name] = builder
}

func (f *ComponentFactory) Create(name string, ctx context.Context, deps *Dependencies) (Component, error) {
	f.mu.RLock()
	builder, exists := f.builders[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no builder registered for component: %s", name)
	}
	return builder.Build(ctx, deps)
}

// Abstract Factory для создания связанных компонентов
type ServiceFactory interface {
	CreateAPI() (ServiceComponent, error)
	CreateProcessor() (ServiceComponent, error)
	CreateRepository() (Component, error)
}

// Concrete Factories
type ImageServiceFactory struct {
	deps *Dependencies
}

func NewImageServiceFactory(deps *Dependencies) *ImageServiceFactory {
	return &ImageServiceFactory{deps: deps}
}

// Builder implementations
type (
	MetricsBuilder  struct{}
	LoggerBuilder   struct{}
	DatabaseBuilder struct{}
	APIBuilder      struct{}
)

func (b *MetricsBuilder) Build(ctx context.Context, deps *Dependencies) (Component, error) {
	return NewMetricsComponent(deps.Config)
}

// Base Component implementation
type BaseComponent struct {
	name   string
	logger *zap.Logger
	status Status
	mu     sync.RWMutex
}

func NewBaseComponent(name string, logger *zap.Logger) BaseComponent {
	return BaseComponent{
		name:   name,
		logger: logger,
		status: Status{Healthy: true},
	}
}

func (c *BaseComponent) Name() string {
	return c.name
}

func (c *BaseComponent) Check(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.status.Healthy {
		return fmt.Errorf("component %s is unhealthy: %v", c.name, c.status.Error)
	}
	return nil
}

// Application composition using DI container
// Container отвечает за:
// 1. Создание компонентов (DI)
// 2. Управление зависимостями
// 3. Конфигурацию компонентов
// Работает на этапе инициализации приложения
type Container struct {
	components map[string]Component
	services   map[string]ServiceComponent
	factory    *ComponentFactory
	deps       *Dependencies
	mu         sync.RWMutex
}

func NewContainer(deps *Dependencies) *Container {
	return &Container{
		components: make(map[string]Component),
		services:   make(map[string]ServiceComponent),
		factory:    NewComponentFactory(),
		deps:       deps,
	}
}

// Get возвращает компонент по имени
func (c *Container) Get(name string) (Component, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if comp, exists := c.components[name]; exists {
		return comp, nil
	}
	return nil, fmt.Errorf("component %s not found", name)
}

// Resolve разрешает зависимости для компонента
func (c *Container) Resolve(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	builder, exists := c.factory.builders[name]
	if !exists {
		return fmt.Errorf("no builder for component %s", name)
	}

	// Создание компонента с его зависимостями
	component, err := builder.Build(context.Background(), c.deps)
	if err != nil {
		return fmt.Errorf("failed to build component %s: %w", name, err)
	}

	c.components[name] = component
	return nil
}

func (c *Container) RegisterComponent(name string, builder ComponentBuilder) {
	c.factory.Register(name, builder)
}

func (c *Container) Build(ctx context.Context) error {
	// Build core components first
	required := []string{"metrics", "logger", "database"}
	for _, name := range required {
		comp, err := c.factory.Create(name, ctx, c.deps)
		if err != nil {
			return fmt.Errorf("failed to build required component %s: %w", name, err)
		}
		c.components[name] = comp
	}

	return nil
}

// Improved App structure
type App struct {
	container *Container
	registry  *Registry
	logger    *zap.Logger
}

// Registry отвечает за:
// 1. Управление жизненным циклом компонентов
// 2. Запуск/остановку сервисов
// 3. Мониторинг состояния
// Работает во время выполнения приложения
type Registry struct {
	runners    []Runner    // Компоненты, требующие запуска
	checkers   []Checker   // Компоненты с проверкой состояния
	components []Component // Все компоненты
	startOrder []string    // Порядок запуска
	services   []ServiceComponent
	logger     *zap.Logger
	mu         sync.RWMutex
}

func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		logger:     logger,
		startOrder: make([]string, 0),
	}
}

// Add регистрирует компонент и определяет его возможности
func (r *Registry) Add(component Component) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.components = append(r.components, component)

	// Проверяем дополнительные интерфейсы
	if runner, ok := component.(Runner); ok {
		r.runners = append(r.runners, runner)
	}
	if checker, ok := component.(Checker); ok {
		r.checkers = append(r.checkers, checker)
	}

	r.startOrder = append(r.startOrder, component.Name())
}

// Start запускает все компоненты в правильном порядке
func (r *Registry) Start(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range r.startOrder {
		r.logger.Info("starting component", zap.String("name", name))

		for _, runner := range r.runners {
			if runner.(Component).Name() == name {
				if err := runner.Run(ctx); err != nil {
					return fmt.Errorf("failed to start %s: %w", name, err)
				}
			}
		}
	}
	return nil
}

// CheckHealth проверяет состояние всех компонентов
func (r *Registry) CheckHealth(ctx context.Context) map[string]Status {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]Status)
	for _, checker := range r.checkers {
		name := checker.(Component).Name()
		err := checker.Check(ctx)
		results[name] = Status{
			Healthy:   err == nil,
			LastCheck: time.Now(),
			Error:     err,
		}
	}
	return results
}

func (r *Registry) Register(component Component) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if service, ok := component.(ServiceComponent); ok {
		r.services = append(r.services, service)
	}
	r.components = append(r.components, component)
}

// Example service implementation
type ImageService struct {
	BaseComponent
	processor ImageProcessor
	repo      ImageRepository
	api       *ImageAPI
}

// ImageServiceBuilder implements ComponentBuilder
type ImageServiceBuilder struct {
	processorFactory ServiceFactory
}

func (b *ImageServiceBuilder) Build(ctx context.Context, deps *Dependencies) (Component, error) {
	factory := NewImageServiceFactory(deps)

	processor, err := factory.CreateProcessor()
	if err != nil {
		return nil, fmt.Errorf("failed to create image processor: %w", err)
	}

	repo, err := factory.CreateRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create image repository: %w", err)
	}

	api, err := factory.CreateAPI()
	if err != nil {
		return nil, fmt.Errorf("failed to create image API: %w", err)
	}

	return &ImageService{
		BaseComponent: NewBaseComponent("image-service", deps.Logger),
		processor:     processor.(ImageProcessor),
		repo:          repo.(ImageRepository),
		api:           api.(*ImageAPI),
	}, nil
}

// Configuration using functional options pattern
type Option func(*App) error

func WithComponent(name string, builder ComponentBuilder) Option {
	return func(app *App) error {
		app.container.RegisterComponent(name, builder)
		return nil
	}
}

func New(ctx context.Context, deps *Dependencies, opts ...Option) (*App, error) {
	container := NewContainer(deps)
	registry := NewRegistry(deps.Logger)

	app := &App{
		container: container,
		registry:  registry,
		logger:    deps.Logger,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(app); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Создаем компоненты через Container
	required := []string{"database", "cache", "metrics"}
	for _, name := range required {
		if err := container.Resolve(name); err != nil {
			return nil, err
		}

		// Получаем созданный компонент
		component, err := container.Get(name)
		if err != nil {
			return nil, err
		}

		// Регистрируем его в Registry
		registry.Add(component)
	}

	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	// Container уже создал все компоненты
	// Registry запускает их
	if err := a.registry.Start(ctx); err != nil {
		return err
	}

	// Периодическая проверка здоровья
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				status := a.registry.CheckHealth(ctx)
				// Обработка результатов проверки
			}
		}
	}()

	return nil
}

// Example usage
func ExampleUsage() {
	ctx := context.Background()
	deps := &Dependencies{
		Config:  loadConfig(),
		Logger:  initLogger(),
		Metrics: initMetrics(),
		DB:      initDatabase(),
	}

	app, err := New(ctx, deps,
		WithComponent("image-service", &ImageServiceBuilder{}),
		WithComponent("user-service", &UserServiceBuilder{}),
	)
	if err != nil {
		panic(err)
	}

	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}

type ServiceDiscovery interface {
	Register(service ServiceComponent) error
	Unregister(service ServiceComponent) error
	Discover(name string) (ServiceComponent, error)
}

type Middleware func(Component) Component

func WithMetrics(metrics MetricsClient) Middleware {
	return func(c Component) Component {
		// Add metrics wrapper
		return &MetricsWrapper{component: c, metrics: metrics}
	}
}
