CREATE TABLE IF NOT EXISTS mapleapps.iam_federated_users_by_email (
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
    profile_data TEXT,
    security_data TEXT,
    metadata TEXT,
    user_plan TEXT,
    storage_limit_bytes BIGINT,
    storage_used_bytes BIGINT
);
