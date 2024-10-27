package config

type Config struct {
	App App `yaml:"app"`
}

type App struct {
	Name string `yaml:"name"`
}

func Load(path string) (*Config, error) {

	return &Config{}, nil
}
