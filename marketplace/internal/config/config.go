package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddress string
	DatabaseURL string
}

func Load() (Config, error) {
	return LoadFrom(os.Getenv)
}

func LoadFrom(getenv func(string) string) (Config, error) {
	databaseURL := getenv("MARKETPLACE_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("MARKETPLACE_DATABASE_URL is required")
	}
	address := getenv("MARKETPLACE_HTTP_ADDRESS")
	if address == "" {
		address = ":8080"
	}
	return Config{
		HTTPAddress: address,
		DatabaseURL: databaseURL,
	}, nil
}
