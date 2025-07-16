CREATE TABLE IF NOT EXISTS mapleapps.maplefile_storage_usage_events_by_user_id_and_event_day_with_asc_event_time (
    user_id UUID,
    event_day DATE,
    event_time TIMESTAMP,
    file_size BIGINT,
    operation TEXT,
    PRIMARY KEY ((user_id, event_day), event_time)
) WITH CLUSTERING ORDER BY (event_time ASC);
