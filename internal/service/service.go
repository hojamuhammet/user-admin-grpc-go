package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

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
    return &UserService{
        cfg: cfg, 
        db: db,
    }
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
    rows, err := us.db.QueryContext(ctx, "SELECT id, first_name, last_name, phone_number, blocked, registration_date, gender, date_of_birth, location, email, profile_photo_url FROM users")
    if err != nil {
        // Log the error and return an internal server error status
        log.Printf("Error querying database: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }
    defer rows.Close()

    var users []*pb.GetUserResponse

    // Iterate over the rows returned by the query
    for rows.Next() {
        var user pb.GetUserResponse
        var registrationTime time.Time
        var dateOfBirthStr string

        // Scan the row data into user and registrationDate
        if err := rows.Scan(
            &user.Id, 
            &user.FirstName, 
            &user.LastName, 
            &user.PhoneNumber, 
            &user.Blocked, 
            &registrationTime,
            &user.Gender,
            &dateOfBirthStr,
            &user.Location,
            &user.Email,
            &user.ProfilePhotoUrl,
        ); err != nil {
            // Log the error and return an internal server error status
            log.Printf("Error scanning rows: %v", err)
            return nil, status.Errorf(codes.Internal, "Internal server error")
        }

        // Assign registrationTimestamp directly to user.RegistrationDate
        user.RegistrationDate = timestamppb.New(registrationTime)

        // Extract the date portion of the dateOfBirthStr (YYYY-MM-DD)
        dateOfBirthStr = dateOfBirthStr[:10] // This removes the "T00:00:00Z" part

        // Parse the dateOfBirthStr into a time.Time
        dateOfBirthTime, err := time.Parse("2006-01-02", dateOfBirthStr)
        if err != nil {
            // Log the error and return an internal server error status
            log.Printf("Error parsing date: %v", err)
            return nil, status.Errorf(codes.Internal, "Internal server error")
        }

        // Convert the dateOfBirthTime into a pb.DateOfBirth message
        user.DateOfBirth = &pb.DateOfBirth{
            Year:  int32(dateOfBirthTime.Year()),
            Month: int32(dateOfBirthTime.Month()),
            Day:   int32(dateOfBirthTime.Day()),
        }

        // Append the user to the list of users
        users = append(users, &user)
    }

    // Check for any errors that occurred during iteration
    if err := rows.Err(); err != nil {
        // Log the error and return an internal server error status
        log.Printf("Error iterating over rows: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    // Return the list of users as a UsersList response
    return &pb.UsersList{Users: users}, nil
}


func (us *UserService) GetUserById(ctx context.Context, req *pb.UserID) (*pb.GetUserResponse, error) {
    // Query to retrieve user by ID
    query := "SELECT id, first_name, last_name, phone_number, blocked, registration_date, gender, date_of_birth, location, email, profile_photo_url FROM users WHERE id = $1"

    // Variables to store user details
    var user pb.GetUserResponse
    var registrationDate pq.NullTime
    var dateOfBirthStr string

    // Execute the query with the user's ID
    err := us.db.QueryRowContext(ctx, query, req.Id).Scan(
        &user.Id,
        &user.FirstName,
        &user.LastName,
        &user.PhoneNumber,
        &user.Blocked,
        &registrationDate, // Scan registration_date as pq.NullTime
        &user.Gender,
        &dateOfBirthStr,
        &user.Location,
        &user.Email,
        &user.ProfilePhotoUrl,
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

     // Extract the date portion of the dateOfBirthStr (YYYY-MM-DD)
     dateOfBirthStr = dateOfBirthStr[:10] // This removes the "T00:00:00Z" part

     // Parse the dateOfBirthStr into a time.Time
     dateOfBirthTime, err := time.Parse("2006-01-02", dateOfBirthStr)
     if err != nil {
         log.Printf("Error parsing date: %v", err)
         return nil, status.Errorf(codes.Internal, "Internal server error")
     }
 
     // Convert the dateOfBirthTime into a pb.DateOfBirth message
     user.DateOfBirth = &pb.DateOfBirth{
         Year:  int32(dateOfBirthTime.Year()),
         Month: int32(dateOfBirthTime.Month()),
         Day:   int32(dateOfBirthTime.Day()),
     }

    return &user, nil
}

func (us *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    // Extract user creation data from req
    firstName := req.FirstName
    lastName := req.LastName
    phoneNumber := req.PhoneNumber
    gender := req.Gender
    location := req.Location
    email := req.Email
    profilePhotoUrl := req.ProfilePhotoUrl

    // Validate the phone number using the regular expression pattern
    if !phoneNumberPattern.MatchString(phoneNumber) {
        return nil, status.Errorf(codes.InvalidArgument, "Invalid phone number format")
    }

    var dateOfBirthTime pq.NullTime

    if req.DateOfBirth != nil {
        dateOfBirthTime.Time = time.Date(int(req.DateOfBirth.Year), time.Month(req.DateOfBirth.Month), int(req.DateOfBirth.Day), 0, 0, 0, 0, time.UTC)
        dateOfBirthTime.Valid = true
    }

    query := `
        INSERT INTO users (first_name, last_name, phone_number, blocked, gender, date_of_birth, location, email, profile_photo_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, first_name, last_name, phone_number, blocked, gender, date_of_birth, location, email, profile_photo_url
    `
    var user pb.CreateUserResponse

    err := us.db.QueryRowContext(ctx, query, firstName, lastName, phoneNumber, false, gender, dateOfBirthTime, location, email, profilePhotoUrl).Scan(
        &user.Id,
        &user.FirstName,
        &user.LastName,
        &user.PhoneNumber,
        &user.Blocked,
        &user.Gender,
        &dateOfBirthTime, // This value is not modified before Scan
        &user.Location,
        &user.Email,
        &user.ProfilePhotoUrl,
    )

    if err != nil {
        log.Printf("Error creating user: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    // Convert the pq.NullTime value to a DateOfBirth protobuf
    if dateOfBirthTime.Valid {
        user.DateOfBirth = &pb.DateOfBirth{
            Year:  int32(dateOfBirthTime.Time.Year()),
            Month: int32(dateOfBirthTime.Time.Month()),
            Day:   int32(dateOfBirthTime.Time.Day()),
        }
    } else {
        // Set user.DateOfBirth to nil when the date of birth is NULL in the database
        user.DateOfBirth = nil
    }

    return &user, nil
}

func (us *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
    // Validate the phone number using the regular expression pattern
    if !phoneNumberPattern.MatchString(req.PhoneNumber) {
        return nil, status.Errorf(codes.InvalidArgument, "Invalid phone number format")
    }

    // Format the dateOfBirth as a string in the format "YYYY-MM-DD"
    dateOfBirthStr := fmt.Sprintf("%04d-%02d-%02d", req.DateOfBirth.Year, req.DateOfBirth.Month, req.DateOfBirth.Day)

    // Define the SQL query to update user details, including the profile_photo_url
    query := `
        UPDATE users
        SET first_name = $2, last_name = $3, phone_number = $4, gender = $5, date_of_birth = $6, location = $7, email = $8, profile_photo_url = $9
        WHERE id = $1
        RETURNING id, first_name, last_name, phone_number, blocked, gender, date_of_birth, location, email, profile_photo_url
    `
    
    // Variables to store updated user details
    var updatedUser pb.UpdateUserResponse
    
    // Execute the query to update the user's details
    err := us.db.QueryRowContext(ctx, query, req.Id, req.FirstName, req.LastName, req.PhoneNumber, req.Gender, dateOfBirthStr, req.Location, req.Email, req.ProfilePhotoUrl).
        Scan(
            &updatedUser.Id,
            &updatedUser.FirstName,
            &updatedUser.LastName,
            &updatedUser.PhoneNumber,
            &updatedUser.Gender,
            &dateOfBirthStr, // Scan date_of_birth as a string
            &updatedUser.Location,
            &updatedUser.Email,
            &updatedUser.ProfilePhotoUrl,
        )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Errorf(codes.NotFound, "User not found")
        }
        log.Printf("Error updating user: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
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
