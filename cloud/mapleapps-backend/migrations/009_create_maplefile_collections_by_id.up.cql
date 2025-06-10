-- Main collections table with ALL Collection struct fields
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_id (
    -- Identifiers
    id UUID PRIMARY KEY,
    owner_id UUID,

    -- Encryption and Content Details
    encrypted_name TEXT,
    collection_type TEXT,
    encrypted_collection_key TEXT,

    -- Hierarchical structure fields
    parent_id UUID,
    ancestor_ids TEXT, -- JSON array of UUIDs

    -- Ownership, timestamps and conflict resolution
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
