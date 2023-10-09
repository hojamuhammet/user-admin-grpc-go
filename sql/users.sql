CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(30),
    last_name VARCHAR(30),
    phone_number VARCHAR(12) NOT NULL UNIQUE,
    blocked BOOLEAN NOT NULL DEFAULT false,
    registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    otp VARCHAR(6) UNIQUE,
    otp_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    gender VARCHAR(10),
    date_of_birth DATE,
    location VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    profile_photo_url VARCHAR(255)
    -- Add a partial unique index for non-null email values
    CREATE UNIQUE INDEX email_unique_idx ON users (email) WHERE email IS NOT NULL;

);
