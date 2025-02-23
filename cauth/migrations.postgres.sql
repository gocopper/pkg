-- +migrate Up
create table if not exists cauth_users
(
    uuid                         text primary key,
    created_at                   timestamp with time zone not null,
    updated_at                   timestamp with time zone not null,
    email                        text unique,
    email_verified_at            timestamp with time zone,
    verification_code            text,
    verification_code_expires_at timestamp with time zone,
    password                     bytea
);

create table if not exists cauth_sessions
(
    uuid                   text primary key,
    created_at             timestamp with time zone not null,
    updated_at             timestamp with time zone not null,
    user_uuid              text                     not null,
    impersonated_user_uuid text,
    token                  bytea                    not null,
    expires_at             timestamp with time zone not null
);

-- +migrate Down
drop table if exists cauth_sessions;
drop table if exists cauth_users;
