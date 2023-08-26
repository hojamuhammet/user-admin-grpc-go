package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/hojamuhammet/user-admin-grpc-go/internal/config"
)

var db *sql.DB

func InitDB(cfg *config.Config) error {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
	cfg.DBUser, cfg.DBPassword, cfg.DBUser, cfg.DBHost, cfg.DBPort)

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to the database")
	return nil
}

func GetDB() *sql.DB {
	return db
}