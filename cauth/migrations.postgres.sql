-- +migrate Up
create table if not exists cauth_users
(
    uuid                    text primary key,
    created_at              timestamp with time zone not null,
    updated_at              timestamp with time zone not null,
    email                   text unique,
    email_verified          boolean not null default false,
    email_verification_code bytea not null,
    username                text unique,
    password                bytea,
    password_reset_token    bytea
);

create table if not exists cauth_sessions
(
    uuid       text primary key,
    created_at timestamp with time zone not null,
    user_uuid  text not null,
    token      bytea not null,
    expires_at timestamp with time zone not null
);

-- +migrate Down
drop table if exists cauth_sessions;
drop table if exists cauth_users;
