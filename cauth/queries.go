package cauth

import (
	"context"
	"database/sql"
	"time"

	"github.com/gocopper/copper/csql"
)

// ErrNotFound is returned when a model does not exist in the repository
var ErrNotFound = sql.ErrNoRows

// NewQueries instantiates and returns Queries.
func NewQueries(querier csql.Querier) *Queries {
	return &Queries{querier: querier}
}

// Queries holds the SQL queries used by cauth
type Queries struct {
	querier csql.Querier
}

// GetUserByUsername queries the users table for a user with the given username.
func (q *Queries) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	const query = `select * from cauth_users where username=?`

	var user User

	err := q.querier.Get(ctx, &user, query, username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail queries the users table for a user with the given email.
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const query = `select * from cauth_users where email=?`

	var user User

	err := q.querier.Get(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// InsertUser creates the given user in cauth_users.
func (q *Queries) InsertUser(ctx context.Context, user *User) error {
	const query = `
	INSERT INTO cauth_users (uuid, created_at, updated_at, email, username, password, password_reset_token)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	RETURNING *`

	var now = time.Now()

	return q.querier.Get(ctx, user, query,
		user.UUID,
		now,
		now,
		user.Email,
		user.Username,
		user.Password,
		user.PasswordResetToken,
	)
}

// UpdateUser updates the given user in cauth_users.
func (q *Queries) UpdateUser(ctx context.Context, user *User) error {
	const query = `
	UPDATE cauth_users SET updated_at=?, password=?, password_reset_token=?
	WHERE uuid=?`

	_, err := q.querier.Exec(ctx, query,
		user.UpdatedAt,
		user.Password,
		user.PasswordResetToken,
		user.UUID,
	)
	return err
}

// GetSession queries the sessions table for a session with the given uuid.
func (q *Queries) GetSession(ctx context.Context, uuid string) (*Session, error) {
	const query = `select * from cauth_sessions where uuid=?`

	var session Session

	err := q.querier.Get(ctx, &session, query, uuid)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// InsertSession creates a new session in cauth_sessions
func (q *Queries) InsertSession(ctx context.Context, session *Session) error {
	const query = `
	INSERT INTO cauth_sessions (uuid, created_at, user_uuid, token, expires_at)
	VALUES (?, ?, ?, ?, ?)
	RETURNING *`

	return q.querier.Get(ctx, session, query,
		session.UUID,
		time.Now(),
		session.UserUUID,
		session.Token,
		session.ExpiresAt,
	)
}

// UpdateSession updates the given session in cauth_sessions.
func (q *Queries) UpdateSession(ctx context.Context, session *Session) error {
	const query = `
	UPDATE cauth_sessions SET expires_at=?
	WHERE uuid=?`

	_, err := q.querier.Exec(ctx, query,
		session.ExpiresAt,
		session.UUID,
	)
	return err
}
