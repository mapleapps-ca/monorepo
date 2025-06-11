-- For queries like: "Show me only collections I OWN" or "Show me only collections SHARED with me"
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id (
    user_id UUID,
    access_type TEXT,
    modified_at TIMESTAMP,
    collection_id UUID,
    permission_level TEXT,
    state TEXT,
    PRIMARY KEY ((user_id, access_type), modified_at, collection_id)
) WITH CLUSTERING ORDER BY (modified_at DESC, collection_id ASC);
