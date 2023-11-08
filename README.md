# User Admin gRPC Go

User Admin gRPC Go is a sample Go application that demonstrates a user management system using gRPC for communication and a PostgreSQL database for data storage.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [License](#license)

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

The User Admin gRPC Go application relies on environment variables for configuration. Create a `.env` file in the project root directory with the following environment variables:

```bash
DB_HOST=your_database_host
DB_PORT=your_database_port
DB_USER=your_database_user
DB_PASSWORD=your_database_password
DB_NAME=your_database_name
GRPC_PORT=your_desired_grpc_port
```

Make sure to replace the placeholders with your actual database and gRPC server details.

## Usage

To compile and run the User Admin gRPC Go application, follow these steps:

1. Start your PostgreSQL database.

2. In the project directory, build the application by running:

   ```bash
   go build cmd/main.go
   ```

3. Run the application:

   ```bash
   ./main
   ```

The gRPC server will start, and you can begin using the API.

## API Documentation

For API documentation, refer to the gRPC service definitions in the `gen` directory. You can use tools like [gRPCurl](https://github.com/fullstorydev/grpcurl) or [gRPC GUI](https://github.com/fullstorydev/grpcui) to explore and interact with the API.

## Contributing

If you'd like to contribute to this project, please follow the standard GitHub flow:

1. Fork the repository.

2. Create a feature branch:

   ```bash
   git checkout -b feature/new-feature
   ```

3. Commit your changes:

   ```bash
   git commit -m "Add new feature"
   ```

4. Push to your fork:

   ```bash
   git push origin feature/new-feature
   ```

5. Create a pull request.

We welcome your contributions!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```

You can copy and paste this content into a `README.md` file in your GitHub repository, and it will display the command sections as code blocks.
