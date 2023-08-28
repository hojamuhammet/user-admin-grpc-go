package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/hojamuhammet/user-admin-grpc-go/internal/config"
)

var db *sql.DB

func InitDB(cfg *config.Config) error {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	log.Printf("Connection String: %s", connectionString)

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Connected to the database")
	return nil
}

func GetDB() *sql.DB {
	return db
}