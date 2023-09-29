CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(30),
    last_name VARCHAR(30),
    phone_number VARCHAR(12) NOT NULL,
    blocked BOOLEAN NOT NULL DEFAULT false,
    registration_date TIMESTAMP DEFAULT TO_TIMESTAMP(TO_CHAR(CURRENT_TIMESTAMP, 'DD-MM-YYYY HH24:MI:SS'), 'DD-MM-YYYY HH24:MI:SS'),
    otp VARCHAR(6),
    otp_created_at TIMESTAMP DEFAULT TO_TIMESTAMP(TO_CHAR(CURRENT_TIMESTAMP, 'DD-MM-YYYY HH24:MI:SS'), 'DD-MM-YYYY HH24:MI:SS'),
    gender VARCHAR(10),
    date_of_birth DATE,
    location VARCHAR(100),
    email VARCHAR(100) UNIQUE NOT NULL
);
