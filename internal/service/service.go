package service

import (
	pb "github.com/hojamuhammet/user-admin-grpc-go/gen"
	"github.com/hojamuhammet/user-admin-grpc-go/internal/config"
	"google.golang.org/grpc"
)

type UserService struct {
	cfg *config.Config
	pb.UnimplementedUserServiceServer
}

func NewUserService(cfg *config.Config) pb.UserServiceServer {
    return &UserService{cfg: cfg}
}

func (us *UserService) RegisterService(server *grpc.Server) {
	pb.RegisterUserServiceServer(server, us)
}