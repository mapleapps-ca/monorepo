-- 5. COLLECTION ACCESS BY USER (Membership)
-- Handles: GetCollectionsSharedWithUser, CheckAccess, GetUserPermissionLevel
-- Query pattern: "What collection IDs can this user access and with what permissions?"
CREATE TABLE IF NOT EXISTS mapleapps.maplefile_collections_by_recipient_id_and_collection_id (
    recipient_id UUID,
    collection_id UUID,
    permission_level TEXT, -- For sync ordering of shared collections
    granted_at TIMESTAMP,  -- For ordering shared collections
    modified_at TIMESTAMP, -- For sync ordering of shared collections
    PRIMARY KEY ((recipient_id), collection_id)
);
