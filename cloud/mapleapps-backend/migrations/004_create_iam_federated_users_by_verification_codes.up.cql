CREATE TABLE IF NOT EXISTS mapleapps.iam_federated_users_by_verification_code (
    code TEXT,
    code_type TEXT,
    user_id UUID,
    email TEXT,
    created_at TIMESTAMP,
    expires_at TIMESTAMP,
    PRIMARY KEY (code, code_type)
);
