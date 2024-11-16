package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sshlykov/shortener/internal/bootstrap/app"
	"github.com/sshlykov/shortener/internal/config"
)

const (
	OkCode = iota
	ErrConfigLoad
	ErrCreateApp
	ErrRunApp
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config", "path to configuration file")

	ctx := context.Background()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("failed to load config: %s\n", err.Error())
		os.Exit(ErrConfigLoad)
	}

	application, err := app.New(cfg)
	if err != nil {
		fmt.Printf("failed to create app: %s\n", err.Error())
		os.Exit(ErrCreateApp)
	}

	if err = application.Run(ctx); err != nil {
		fmt.Printf("failed to run app: %s\n", err.Error())
		os.Exit(ErrRunApp)
	}

	os.Exit(OkCode)
}
