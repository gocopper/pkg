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

// GetUserByUUID queries the users table for a user with the given uuid.
func (q *Queries) GetUserByUUID(ctx context.Context, uuid string) (*User, error) {
	const query = `select * from cauth_users where uuid=?`

	var user User

	err := q.querier.Get(ctx, &user, query, uuid)
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
	INSERT INTO cauth_users (uuid, created_at, updated_at, email, password, email_verified_at, verification_code, verification_code_expires_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING *`

	var now = time.Now()

	return q.querier.Get(ctx, user, query,
		user.UUID,
		now,
		now,
		user.Email,
		user.Password,
		user.EmailVerifiedAt,
		user.VerificationCode,
		user.VerificationCodeExpiresAt,
	)
}

// UpdateUser updates the given user in cauth_users.
func (q *Queries) UpdateUser(ctx context.Context, user *User) error {
	const query = `
	UPDATE cauth_users SET updated_at=?, password=?, email_verified_at=?, verification_code=?, verification_code_expires_at=?
	WHERE uuid=?`

	_, err := q.querier.Exec(ctx, query,
		user.UpdatedAt,
		user.Password,
		user.EmailVerifiedAt,
		user.VerificationCode,
		user.VerificationCodeExpiresAt,
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
	INSERT INTO cauth_sessions (uuid, created_at, updated_at, user_uuid, token, expires_at)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING *`

	return q.querier.Get(ctx, session, query,
		session.UUID,
		session.CreatedAt,
		session.UpdatedAt,
		session.UserUUID,
		session.Token,
		session.ExpiresAt,
	)
}

// UpdateSession updates the given session in cauth_sessions.
func (q *Queries) UpdateSession(ctx context.Context, session *Session) error {
	const query = `
	UPDATE cauth_sessions SET updated_at=?, expires_at=?, impersonated_user_uuid=?
	WHERE uuid=?`

	_, err := q.querier.Exec(ctx, query,
		session.UpdatedAt,
		session.ExpiresAt,
		session.ImpersonatedUserUUID,
		session.UUID,
	)
	return err
}
