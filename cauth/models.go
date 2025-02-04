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

	Email    string `db:"email" json:"email"`
	Password []byte `db:"password" json:"-"`

	EmailVerifiedAt           *time.Time `db:"email_verified_at" json:"-"`
	VerificationCode          *string    `db:"verification_code" json:"-"`
	VerificationCodeExpiresAt *time.Time `db:"verification_code_expires_at" json:"-"`
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
