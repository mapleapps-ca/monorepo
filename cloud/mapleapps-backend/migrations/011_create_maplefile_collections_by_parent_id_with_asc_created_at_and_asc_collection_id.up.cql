-- Efficient parent/child queries (eliminates ALLOW FILTERING) with proper Cassandra naming
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_parent_id_with_asc_created_at_and_asc_collection_id (
    parent_id UUID,
    created_at TIMESTAMP,
    collection_id UUID,
    owner_id UUID, -- Denormalized to avoid two-step queries
    state TEXT,
    PRIMARY KEY ((parent_id), created_at, collection_id)
) WITH CLUSTERING ORDER BY (created_at ASC, collection_id ASC);
