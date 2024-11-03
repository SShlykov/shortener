package config

import (
	"bufio"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type IConfig interface {
	Default()
	Validate() error
}

type Config struct {
	App App `yaml:"app"`
}

type App struct {
	Name string `yaml:"name"`
}

func Load(path string) (*Config, error) {

	return &Config{}, nil
}

func load(config *Config) error {
	var (
		err  error
		name string
		file *os.File
	)

	if len(os.Args) < 2 || os.Args[1] == "" {
		name = defaultConfigFile
	} else {
		name = os.Args[1]
	}

	if file, err = os.Open(name); err != nil {
		return err
	}

	config.Default()

	dec := yaml.NewDecoder(bufio.NewReader(file))
	dec.KnownFields(true)
	if err = dec.Decode(config); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (c *Config) Default() {
	c.App.Name = "shortener"
}

func (c *Config) Validate() error {
	return nil
}
