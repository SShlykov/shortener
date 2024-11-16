package app

import (
	"fmt"
	"io"

	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/internal/services"
	"github.com/sshlykov/shortener/pkg/logger"
)

// ServiceBuilder абстрагирует создание сервиса
type ServiceBuilder interface {
	Build(deps ServiceDependencies) (Service, error)
	Name() string
}

// ServiceDependencies содержит общие зависимости для сервисов
type ServiceDependencies struct {
	Logger  *logger.Logger
	Metrics MetricsCollector
	//Database Database
	//Cache           Cache
	//EventBus        EventBus
	//ExternalClients ExternalClients
}

type ExternalClients struct {
	//Payment PaymentGateway
	//Email   EmailProvider
	// Другие внешние клиенты
}

// BuildServices создает и инициализирует все сервисы приложения
func BuildServices(cfg config.ServicesConfig, deps ServiceDependencies) ([]Service, error) {
	builders := []ServiceBuilder{
		NewUserServiceBuilder(cfg.User),
		NewOrderServiceBuilder(cfg.Order),
		//NewPaymentServiceBuilder(cfg.Payment),
		// Добавляем другие билдеры по мере необходимости
	}

	services := make([]Service, 0, len(builders))

	for _, builder := range builders {
		svc, err := builder.Build(deps)
		if err != nil {
			// При ошибке закрываем уже созданные сервисы
			for _, s := range services {
				if closer, ok := s.(io.Closer); ok {
					_ = closer.Close()
				}
			}
			return nil, fmt.Errorf("failed to build %s service: %w", builder.Name(), err)
		}
		services = append(services, svc)
	}

	return services, nil
}

// Пример билдера сервиса
type UserServiceBuilder struct {
	cfg config.UserServiceConfig
}

func NewUserServiceBuilder(cfg config.UserServiceConfig) *UserServiceBuilder {
	return &UserServiceBuilder{cfg: cfg}
}

func (b *UserServiceBuilder) Name() string {
	return "user"
}

func (b *UserServiceBuilder) Build(deps ServiceDependencies) (Service, error) {
	if !b.cfg.Enabled {
		return nil, nil
	}

	// Создаем репозиторий для работы с данными
	repo, err := NewUserRepository(deps.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create user repository: %w", err)
	}

	// Создаем кэш-слой
	cache := NewUserCache(deps.Cache, b.cfg.CacheTTL)

	// Создаем домейн-сервис
	domainService := NewUserDomainService(
		repo,
		cache,
		deps.EventBus,
		UserDomainConfig{
			MaxRetries:        b.cfg.MaxRetries,
			ValidationTimeout: b.cfg.ValidationTimeout,
		},
	)

	// Создаем основной сервис
	svc := &services.UserService{
		BaseService: services.BaseService{
			name:    b.Name(),
			logger:  deps.Logger,
			metrics: deps.Metrics,
		},
		domain: domainService,
	}

	return svc, nil
}

// Пример билдера сервиса
type OrderServiceBuilder struct {
	cfg config.OrderServiceConfig
}

func NewOrderServiceBuilder(cfg config.OrderServiceConfig) *OrderServiceBuilder {
	return &OrderServiceBuilder{cfg: cfg}
}

func (b *OrderServiceBuilder) Name() string {
	return "order"
}

func (b *OrderServiceBuilder) Build(deps ServiceDependencies) (Service, error) {
	if !b.cfg.Enabled {
		return nil, nil
	}

	repo, err := NewOrderRepository(deps.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create order repository: %w", err)
	}

	processor := NewOrderProcessor(
		repo,
		deps.EventBus,
		deps.ExternalClients.Payment,
		OrderProcessorConfig{
			Timeout:   b.cfg.ProcessTimeout,
			BatchSize: b.cfg.BatchSize,
		},
	)

	svc := &OrderService{
		BaseService: BaseService{
			name:    b.Name(),
			logger:  deps.Logger,
			metrics: deps.Metrics,
		},
		processor: processor,
	}

	return svc, nil
}
