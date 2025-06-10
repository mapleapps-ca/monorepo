-- Efficient hierarchy traversal (eliminates recursive queries)
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_ancestor (
    ancestor_id UUID,
    depth INT,
    collection_id UUID,
    state TEXT,
    PRIMARY KEY ((ancestor_id), depth, collection_id)
) WITH CLUSTERING ORDER BY (depth ASC, collection_id ASC);
