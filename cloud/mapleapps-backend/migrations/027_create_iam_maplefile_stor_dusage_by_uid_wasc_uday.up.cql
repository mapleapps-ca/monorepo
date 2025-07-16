CREATE TABLE IF NOT EXISTS mapleapps.maplefile_storage_daily_usage_by_user_id_with_asc_usage_day (
    user_id UUID,
    usage_day DATE,
    total_bytes BIGINT,
    total_add_bytes BIGINT,
    total_remove_bytes BIGINT,
    PRIMARY KEY ((user_id), usage_day)
) WITH CLUSTERING ORDER BY (usage_day ASC);
