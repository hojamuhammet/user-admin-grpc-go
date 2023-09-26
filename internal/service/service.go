package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"

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

// NewUserService creates a new instance of UserService with the provided configuration and database connection.
func NewUserService(cfg *config.Config, db *sql.DB) pb.UserServiceServer {
    return &UserService{cfg: cfg, db: db}
}

// RegisterService registers the UserService with a gRPC server.
func (us *UserService) RegisterService(server *grpc.Server) {
	pb.RegisterUserServiceServer(server, us)
}

// Regular expression pattern for a valid phone number
var phoneNumberPattern = regexp.MustCompile(`^\+993\d{8}$`)

// GetAllUsers retrieves a list of all users from the database.
func (us *UserService) GetAllUsers(ctx context.Context, empty *pb.Empty) (*pb.UsersList, error) {
    // Execute a SQL query to select user data from the database
	rows, err := us.db.QueryContext(ctx, "SELECT id, first_name, last_name, phone_number, blocked, registration_date FROM users")
    if err != nil {
        log.Printf("Error querying database: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }
    defer rows.Close()

    var users []*pb.User

    // Iterate over the rows returned by the query
    for rows.Next() {
        var user pb.User
        var registrationDate pq.NullTime

        // Scan the row data into user and registrationDate
        if err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Blocked, &registrationDate); err != nil {
            log.Printf("Error scanning rows: %v", err)
            return nil, status.Errorf(codes.Internal, "Internal server error")
        }

        // If registrationDate is valid, convert it to a protobuf Timestamp
        if registrationDate.Valid {
            user.RegistrationDate = timestamppb.New(registrationDate.Time)
        }

        // Append the user to the list of users
        users = append(users, &user)
    }

    // Check for any errors that occurred during iteration
    if err := rows.Err(); err != nil {
        log.Printf("Error iterating over rows: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    // Return the list of users as a UsersList response
    return &pb.UsersList{Users: users}, nil
}

func (us *UserService) GetUserById(ctx context.Context, req *pb.UserID) (*pb.User, error) {
    // Query to retrieve user by ID
    query := "SELECT id, first_name, last_name, phone_number, blocked, registration_date FROM users WHERE id = $1"

    // Variables to store user details
    var user pb.User
    var registrationDate pq.NullTime

    // Execute the query with the user's ID
    err := us.db.QueryRowContext(ctx, query, req.Id).Scan(
        &user.Id,
        &user.FirstName,
        &user.LastName,
        &user.PhoneNumber,
        &user.Blocked,
        &registrationDate, // Scan registration_date as pq.NullTime
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Errorf(codes.NotFound, "User not found")
        }
        log.Printf("Error fetching user by ID: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    // Check if registrationDate is NULL in the database
    if registrationDate.Valid {
        // Assign the registrationDate directly to user.RegistrationDate
        user.RegistrationDate = timestamppb.New(registrationDate.Time)
    } else {
        user.RegistrationDate = nil // Set RegistrationDate to nil in the protobuf message
    }

    return &user, nil
}

func (us *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    // Extract user creation data from req
    firstName := req.FirstName
    lastName := req.LastName
    phoneNumber := req.PhoneNumber

    // Validate the phone number using the regular expression pattern
    if !phoneNumberPattern.MatchString(phoneNumber) {
        return nil, status.Errorf(codes.InvalidArgument, "Invalid phone number format")
    }

    // Insert the new user into the database
    query := `
        INSERT INTO users (first_name, last_name, phone_number, blocked)
        VALUES ($1, $2, $3, $4)
        RETURNING id, first_name, last_name, phone_number, blocked, registration_date
    `
    var user pb.User
    var registrationDate pq.NullTime
    err := us.db.QueryRowContext(ctx, query, firstName, lastName, phoneNumber, false).Scan(
        &user.Id,
        &user.FirstName,
        &user.LastName,
        &user.PhoneNumber,
        &user.Blocked,
        &registrationDate,
    )
    if err != nil {
        log.Printf("Error creating user: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }
    
    if registrationDate.Valid {
        user.RegistrationDate = timestamppb.New(registrationDate.Time)
    } else {
        user.RegistrationDate = nil // Set RegistrationDate to nil in the protobuf message
    }

    return &user, nil
}

func (us *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
    // Validate the phone number using the regular expression pattern
    if !phoneNumberPattern.MatchString(req.PhoneNumber) {
        return nil, status.Errorf(codes.InvalidArgument, "Invalid phone number format")
    }

    // Define the SQL query to update user details
    query := `
        UPDATE users
        SET first_name = $2, last_name = $3, phone_number = $4
        WHERE id = $1
        RETURNING id, first_name, last_name, phone_number, blocked, registration_date
    `
    
    // Variables to store updated user details
    var updatedUser pb.User
    var registrationDate pq.NullTime
    
    // Execute the query to update the user's details
    err := us.db.QueryRowContext(ctx, query, req.Id, req.FirstName, req.LastName, req.PhoneNumber).
        Scan(
            &updatedUser.Id,
            &updatedUser.FirstName,
            &updatedUser.LastName,
            &updatedUser.PhoneNumber,
            &updatedUser.Blocked,
            &registrationDate, // Scan registration_date as pq.NullTime
        )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Errorf(codes.NotFound, "User not found")
        }
        log.Printf("Error updating user: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }
    
    // Check if registrationDate is NULL in the database
    if registrationDate.Valid {
        // Assign the registrationDate directly to updatedUser.RegistrationDate
        updatedUser.RegistrationDate = timestamppb.New(registrationDate.Time)
    } else {
        updatedUser.RegistrationDate = nil // Set RegistrationDate to nil in the protobuf message
    }
    
    return &updatedUser, nil
}

// DeleteUser deletes a user from the database by their ID and returns an empty response.
func (us *UserService) DeleteUser(ctx context.Context, userID *pb.UserID) (*pb.Empty, error) {
	log.Printf("Deleting user with ID: %d", userID.Id)

	// Execute a DELETE query with a WHERE clause to remove the user with the given ID.
	result, err := us.db.Exec("DELETE FROM users WHERE id=$1", userID.Id)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		return nil, status.Error(codes.Internal, "Failed to delete user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return nil, status.Error(codes.Internal, "Failed to delete user")
	}

	if rowsAffected == 0 {
		log.Printf("User not found with ID: %d", userID.Id)
		return nil, status.Error(codes.NotFound, "User not found")
	}

	log.Printf("User with ID %d successfully deleted", userID.Id)
	return &pb.Empty{}, nil
}

func (us *UserService) toggleBlockStatus(ctx context.Context, userID *pb.UserID, blocked bool) error {
	// Execute an UPDATE query with a WHERE clause to set the "blocked" field to the specified status for the given user ID.
	result, err := us.db.Exec("UPDATE users SET blocked=$1 WHERE id=$2", blocked, userID.Id)
	if err != nil {
		log.Printf("Failed to update user status (UserID: %d): %v", userID.Id, err)
		return status.Error(codes.Internal, "Failed to update user status")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to retrieve rows affected (UserID: %d): %v", userID.Id, err)
		return status.Error(codes.Internal, "Failed to retrieve rows affected")
	}

	if rowsAffected == 0 {
		log.Printf("User not found (UserID: %d)", userID.Id)
		return status.Error(codes.NotFound, fmt.Sprintf("User with ID %d not found", userID.Id))
	}

	return nil
}

// BlockUser updates the "blocked" status of a user in the database and returns an empty response.
func (us *UserService) BlockUser(ctx context.Context, userID *pb.UserID) (*pb.Empty, error) {
	if err := us.toggleBlockStatus(ctx, userID, true); err != nil {
		if status.Code(err) == codes.NotFound {
			log.Printf("User not found: %v", err)
			return nil, status.Error(codes.NotFound, "User not found")
		}
		log.Printf("Internal server error (BlockUser): %v", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	log.Printf("User with ID %d successfully blocked", userID.Id)
	return &pb.Empty{}, nil
}

// UnblockUser updates the "blocked" status of a user in the database and returns an empty response.
func (us *UserService) UnblockUser(ctx context.Context, userID *pb.UserID) (*pb.Empty, error) {
	if err := us.toggleBlockStatus(ctx, userID, false); err != nil {
		if status.Code(err) == codes.NotFound {
			log.Printf("User not found: %v", err)
			return nil, status.Error(codes.NotFound, "User not found")
		}
		log.Printf("Internal server error (UnblockUser): %v", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	
	log.Printf("User with ID %d successfully unblocked", userID.Id)
	return &pb.Empty{}, nil
}
