-- User access table (owners + members) with proper Cassandra naming
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id (
    user_id UUID,
    modified_at TIMESTAMP,
    collection_id UUID,
    access_type TEXT, -- 'owner' or 'member'
    permission_level TEXT, -- null for owners, actual permission for members
    state TEXT,
    PRIMARY KEY ((user_id), modified_at, collection_id)
) WITH CLUSTERING ORDER BY (modified_at DESC, collection_id ASC);
