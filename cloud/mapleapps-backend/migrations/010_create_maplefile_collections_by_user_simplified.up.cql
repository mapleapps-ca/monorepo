-- Unified user access table (replaces multiple existing tables)
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_user_simplified (
    user_id UUID,
    collection_id UUID,
    access_type TEXT, -- 'owner' or 'member'
    permission_level TEXT, -- null for owners, actual permission for members
    modified_at TIMESTAMP,
    state TEXT,
    parent_id UUID,
    created_at TIMESTAMP,
    PRIMARY KEY ((user_id), modified_at, collection_id)
) WITH CLUSTERING ORDER BY (modified_at DESC, collection_id ASC);
