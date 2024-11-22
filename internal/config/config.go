package config

import (
	"os"
	"time"
)

type Config struct {
	App    App    `yaml:"app"`
	Health Health `yaml:"health"`
	Web    Web    `yaml:"web"`
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

type Web struct {
	Port int `yaml:"port"`

	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`

	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
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
