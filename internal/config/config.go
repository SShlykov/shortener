package config

import (
	"os"
	"time"
)

type Config struct {
	App    App    `yaml:"app"`
	Health Health `yaml:"health"`
	Logger Logger `yaml:"logger"`
	DB     DB     `yaml:"db"`
}

type App struct {
	Name                 string        `yaml:"name"`
	Version              string        `yaml:"version"`
	Env                  string        `yaml:"env"`
	TerminateTimeout     time.Duration `yaml:"terminate_timeout"`
	OtelAgent            string        `yaml:"otel_agent"`
	ReadinessCheckPeriod time.Duration `yaml:"readiness_check_period"`
}

type DB struct {
	RefreshTimeout time.Duration `yaml:"refresh_timeout"`
}

type Health struct {
	Port int `yaml:"port"`

	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`

	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type Logger struct {
	Level string `yaml:"level"`
	Mode  string `yaml:"mode"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, ErrConfigPathNotSpecified
	}

	path += "/default.yaml"
	if _, err := os.Stat(path); err != nil {
		return nil, ErrConfigFileNotFound
	}

	cfg := &Config{}
	if err := ReadConfig(path, cfg); err != nil {
		return nil, ErrCantReadConfigFile
	}

	return cfg, nil
}

// --------------------------

type AppConfig struct {
	Logger   LoggerConfig   `yaml:"logger"`
	Health   HealthConfig   `yaml:"health"`
	Services ServicesConfig `yaml:"services"`
	//Database DatabaseConfig `yaml:"database"`
	//Cache    CacheConfig    `yaml:"cache"`
	//EventBus EventBusConfig `yaml:"event_bus"`

	// Конфигурация внешних сервисов
	//Payment PaymentConfig `yaml:"payment"`
	//Email   EmailConfig   `yaml:"email"`
}

type ServicesConfig struct {
	User  UserServiceConfig  `yaml:"user"`
	Order OrderServiceConfig `yaml:"order"`
	//Payment PaymentServiceConfig `yaml:"payment"`
}

type UserServiceConfig struct {
	BaseServiceConfig
	CacheTTL          time.Duration `yaml:"cache_ttl"`
	MaxRetries        int           `yaml:"max_retries"`
	ValidationTimeout time.Duration `yaml:"validation_timeout"`
}

type OrderServiceConfig struct {
	BaseServiceConfig
	ProcessTimeout time.Duration `yaml:"process_timeout"`
	BatchSize      int           `yaml:"batch_size"`
}

// общие настройки, используемые всеми сервисами
type BaseServiceConfig struct {
	Enabled  bool   `yaml:"enabled"`
	LogLevel string `yaml:"log_level"`
}

type LoggerConfig struct {
	Level int `yaml:"level"`
	Mode  int `yaml:"mode"`
}

type HealthConfig struct {
	Port            int           `yaml:"port"`
	Host            string        `yaml:"host"`
	CheckInterval   time.Duration `yaml:"check_interval"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}
