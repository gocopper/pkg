package cauth

import (
	"time"
)

// User represents a user who has created an account. This model stores their login credentials
// as well as their metadata.
type User struct {
	UUID      string    `db:"uuid" json:"uuid"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`

	Email    *string `db:"email" json:"email,omitempty"`
	Username *string `db:"username" json:"username,omitempty"`

	Password           []byte `db:"password" json:"-"`
	PasswordResetToken []byte `db:"password_reset_token" json:"-"`

	EmailVerified         bool   `db:"email_verified" json:"email_verified"`
	EmailVerificationCode []byte `db:"email_verification_code" json:"-"`
}

// Session represents a single logged-in session that a user is able create after providing valid
// login credentials.
type Session struct {
	UUID      string    `db:"uuid" json:"uuid"`
	CreatedAt time.Time `db:"created_at" json:"-"`

	UserUUID  string    `db:"user_uuid" json:"user_uuid"`
	Token     []byte    `db:"token" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}
