package config

import (
	"errors"
	"fmt"
	"os"
)

func GetDSN() (dsn string, err error) {
	dsn = os.Getenv("DB_DSN")
	if dsn != "" {
		return dsn, nil
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		err = errors.Join(err, ErrCantReadHostName)
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		err = errors.Join(err, ErrCantReadPort)
	}

	user := os.Getenv("DB_USERNAME")
	if user == "" {
		err = errors.Join(err, ErrCantReadUserName)
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		err = errors.Join(err, ErrCantReadPassword)
	}

	dbName := os.Getenv("DB_DATABASE")
	if dbName == "" {
		err = errors.Join(err, ErrCantReadDBName)
	}

	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	if err != nil {
		return "", err
	}

	dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)
	return dsn, nil
}
