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
