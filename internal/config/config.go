package config

import (
	"fmt"
	"os"
)

// Config contains configuration settings for the application.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	GRPCPort   string
}

// LoadConfig loads the configuration from environment variables and returns a Config instance.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		GRPCPort:   os.Getenv("GRPC_PORT"),
	}

	missingFields := []string{}
	if cfg.DBHost == "" {
		missingFields = append(missingFields, "DB_HOST")
	}
	if cfg.DBPort == "" {
		missingFields = append(missingFields, "DB_PORT")
	}
	if cfg.DBUser == "" {
		missingFields = append(missingFields, "DB_USER")
	}
	if cfg.DBPassword == "" {
		missingFields = append(missingFields, "DB_PASSWORD")
	}
	if cfg.DBName == "" {
		missingFields = append(missingFields, "DB_NAME")
	}

	if len(missingFields) > 0 {
		return nil, fmt.Errorf("missing or empty required configuration values: %s", missingFields)
	}

	return cfg, nil
}
