CREATE TABLE files (
    id                    TEXT PRIMARY KEY,
    key                   TEXT NOT NULL,
    local_path            TEXT NOT NULL,
    checksum              INTEGER NOT NULL,
    etag                  TEXT NOT NULL,
    bucket                TEXT NOT NULL,
    version               TEXT NOT NULL,
    salt                  BLOB NOT NULL,
    encryption_algorithm  INTEGER NOT NULL,
    key_params            TEXT NOT NULL,
    created_at_timestamp  DATETIME NOT NULL
);
