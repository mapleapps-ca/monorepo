REATE TABLE IF NOT EXISTS mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id (
    owner_id UUID,
    modified_at TIMESTAMP,
    file_id UUID,

    -- ALL FIELDS DUPLICATED
    collection_id UUID,
    encrypted_metadata TEXT,
    encrypted_file_key TEXT,
    encryption_version TEXT,
    encrypted_hash TEXT,
    encrypted_file_object_key TEXT,
    encrypted_file_size_in_bytes BIGINT,
    encrypted_thumbnail_object_key TEXT,
    encrypted_thumbnail_size_in_bytes BIGINT,
    created_at TIMESTAMP,
    created_by_user_id UUID,
    modified_by_user_id UUID,
    version BIGINT,
    state TEXT,
    tombstone_version BIGINT,
    tombstone_expiry TIMESTAMP,

    PRIMARY KEY ((owner_id), modified_at, file_id)
) WITH CLUSTERING ORDER BY (modified_at DESC, file_id ASC);
