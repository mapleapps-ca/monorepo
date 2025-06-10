-- Normalized members table - stores individual CollectionMembership records
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collection_members (
    collection_id UUID,
    recipient_id UUID,
    member_id UUID, -- CollectionMembership.ID
    recipient_email TEXT,
    granted_by_id UUID,
    encrypted_collection_key BLOB,
    permission_level TEXT,
    created_at TIMESTAMP,
    is_inherited BOOLEAN,
    inherited_from_id UUID,
    PRIMARY KEY ((collection_id), recipient_id)
);
