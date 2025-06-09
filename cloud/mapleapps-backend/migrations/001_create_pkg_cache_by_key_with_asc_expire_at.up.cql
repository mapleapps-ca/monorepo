CREATE TABLE IF NOT EXISTS mapleapps.pkg_cache_by_key_with_asc_expire_at (
    key text PRIMARY KEY,
    expires_at timestamp,
    value blob
);
