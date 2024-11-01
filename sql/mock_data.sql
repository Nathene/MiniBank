-- Insert some mock data into the account table
INSERT INTO account (first_name, last_name, email, phone_number, encrypted_password, balance, created_at, updated_at) VALUES
('John', 'Doe', 'john.doe@example.com', 1234567890, 'encrypted_password_1', 1000.00, '2024-11-01T12:00:00Z', '2024-11-01T12:00:00Z'),
('Jane', 'Smith', 'jane.smith@example.com', 9876543210, 'encrypted_password_2', 500.50, '2024-10-20T09:30:00Z', '2024-10-20T09:30:00Z'),
('Alice', 'Johnson', 'alice.johnson@example.com', 5551237890, 'encrypted_password_3', 2500.75, '2024-09-15T15:45:00Z', '2024-09-15T15:45:00Z'),
('Bob', 'Williams', 'bob.williams@example.com', 1112223330, 'encrypted_password_4', 100.00, '2024-11-01T08:15:00Z', '2024-11-01T08:15:00Z');