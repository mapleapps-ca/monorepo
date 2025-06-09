CREATE TABLE IF NOT EXISTS mapleapps.collections_by_id (
    -- Identifiers
    id UUID PRIMARY KEY,
    owner_id UUID,

    -- Encryption and Content Details (serialized as JSON)
    encrypted_name TEXT,
    collection_type TEXT, -- "folder" or "album"
    encrypted_collection_key TEXT, -- Serialized keys.EncryptedCollectionKey JSON

    -- Sharing (serialized as JSON)
    members TEXT, -- Serialized []CollectionMembership JSON

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
