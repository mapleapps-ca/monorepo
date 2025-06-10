-- Main collections table with ALL Collection struct fields
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_id (
    -- Identifiers
    id UUID PRIMARY KEY,
    owner_id UUID,

    -- Encryption and Content Details
    encrypted_name TEXT,
    collection_type TEXT, -- "folder" or "album"
    encrypted_collection_key TEXT, -- Serialized keys.EncryptedCollectionKey JSON

    -- Sharing - REMOVED: members field (now normalized in separate table)
    -- members TEXT, -- We remove this to avoid JSON complexity

    -- Hierarchical structure fields
    parent_id UUID, -- Can be null for root collections
    ancestor_ids TEXT, -- Serialized []UUID JSON array for efficient querying

    -- Ownership, timestamps and conflict resolution
    created_at TIMESTAMP,
    created_by_user_id UUID,
    modified_at TIMESTAMP,
    modified_by_user_id UUID,
    version BIGINT, -- uint64 maps to BIGINT

    -- State management
    state TEXT, -- "active", "deleted", "archived"
    tombstone_version BIGINT,
    tombstone_expiry TIMESTAMP
);
