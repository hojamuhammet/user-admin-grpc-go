package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	pb "github.com/hojamuhammet/user-admin-grpc-go/gen"
	"github.com/hojamuhammet/user-admin-grpc-go/internal/config"
	"github.com/hojamuhammet/user-admin-grpc-go/internal/utils"
	"github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (us *UserService) GetAllUsers(ctx context.Context, req *pb.PaginationRequest) (*pb.UsersList, error) {
    pageSize := req.PageSize
    if pageSize <= 0 {
        pageSize = 12 // Default page size
    }
    page := req.Page
    previousPage := req.PreviousPage

    // Handle negative page numbers and set them to 1
    if page <= 0 {
        page = 1
    }

    // Handle previous page requests with no previous page available
    if previousPage <= 0 {
        previousPage = 1
    }

    var query string
    var args []interface{}

    // Calculate the offset based on the page or previous page
    offset := (page - 1) * pageSize

    query = "SELECT id, first_name, last_name, phone_number, blocked, registration_date, gender, date_of_birth, location, email, profile_photo_url FROM users ORDER BY id LIMIT $1 OFFSET $2"
    args = append(args, pageSize, offset)

    // Execute the SQL query to select user data from the database
    rows, err := us.db.QueryContext(ctx, query, args...)
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
        var firstName sql.NullString
        var lastName sql.NullString
        var registrationDate time.Time
        var gender sql.NullString
        var dateOfBirth sql.NullTime
        var location sql.NullString
        var email sql.NullString
        var profilePhotoUrl sql.NullString

        // Scan the row data into user, registrationDate, and other fields
        if err := rows.Scan(
            &user.Id,
            &firstName,
            &lastName,
            &user.PhoneNumber,
            &user.Blocked,
            &registrationDate,
            &gender,
            &dateOfBirth,
            &location,
            &email,
            &profilePhotoUrl,
        ); err != nil {
            // Log the error and return an internal server error status
            log.Printf("Error scanning rows: %v", err)
            return nil, status.Errorf(codes.Internal, "Internal server error")
        }

        // Convert the registration timestamp to the custom type
        customTimestamp := &pb.CustomTimestamp{
            Year:   int32(registrationDate.Year()),
            Month:  int32(registrationDate.Month()),
            Day:    int32(registrationDate.Day()),
            Hour:   int32(registrationDate.Hour()),
            Minute: int32(registrationDate.Minute()),
            Second: int32(registrationDate.Second()),
        }

        // Set the custom registration date in the user response
        user.RegistrationDate = customTimestamp

        user.FirstName = utils.NullableStringToString(firstName.Valid, firstName.String)
        user.LastName = utils.NullableStringToString(lastName.Valid, lastName.String)
        user.Gender = utils.NullableStringToString(gender.Valid, gender.String)

        // Convert the pq.NullTime value to a DateOfBirth protobuf [for response only]
        if dateOfBirth.Valid {
            user.DateOfBirth = &pb.DateOfBirth{
                Year:  int32(dateOfBirth.Time.Year()),
                Month: int32(dateOfBirth.Time.Month()),
                Day:   int32(dateOfBirth.Time.Day()),
            }
        } else {
            user.DateOfBirth = nil // Set user.DateOfBirth to nil when date of birth is NULL
        }

        user.Location = utils.NullableStringToString(location.Valid, location.String)
        user.Email = utils.NullableStringToString(email.Valid, email.String)
        user.ProfilePhotoUrl = utils.NullableStringToString(profilePhotoUrl.Valid, profilePhotoUrl.String)

        // Append the user to the list of users
        users = append(users, &user)
    }

    // Check for any errors that occurred during iteration
    if err := rows.Err(); err != nil {
        // Log the error and return an internal server error status
        log.Printf("Error iterating over rows: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    // Determine the next page number based on the current page and page size
    nextPage := page + 1
    previousPage = page - 1

    // Return the list of users as a UsersList response
    return &pb.UsersList{Users: users, NextPage: nextPage, PreviousPage: previousPage}, nil
}

func (us *UserService) GetUserById(ctx context.Context, req *pb.UserID) (*pb.GetUserResponse, error) {
    // Query to retrieve user by ID
    query := "SELECT id, first_name, last_name, phone_number, blocked, registration_date, gender, date_of_birth, location, email, profile_photo_url FROM users WHERE id = $1"

    // Variables to store user details
    var user pb.GetUserResponse
    var firstName sql.NullString
    var lastName sql.NullString
    var registrationDate time.Time
    var gender sql.NullString
    var dateOfBirth sql.NullTime
    var location sql.NullString
    var email sql.NullString
    var profilePhotoUrl sql.NullString
    

    // Execute the query with the user's ID
    err := us.db.QueryRowContext(ctx, query, req.Id).Scan(
        &user.Id,
        &firstName,
        &lastName,
        &user.PhoneNumber,
        &user.Blocked,
        &registrationDate, // Scan registration_date as pq.NullTime
        &gender,
        &dateOfBirth,
        &location,
        &email,
        &profilePhotoUrl,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Errorf(codes.NotFound, "User not found")
        }
        log.Printf("Error fetching user by ID: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    // Convert the registration timestamp to the custom type
    customTimestamp := &pb.CustomTimestamp{
        Year:   int32(registrationDate.Year()),
        Month:  int32(registrationDate.Month()),
        Day:    int32(registrationDate.Day()),
        Hour:   int32(registrationDate.Hour()),
        Minute: int32(registrationDate.Minute()),
        Second: int32(registrationDate.Second()),
    }

    // Set the custom registration date in the user response
    user.RegistrationDate = customTimestamp

    user.FirstName = utils.NullableStringToString(firstName.Valid, firstName.String)
    user.LastName = utils.NullableStringToString(lastName.Valid, lastName.String)
    user.Gender = utils.NullableStringToString(gender.Valid, gender.String)

    // Convert the pq.NullTime value to a DateOfBirth protobuf [for response only]
    if dateOfBirth.Valid {
        user.DateOfBirth = &pb.DateOfBirth{
            Year:  int32(dateOfBirth.Time.Year()),
            Month: int32(dateOfBirth.Time.Month()),
            Day:   int32(dateOfBirth.Time.Day()),
        }
    } else {
        user.DateOfBirth = nil // Set user.DateOfBirth to nil when date of birth is NULL
    }

    user.Location = utils.NullableStringToString(location.Valid, location.String)
    user.Email = utils.NullableStringToString(email.Valid, email.String)
    user.ProfilePhotoUrl = utils.NullableStringToString(profilePhotoUrl.Valid, profilePhotoUrl.String)

    return &user, nil
}

func (us *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    // Validate the phone number using the regular expression pattern
    if !phoneNumberPattern.MatchString(req.PhoneNumber) {
        return nil, status.Errorf(codes.InvalidArgument, "Invalid phone number format")
    }

    var firstName = utils.CreateNullString(req.FirstName)
    var lastName = utils.CreateNullString(req.LastName)
    var gender = utils.CreateNullString(req.Gender)

    var dateOfBirth pq.NullTime
    if req.DateOfBirth != nil {
        dateOfBirth.Time = utils.ToDate(req.DateOfBirth.Year, req.DateOfBirth.Month, req.DateOfBirth.Day)
        dateOfBirth.Valid = true
    }

    var location = utils.CreateNullString(req.Location)
    var email = utils.CreateNullString(req.Email)
    var profilePhotoUrl = utils.CreateNullString(req.ProfilePhotoUrl)

    query := `
        INSERT INTO users (first_name, last_name, phone_number, blocked, gender, date_of_birth, location, email, profile_photo_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, first_name, last_name, phone_number, blocked, gender, date_of_birth, location, email, profile_photo_url
    `
    var user pb.CreateUserResponse

    // Execute the SQL query and scan the result into user
    err := us.db.QueryRowContext(ctx, query, firstName, lastName, req.PhoneNumber, false, gender, dateOfBirth, location, email, profilePhotoUrl).Scan(
        &user.Id,
        &firstName,
        &lastName,
        &user.PhoneNumber,
        &user.Blocked,
        &gender,
        &dateOfBirth, // Not modified before Scan
        &location,
        &email,
        &profilePhotoUrl,
    )

    if err != nil {
        // Log the error and return an internal server error to the client
        log.Printf("Error creating user: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    user.FirstName = utils.NullableStringToString(firstName.Valid, firstName.String)
    user.LastName = utils.NullableStringToString(lastName.Valid, lastName.String)
    user.Gender = utils.NullableStringToString(gender.Valid, gender.String)

    // Convert the pq.NullTime value to a DateOfBirth protobuf [for response only]
    if dateOfBirth.Valid {
        user.DateOfBirth = &pb.DateOfBirth{
            Year:  int32(dateOfBirth.Time.Year()),
            Month: int32(dateOfBirth.Time.Month()),
            Day:   int32(dateOfBirth.Time.Day()),
        }
    } else {
        user.DateOfBirth = nil // Set user.DateOfBirth to nil when date of birth is NULL
    }

    user.Location = utils.NullableStringToString(location.Valid, location.String)
    user.Email = utils.NullableStringToString(email.Valid, email.String)
    user.ProfilePhotoUrl = utils.NullableStringToString(profilePhotoUrl.Valid, profilePhotoUrl.String)

    return &user, nil
}

func (us *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
    var updatedUser pb.UpdateUserResponse

    var updatedFirstName sql.NullString
    var updatedLastName sql.NullString
    var updatedGender sql.NullString
    var updatedDateOfBirth pq.NullTime
    var updatedLocation sql.NullString
    var updatedEmail sql.NullString
    var updatedProfilePhotoUrl sql.NullString

    // Build the UPDATE query
    query := "UPDATE users SET "
    var args []interface{}

    // Check and add fields to the query and args based on the provided fields in the request
    argCount := 1

    if req.FirstName == "null" {
        query += "first_name = NULL, "
    } else if req.FirstName != "" {
        query += "first_name = $" + strconv.Itoa(argCount) + ", "
        args = append(args, req.FirstName)
        argCount++
    }

    if req.LastName == "null" {
        query += "last_name = NULL, "
    } else if req.LastName != "" {
        query += "last_name = $" + strconv.Itoa(argCount) + ", "
        args = append(args, req.LastName)
        argCount++
    }

    // Validate the phone number using the regular expression pattern
    if req.PhoneNumber != "" {
        if !phoneNumberPattern.MatchString(req.PhoneNumber) {
            return nil, status.Errorf(codes.InvalidArgument, "Invalid phone number format")
        }
        query += "phone_number = $" + strconv.Itoa(argCount) + ", "
        args = append(args, req.PhoneNumber)
        argCount++
    }

    if req.Gender == "null" {
        query += "gender = NULL, "
    } else if req.Gender != "" {
        query += "gender = $" + strconv.Itoa(argCount) + ", "
        args = append(args, req.Gender)
        argCount++
    }

    if req.DateOfBirth != nil && req.DateOfBirth.Year == 0 && req.DateOfBirth.Month == 0 && req.DateOfBirth.Day == 0 {
        query += "date_of_birth = NULL, "
    } else if req.DateOfBirth != nil {
        dateOfBirth := time.Date(int(req.DateOfBirth.Year), time.Month(req.DateOfBirth.Month), int(req.DateOfBirth.Day), 0, 0, 0, 0, time.UTC)
        query += "date_of_birth = $" + strconv.Itoa(argCount) + ", "
        args = append(args, dateOfBirth)
        argCount++
    }

    if req.Location == "null" {
        query += "location = NULL, "
    } else if req.Location != "" {
        query += "location = $" + strconv.Itoa(argCount) + ", "
        args = append(args, req.Location)
        argCount++
    }

     // Handle the email field using pq.NullString
    if req.Email == "null" {
        query += "email = NULL, " // Set the email to NULL in the database
    } else if req.Email != "" {
        query += "email = $" + strconv.Itoa(argCount) + ", "
        args = append(args, req.Email)
        argCount++
    }

    if req.ProfilePhotoUrl == "null" {
        query += "profile_photo_url = NULL, "
    } else if req.ProfilePhotoUrl != "" {
        query += "profile_photo_url = $" + strconv.Itoa(argCount) + ", "
        args = append(args, req.ProfilePhotoUrl)
        argCount++
    }

    // Remove the trailing comma and space from the query
    query = strings.TrimSuffix(query, ", ")

    // Add the WHERE clause to identify the user by ID
    query += " WHERE id = $" + strconv.Itoa(argCount)
    args = append(args, req.Id)

    // Define the SQL query to return the updated user details
    query += " RETURNING id, first_name, last_name, phone_number, blocked, gender, date_of_birth, location, email, profile_photo_url"

    // Execute the query to update the user's details
    err := us.db.QueryRowContext(ctx, query, args...).
        Scan(
            &updatedUser.Id,
            &updatedFirstName,
            &updatedLastName,
            &updatedUser.PhoneNumber,
            &updatedUser.Blocked,
            &updatedGender,
            &updatedDateOfBirth,
            &updatedLocation,
            &updatedEmail,
            &updatedProfilePhotoUrl,
        )

        updatedUser.FirstName = utils.NullableStringToString(updatedFirstName.Valid, updatedFirstName.String)
        updatedUser.LastName = utils.NullableStringToString(updatedLastName.Valid, updatedLastName.String)
        updatedUser.Gender = utils.NullableStringToString(updatedGender.Valid, updatedGender.String)

        if updatedDateOfBirth.Valid {
            updatedUser.DateOfBirth = &pb.DateOfBirth{
                Year:  int32(updatedDateOfBirth.Time.Year()),
                Month: int32(updatedDateOfBirth.Time.Month()),
                Day:   int32(updatedDateOfBirth.Time.Day()),
            }
        } else {
            updatedUser.DateOfBirth = nil
        }

        updatedUser.Location = utils.NullableStringToString(updatedLocation.Valid, updatedLocation.String)
        updatedUser.Email = utils.NullableStringToString(updatedEmail.Valid, updatedEmail.String)
        updatedUser.ProfilePhotoUrl = utils.NullableStringToString(updatedProfilePhotoUrl.Valid, updatedProfilePhotoUrl.String)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Errorf(codes.NotFound, "User not found")
        }
        log.Printf("Error updating user: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal server error")
    }

    return &updatedUser, nil
}

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

func (us *UserService) ToggleBlockStatus(ctx context.Context, userID *pb.UserID, blocked bool) error {
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

func (us *UserService) BlockUser(ctx context.Context, userID *pb.UserID) (*pb.Empty, error) {
	if err := us.ToggleBlockStatus(ctx, userID, true); err != nil {
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

func (us *UserService) UnblockUser(ctx context.Context, userID *pb.UserID) (*pb.Empty, error) {
	if err := us.ToggleBlockStatus(ctx, userID, false); err != nil {
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
