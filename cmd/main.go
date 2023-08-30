package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hojamuhammet/user-admin-grpc-go/internal/config"
	"github.com/hojamuhammet/user-admin-grpc-go/internal/database"
	"github.com/hojamuhammet/user-admin-grpc-go/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the database connection
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize the database: %v", err)
	}
	defer db.Close() // Close the database connection when the program exits

	// Create a gRPC server
	grpcServer := server.NewServer(ctx, cfg, db)

	// Start the gRPC server
	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down gRPC server...")
	grpcServer.Stop()
	log.Println("gRPC server stopped")

	// Add any additional cleanup logic here
	log.Println("Application gracefully terminated")
}