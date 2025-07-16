CREATE TABLE IF NOT EXISTS mapleapps.maplefile_storage_usage_events_by_user_id_and_event_day_with_asc_event_time_and_asc_file_id (
    user_id UUID,
    event_day DATE,
    event_time TIMESTAMP,
    file_id UUID,
    file_size BIGINT,
    operation TEXT, /* 'add' or 'remove' */
    PRIMARY KEY ((user_id, event_day), event_time, file_id)
) WITH CLUSTERING ORDER BY (event_time ASC);
