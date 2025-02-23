package cauth

import (
	"encoding/json"
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
	UUID      string    `db:"uuid"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	UserUUID             string    `db:"user_uuid"`
	ImpersonatedUserUUID *string   `db:"impersonated_user_uuid"`
	Token                []byte    `db:"token"`
	ExpiresAt            time.Time `db:"expires_at"`
}

func (s *Session) CurrentUserID() string {
	if s.ImpersonatedUserUUID != nil {
		return *s.ImpersonatedUserUUID
	}
	return s.UserUUID
}

func (s *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		UUID      string    `json:"uuid"`
		CreatedAt time.Time `json:"created_at"`
		UserUUID  string    `json:"user_uuid"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		UUID:      s.UUID,
		CreatedAt: s.CreatedAt,
		UserUUID:  s.CurrentUserID(),
		ExpiresAt: s.ExpiresAt,
	})
}
