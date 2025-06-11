CREATE TABLE IF NOT EXISTS mapleapps.maplefile_files_by_id (
    -- Identifiers
    id UUID PRIMARY KEY,
    collection_id UUID,
    owner_id UUID,

    -- Encryption and Content Details
    encrypted_metadata TEXT,
    encrypted_file_key TEXT, -- JSON serialized
    encryption_version TEXT,
    encrypted_hash TEXT,

    -- File Storage Details
    encrypted_file_object_key TEXT,
    encrypted_file_size_in_bytes BIGINT,

    -- Thumbnail Storage Details
    encrypted_thumbnail_object_key TEXT,
    encrypted_thumbnail_size_in_bytes BIGINT,

    -- Timestamps and versioning
    created_at TIMESTAMP,
    created_by_user_id UUID,
    modified_at TIMESTAMP,
    modified_by_user_id UUID,
    version BIGINT,

    -- State management
    state TEXT,
    tombstone_version BIGINT,
    tombstone_expiry TIMESTAMP
);
