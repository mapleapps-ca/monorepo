-- Normalized members table with proper Cassandra naming
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collection_members_by_collection_id_and_recipient_id (
    collection_id UUID,
    recipient_id UUID,
    member_id UUID,
    recipient_email TEXT,
    granted_by_id UUID,
    encrypted_collection_key BLOB,
    permission_level TEXT,
    created_at TIMESTAMP,
    is_inherited BOOLEAN,
    inherited_from_id UUID,
    PRIMARY KEY ((collection_id), recipient_id)
);
