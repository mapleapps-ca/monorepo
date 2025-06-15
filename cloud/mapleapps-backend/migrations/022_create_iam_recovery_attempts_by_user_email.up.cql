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
