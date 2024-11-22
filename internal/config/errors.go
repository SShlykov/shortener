package config

import "errors"

var (
	ErrConfigPathNotSpecified = errors.New("config path not specified")
	ErrConfigFileNotFound     = errors.New("config file not found")
	ErrCantReadConfigFile     = errors.New("can't read config file")
	ErrCantReadHostName       = errors.New("can't read hostname")
	ErrCantReadPort           = errors.New("can't read port")
	ErrCantReadUserName       = errors.New("can't read username")
	ErrCantReadPassword       = errors.New("can't read password")
	ErrCantReadDBName         = errors.New("can't read db name")
)
