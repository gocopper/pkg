-- +migrate Up
create table if not exists cauth_users
(
    uuid                 text primary key,
    created_at           datetime not null,
    updated_at           datetime not null,
    email                text unique,
    email_verified       integer  not null default 0,
    username             text unique,
    password             blob,
    password_reset_token blob
);

create table if not exists cauth_sessions
(
    uuid       text primary key,
    created_at datetime not null,
    user_uuid  text     not null,
    token      blob     not null,
    expires_at datetime not null
);

-- +migrate Down
drop table if exists cauth_users;
drop table if exists cauth_sessions;
