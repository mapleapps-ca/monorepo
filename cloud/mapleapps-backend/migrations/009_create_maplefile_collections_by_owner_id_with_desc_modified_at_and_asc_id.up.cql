-- 2. COLLECTIONS BY OWNER (Enhanced for root filtering)
-- Handles: GetAllByUserID, FindRootCollections (with parent_id filtering)
-- Query pattern: "What collection IDs does this user own?"
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_owner_id_with_desc_modified_at_and_asc_id (
    owner_id UUID,
    modified_at TIMESTAMP,
    collection_id UUID,
    state TEXT,      -- For server-side state filtering
    parent_id UUID,  -- For root collection filtering (null = root)
    created_at TIMESTAMP, -- For root collections ordering
    PRIMARY KEY ((owner_id), modified_at, collection_id)
) WITH CLUSTERING ORDER BY (modified_at DESC, collection_id ASC);
