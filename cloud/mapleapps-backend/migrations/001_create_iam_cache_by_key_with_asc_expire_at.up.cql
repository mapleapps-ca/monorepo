CREATE TABLE IF NOT EXISTS pkg_cache_by_key_with_asc_expire_at (
  partition text,         -- arbitrary or static value since we're caching by keys
  key text,
  value blob,
  expires_at timestamp,
  PRIMARY KEY ((partition), expires_at, key)
) WITH CLUSTERING ORDER BY (expires_at ASC);
