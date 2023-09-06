package server

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/hojamuhammet/go-grpc-otp-rabbitmq/internal/pkg/config"

	pb "github.com/hojamuhammet/go-grpc-otp-rabbitmq/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents your gRPC server.
type Server struct {
    server *grpc.Server
	pb.UnimplementedUserServiceServer
	
}

// NewServer creates a new instance of the Server.
func NewServer() *Server {
    return &Server{}
}

// Start starts the gRPC server.
func (s *Server) Start(ctx context.Context, cfg *config.Config) error {
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        return err
    }

    s.server = grpc.NewServer()

    // Register any gRPC services here if needed

    // Example: pb.RegisterSomeServiceServer(s.server, &someService{})

    reflection.Register(s.server)

    log.Printf("gRPC server started on port %s", cfg.GRPCPort)

    return s.server.Serve(lis)
}

// Stop stops the gRPC server gracefully.
func (s *Server) Stop() {
    if s.server != nil {
        s.server.GracefulStop()
    }
}

// Wait waits for the server to finish gracefully.
func (s *Server) Wait() {
	var wg sync.WaitGroup
    wg.Wait()
}
