package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/hojamuhammet/user-admin-grpc-go/gen"
	"github.com/hojamuhammet/user-admin-grpc-go/internal/config"
	service "github.com/hojamuhammet/user-admin-grpc-go/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	ctx context.Context
	cfg *config.Config
	server *grpc.Server
	wg sync.WaitGroup
	pb.UnimplementedUserServiceServer
}

func NewServer(ctx context.Context, cfg *config.Config) *Server {
	return &Server {
		ctx: ctx,
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.cfg.GRPCPort))
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()

	pb.RegisterUserServiceServer(s.server, &service.UserService{})

	reflection.Register(s.server)

	log.Printf("gRPC server started on port %s", s.cfg.GRPCPort)
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	if s.server != nil {
		log.Println("Shutting down gRPC server...")
		s.server.GracefulStop()
	}
}

func (s *Server) Wait() {
	s.wg.Wait()
}