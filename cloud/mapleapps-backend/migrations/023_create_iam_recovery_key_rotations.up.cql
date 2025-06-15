-- Primary table for recovery key rotations
CREATE TABLE IF NOT EXISTS mapleapps.iam_recovery_key_rotations (
    id UUID PRIMARY KEY,
    user_id UUID,
    old_recovery_key_hash TEXT,
    new_recovery_key_hash TEXT,
    rotated_at TIMESTAMP,
    rotated_by UUID,
    reason TEXT
);
