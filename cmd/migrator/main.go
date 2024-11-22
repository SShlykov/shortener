package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"

	"github.com/sshlykov/shortener/internal/config"
	_ "github.com/sshlykov/shortener/migrations"
	"github.com/sshlykov/shortener/pkg/backoff"
)

const (
	timeout = 30 * time.Second
)

var (
	ErrCantStartApplication = errors.New("application not started")
	ErrTimeoutExceeded      = errors.New("error time exceeded")
	ErrCantPingDatabase     = errors.New("database not available")
)

func main() {
	data, err := os.ReadFile("./MIGRATION")
	if err != nil {
		log.Fatalf("Error reading MIGRATION file: %v\n", err)
	}
	if strings.TrimSpace(string(data)) != "true" {
		log.Printf("MIGRATION file does not contain true; content: %s\n", string(data))
		return
	}
	dbDSN, err := config.GetDSN()
	if err != nil {
		log.Fatalf("Error getting database DSN: %v\n", err)
	}

	fmt.Println(dbDSN)

	var db *sql.DB
	startTime := time.Now()
	h := func() error {
		if time.Since(startTime) > timeout {
			return backoff.Permanent(ErrTimeoutExceeded)
		}
		db, err = sql.Open("postgres", dbDSN)
		if err != nil {
			fmt.Println(err)
			return ErrCantStartApplication
		}
		if perr := db.Ping(); perr != nil {
			return ErrCantPingDatabase
		}
		return backoff.Permanent(nil)
	}
	if err = backoff.Retry(h, backoff.NewExponentialBackOff()); err != nil {
		log.Fatalf("failed to open db connection: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close db connection: %v\n", err)
		}
	}()
	if len(os.Args) < 2 {
		log.Fatalf("usage: use goose command like 'go run ... command' \n")
	}

	cmd := os.Args[1]
	if err = goose.RunContext(context.Background(), cmd, db, "./migrations"); err != nil {
		log.Fatalf("failed to run goose command: %v\n", err)
	}
}
