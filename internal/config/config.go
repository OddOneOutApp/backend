package config

import (
	"os"
	"strings"

	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/joho/godotenv"
)

type Config struct {
	Host   string `env:"HOST" envDefault:"localhost"`
	Secure bool
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		utils.Logger.Errorf("Error loading .env file: %v", err)
	}

	cfg := &Config{
		Host:   os.Getenv("HOST"),
		Secure: strings.ToLower(os.Getenv("SECURE")) == "true",
	}

	validate(cfg)

	return cfg
}

func validate(cfg *Config) {
	if cfg.Host == "" {
		utils.Logger.Fatal("Host (e.g. example.com) must be set in environment variables")
	}
}
