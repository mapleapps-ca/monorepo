-- Index by user ID for history queries
CREATE TABLE IF NOT EXISTS mapleapps.iam_recovery_key_rotations_by_user_id (
    user_id UUID,
    rotated_at TIMESTAMP,
    id UUID,
    old_recovery_key_hash TEXT,
    new_recovery_key_hash TEXT,
    rotated_by UUID,
    reason TEXT,
    PRIMARY KEY ((user_id), rotated_at, id)
) WITH CLUSTERING ORDER BY (rotated_at DESC, id ASC);
