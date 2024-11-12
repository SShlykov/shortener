package config

import "errors"

var (
	ErrConfigPathNotSpecified = errors.New("config path not specified")
	ErrConfigFileNotFound     = errors.New("config file not found")
	ErrCantReadConfigFile     = errors.New("can't read config file")
)
