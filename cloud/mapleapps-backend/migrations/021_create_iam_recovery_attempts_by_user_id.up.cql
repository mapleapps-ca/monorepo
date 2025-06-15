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
