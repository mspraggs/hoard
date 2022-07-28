CREATE SCHEMA IF NOT EXISTS files;

CREATE TABLE files.files (
    id                    TEXT PRIMARY KEY,
    key                   TEXT NOT NULL,
    local_path            TEXT NOT NULL,
    checksum              INTEGER NOT NULL,
    etag                  TEXT NOT NULL,
    bucket                TEXT NOT NULL,
    version               TEXT NOT NULL,
    created_at_timestamp  TIMESTAMPTZ NOT NULL
);

CREATE INDEX files_local_path_idx ON files.files (local_path);
