CREATE TABLE IF NOT EXISTS pkg_cache_by_key_with_asc_expire_at (
    key text PRIMARY KEY,
    expires_at timestamp,
    value blob
);

-- Create an index on expires_at for efficient purging
CREATE INDEX IF NOT EXISTS idx_cache_expires_at ON pkg_cache_by_key_with_asc_expire_at (expires_at);
