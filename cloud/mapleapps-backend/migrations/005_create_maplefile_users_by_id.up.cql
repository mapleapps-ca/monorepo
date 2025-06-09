CREATE TABLE IF NOT EXISTS mapleapps.maplefile_users_by_id (
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
    profile_data TEXT,
    security_data TEXT,
    metadata TEXT
);
