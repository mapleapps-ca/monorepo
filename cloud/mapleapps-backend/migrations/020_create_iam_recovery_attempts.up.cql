-- Primary table for recovery attempts
CREATE TABLE IF NOT EXISTS mapleapps.iam_recovery_attempts (
    id UUID PRIMARY KEY,
    user_id UUID,
    email TEXT,
    method TEXT,
    ip_address TEXT,
    user_agent TEXT,
    status TEXT,
    failure_reason TEXT,
    challenge_id TEXT,
    attempted_at TIMESTAMP,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP
);

-- Index by user ID for history queries
CREATE TABLE IF NOT EXISTS mapleapps.iam_recovery_attempts_by_user_id (
    user_id UUID,
    attempted_at TIMESTAMP,
    id UUID,
    email TEXT,
    method TEXT,
    ip_address TEXT,
    user_agent TEXT,
    status TEXT,
    failure_reason TEXT,
    challenge_id TEXT,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP,
    PRIMARY KEY ((user_id), attempted_at, id)
) WITH CLUSTERING ORDER BY (attempted_at DESC, id ASC);

-- Index by email for rate limiting
CREATE TABLE IF NOT EXISTS mapleapps.iam_recovery_attempts_by_email (
    email TEXT,
    attempted_at TIMESTAMP,
    id UUID,
    user_id UUID,
    method TEXT,
    ip_address TEXT,
    user_agent TEXT,
    status TEXT,
    failure_reason TEXT,
    challenge_id TEXT,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP,
    PRIMARY KEY ((email), attempted_at, id)
) WITH CLUSTERING ORDER BY (attempted_at DESC, id ASC);
