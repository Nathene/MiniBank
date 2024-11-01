CREATE TABLE IF NOT EXISTS account (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Change SERIAL to INTEGER PRIMARY KEY AUTOINCREMENT
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(50),
    phone_number INTEGER,                -- Change SERIAL to INTEGER
    encrypted_password VARCHAR(100),
    balance INTEGER,                     -- Change SERIAL to INTEGER
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);