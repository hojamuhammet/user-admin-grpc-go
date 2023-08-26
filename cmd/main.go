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
	"github.com/hojamuhammet/user-admin-grpc-go/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the database connection
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize the database: %v", err)
	}
	defer database.GetDB().Close() // Close the database connection when the program exits

	// Create a gRPC server
	grpcServer := server.NewServer(ctx, cfg)

	// Create a user service instance
	userSvc := service.NewUserService(cfg)

	// Start the gRPC server
	go func() {
		if err := grpcServer.Start(userSvc); err != nil {
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
