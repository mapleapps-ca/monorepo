-- 6. COLLECTIONS BY ANCESTOR (for hierarchy traversal)
-- Handles: FindDescendants efficiently
-- Query pattern: "What collection IDs are descendants of this ancestor?"
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id (
    ancestor_id UUID,
    depth INT, -- How many levels deep (1 = direct child, 2 = grandchild, etc.)
    collection_id UUID,
    state TEXT, -- For filtering out deleted descendants
    PRIMARY KEY ((ancestor_id), depth, collection_id)
) WITH CLUSTERING ORDER BY (depth ASC, collection_id ASC);
