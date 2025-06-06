-- Create secondary indexes for efficient queries on federated_users table
CREATE INDEX IF NOT EXISTS federated_users_email_idx ON mapleapps.federated_users (email);
