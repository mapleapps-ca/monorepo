-- Create the federated_users table
-- This table stores user account information with end-to-end encryption support
-- Table 1: Primary lookup by ID
CREATE TABLE IF NOT EXISTS mapleapps.federated_users_by_id (
    id UUID PRIMARY KEY,
    email TEXT,
    first_name TEXT,
    last_name TEXT,
    name TEXT,
    lexical_name TEXT,
    role TINYINT,
    status TINYINT,
    timezone TEXT,
    created_at TIMESTAMP,
    modified_at TIMESTAMP,
    -- Store less frequently accessed data as JSON
    profile_data TEXT,  -- JSON with address, shipping info, etc.
    security_data TEXT, -- JSON with E2EE fields
    metadata TEXT      -- JSON with tracking fields
);

-- Table 2: Lookup by email (your main alternate access pattern)
CREATE TABLE IF NOT EXISTS mapleapps.federated_users_by_email (
    email TEXT PRIMARY KEY,
    id UUID,
    first_name TEXT,
    last_name TEXT,
    name TEXT,
    lexical_name TEXT,
    role TINYINT,
    status TINYINT,
    timezone TEXT,
    created_at TIMESTAMP,
    modified_at TIMESTAMP,
    -- Store less frequently accessed data as JSON
    profile_data TEXT,  -- JSON with address, shipping info, etc.
    security_data TEXT, -- JSON with E2EE fields
    metadata TEXT      -- JSON with tracking fields
);

-- Table 3: Lookup by verification codes with TTL support
CREATE TABLE IF NOT EXISTS mapleapps.users_by_verification_code (
   code TEXT,
    code_type TEXT, -- 'email_verification' or 'password_reset'
    user_id UUID,
    email TEXT,
    created_at TIMESTAMP,
    expires_at TIMESTAMP,
    PRIMARY KEY (code, code_type)
);
