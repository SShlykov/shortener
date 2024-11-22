package dto

import (
	"errors"

	"github.com/labstack/echo/v4"
)

var (
	ErrSecretEmpty = errors.New("secret is empty")
)

type SecretRequest struct {
	Secret string `json:"secret"`
}

func (s *SecretRequest) Validate() error {
	if s.Secret == "" {
		return ErrSecretEmpty
	}

	return nil
}

func EjectSecret(ectx echo.Context) (*SecretRequest, error) {
	var secret SecretRequest
	if err := ectx.Bind(&secret); err != nil {
		return nil, err
	}

	return &secret, nil
}
