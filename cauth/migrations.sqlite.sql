-- +migrate Up
CREATE TABLE IF NOT EXISTS cauth_users
(
    uuid                         TEXT PRIMARY KEY,
    created_at                   DATETIME NOT NULL,
    updated_at                   DATETIME NOT NULL,
    email                        TEXT UNIQUE,
    email_verified_at            DATETIME,
    verification_code            TEXT,
    verification_code_expires_at DATETIME,
    password                     BLOB
);

CREATE TABLE IF NOT EXISTS cauth_sessions
(
    uuid       TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    user_uuid  TEXT     NOT NULL,
    token      BLOB     NOT NULL,
    expires_at DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS cauth_sessions;
DROP TABLE IF EXISTS cauth_users;