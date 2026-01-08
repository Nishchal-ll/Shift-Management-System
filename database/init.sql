-- Create Users Table
CREATE TABLE IF NOT EXISTS users (
    username VARCHAR(255) PRIMARY KEY,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user'
);

-- Create Other Tables
CREATE TABLE IF NOT EXISTS shift_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    hours VARCHAR(50) NOT NULL,
    quota INT DEFAULT 5,
    start_time VARCHAR(10) NOT NULL,
    end_time VARCHAR(10) NOT NULL
);

CREATE TABLE IF NOT EXISTS allocations (
    id SERIAL PRIMARY KEY,
    employee_name VARCHAR(255) REFERENCES users(username),
    shift_name VARCHAR(50) REFERENCES shift_types(name),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'Confirmed',
    new_requested_shift VARCHAR(50) DEFAULT ''
);

-- INSERT YOUR SPECIFIC CREDENTIALS
INSERT INTO users (username, password, role) VALUES 
    ('admin', 'admin123', 'admin')
ON CONFLICT (username) DO NOTHING;

