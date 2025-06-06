-- Create users table with proper data types and constraints
-- Using UUID for ID provides good distribution across Cassandra nodes
-- Note: Keyspace is now created dynamically during connection setup
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT,
    name TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Create index on email for efficient lookups
-- In production, consider using a separate table for email->id mapping
CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);
