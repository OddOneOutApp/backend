package config

import (
	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/joho/godotenv"
)

type Config struct {
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		utils.Logger.Errorf("Error loading .env file: %v", err)
	}

	cfg := &Config{}

	validate(cfg)

	return cfg
}

func validate(cfg *Config) {
	/* if cfg.TestField == "" {
		utils.Logger.Fatal("Test field must be set in environment variables")
	} */
}
