package service

import (
	"context"
	"database/sql"
	"log"

	pb "github.com/hojamuhammet/user-admin-grpc-go/gen"
	"github.com/hojamuhammet/user-admin-grpc-go/internal/config"
	"github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	cfg *config.Config
	db *sql.DB
	pb.UnimplementedUserServiceServer
}

func NewUserService(cfg *config.Config, db *sql.DB) pb.UserServiceServer {
    return &UserService{cfg: cfg,
		db: db}

}

func (us *UserService) RegisterService(server *grpc.Server) {
	pb.RegisterUserServiceServer(server, us)
}

func (us *UserService) GetAllUsers(ctx context.Context, empty *pb.Empty) (*pb.UsersList, error) {
	rows, err := us.db.QueryContext(ctx, "SELECT id, first_name, last_name, phone_number, blocked, registration_date FROM users")
    if err != nil {
        log.Printf("Error querying database: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }
    defer rows.Close()

    var users []*pb.User

    for rows.Next() {
        var user pb.User
        var registrationDate pq.NullTime
        if err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Blocked, &registrationDate); err != nil {
            log.Printf("Error scanning rows: %v", err)
            return nil, status.Errorf(codes.Internal, "Internal server error")
        }

        if registrationDate.Valid {
            user.RegistrationDate = timestamppb.New(registrationDate.Time)
        }

        users = append(users, &user)
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error iterating over rows: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    return &pb.UsersList{Users: users}, nil
}