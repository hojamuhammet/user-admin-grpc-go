# User Administration

User administration - is a simple Go back end that demonstrates a user management system using gRPC for communication and a PostgreSQL database for data storage.

## Table of Contents

- [User Administration](#user-administration)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Compilation of Proto Files](#compilation-of-proto-files)
  - [Usage](#usage)

## Prerequisites

Before you can run the User Admin gRPC Go application, you need to ensure that you have the following software and resources installed on your system:

- **Go**: You must have Go installed. You can download it from [https://golang.org/dl/](https://golang.org/dl/).

- **PostgreSQL**: You need a PostgreSQL database up and running. You can install it locally or use a remote server. Make sure to configure the database connection details in the `.env` file (see [Configuration](#configuration)).

## Installation

1. Clone this repository to your local machine:

   ```bash
   git clone https://github.com/hojamuhammet/user-admin-grpc-go.git
   ```

2. Change to the project directory:

   ```bash
   cd user-admin-grpc-go
   ```

3. Install the required dependencies by running:

   ```bash
   go get -u ./...
   ```

## Configuration
Create a .env file in the project root directory with the following environment variables:
```bash
DB_HOST=your_database_host
DB_PORT=your_database_port
DB_USER=your_database_user
DB_PASSWORD=your_database_password
DB_NAME=your_database_name
GRPC_PORT=your_desired_grpc_port
```

# Compilation of Proto Files
1. Install protoc:

   You can download and install protoc from the official Protocol Buffers website: https://protobuf.dev/downloads/

2. Install the Go Protocol Buffers plugin:

   You can install the Go Protocol Buffers plugin using the following command:
   ```bash
   go get google.golang.org/protobuf/cmd/protoc-gen-go
   ```
3. Compile the .proto files:

   Run the following command in the project root directory to compile the .proto files:
   ```bash
   protoc --go_out=. --go-grpc_out=. api/user.proto
   ```
   This command generates Go code for the gRPC service in the gen/ directory based on the user.proto file.

## Usage

To run the application, follow these steps:

1. Start your PostgreSQL database.

2. In the project directory, build the application by running:

   ```bash
   go build cmd/main.go
   ```

3. Run the application:

   ```bash
   ./main
   ```
