package config

import (
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
	HTTPPort   string
}

// LoadConfig loads the configuration from environment variables and returns a Config instance.
func LoadConfig() (*Config, error) {
	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		GRPCPort:   os.Getenv("GRPC_PORT"),
		HTTPPort: 	os.Getenv("HTTP_Port"),
	}, nil
}
