CREATE TABLE file_uploads (
    id                    TEXT PRIMARY KEY,
    local_path            TEXT NOT NULL,
    bucket                TEXT NOT NULL,
    version               TEXT NOT NULL,
    salt                  BLOB NOT NULL,
    encryption_algorithm  INTEGER NOT NULL,
    created_at_timestamp  DATETIME NOT NULL,
    uploaded_at_timestamp DATETIME NOT NULL,
    deleted_at_timestamp  DATETIME NOT NULL
);

CREATE TABLE file_uploads_history (
    request_id            TEXT PRIMARY KEY,
    id                    TEXT NOT NULL,
    local_path            TEXT NOT NULL,
    bucket                TEXT NOT NULL,
    version               TEXT NOT NULL,
    salt                  BLOB NOT NULL,
    encryption_algorithm  INTEGER NOT NULL,
    created_at_timestamp  DATETIME NOT NULL,
    uploaded_at_timestamp DATETIME NOT NULL,
    deleted_at_timestamp  DATETIME NOT NULL,
    change_type           INTEGER NOT NULL
);
