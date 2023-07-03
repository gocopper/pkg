package cauth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gocopper/pkg/cmailer"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/pkg/crandom"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when a credential check fails. This usually happens during the login process.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserAlreadyExists is returned when a user already exists during the signup process.
	ErrUserAlreadyExists = errors.New("user already exists")
)

// NewSvc instantiates and returns a new Svc.
func NewSvc(queries *Queries, mailer cmailer.Mailer, config Config) (*Svc, error) {
	return &Svc{
		queries: queries,
		mailer:  mailer,
		config:  config,
	}, nil
}

// Svc provides methods to manage users and sessions.
type Svc struct {
	queries *Queries
	mailer  cmailer.Mailer
	config  Config
}

// SessionResult is usually used when a new session is created. It holds the plain session token that can be used
// to authenticate the session as well as other related entities such as the user and session.
type SessionResult struct {
	User              *User    `json:"user"`
	Session           *Session `json:"session"`
	PlainSessionToken string   `json:"plain_session_token"`
	NewUser           bool     `json:"new_user"`
}

// SignupParams hold the params needed to signup a new user.
type SignupParams struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Password *string `json:"password"`
}

// LoginParams hold the params needed to login a user.
type LoginParams struct {
	Email            *string `json:"email"`
	Username         *string `json:"username"`
	Password         *string `json:"password"`
	VerificationCode *string `json:"verification_code"`
}

// VerifyEmailParams hold the params needed to verify an email.
type VerifyEmailParams struct {
	Email            string `json:"email"`
	VerificationCode string `json:"verification_code"`
}

// Signup creates a new user. If contact methods such as email or phone are provided, it will send verification
// codes so them. It creates a new session for this newly created user and returns that.
func (s *Svc) Signup(ctx context.Context, p SignupParams) (*SessionResult, error) {
	if p.Username != nil && p.Password != nil {
		return s.signupWithUsernamePassword(ctx, *p.Username, *p.Password)
	}

	if p.Email != nil {
		return s.signupWithEmail(ctx, *p.Email, p.Password)
	}

	return nil, errors.New("invalid signup params")
}

func (s *Svc) signupWithEmail(ctx context.Context, email string, password *string) (*SessionResult, error) {
	var (
		newUser          = false
		verificationCode = strconv.Itoa(int(crandom.GenerateRandomNumericalCode(s.config.VerificationCodeLen)))
	)

	hashedVerificationCode, err := bcrypt.GenerateFromPassword([]byte(verificationCode), bcrypt.DefaultCost)
	if err != nil {
		return nil, cerrors.New(err, "failed to hash verification code", nil)
	}

	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, cerrors.New(err, "failed to get user by email", map[string]interface{}{
			"email": email,
		})
	} else if err == nil && len(user.Password) > 0 {
		// User should not be able to signup with an email that already exists
		return nil, ErrUserAlreadyExists
	} else if err == nil && len(user.Password) == 0 {
		user.UpdatedAt = time.Now()
		user.EmailVerificationCode = hashedVerificationCode
		user.EmailVerified = false

		err = s.queries.UpdateUser(ctx, user)
		if err != nil {
			return nil, cerrors.New(err, "failed to update user", nil)
		}
	} else if errors.Is(err, ErrNotFound) {
		newUser = true
		user = &User{
			UUID:                  uuid.New().String(),
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			Email:                 &email,
			EmailVerificationCode: hashedVerificationCode,
			EmailVerified:         false,
		}

		if password != nil {
			hp, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
			if err != nil {
				return nil, cerrors.New(err, "failed to hash password", nil)
			}
			user.Password = hp
		}

		err = s.queries.InsertUser(ctx, user)
		if err != nil {
			return nil, cerrors.New(err, "failed to insert user", nil)
		}
	}

	emailBody := fmt.Sprintf("Your verification code is %s", verificationCode)

	err = s.mailer.Send(ctx, cmailer.SendParams{
		From:      s.config.VerificationEmailFrom,
		To:        []string{email},
		Subject:   s.config.VerificationEmailSubject,
		PlainBody: &emailBody,
	})
	if err != nil {
		return nil, cerrors.New(err, "failed to send verification code email", map[string]interface{}{
			"to": email,
		})
	}

	if password == nil {
		// If no password is provided, we don't create a session
		// because the user will login with the verification code
		return &SessionResult{
			User:    user,
			NewUser: newUser,
		}, nil
	}

	session, plainSessionToken, err := s.createSession(ctx, user.UUID)
	if err != nil {
		return nil, cerrors.New(err, "failed to create session", nil)
	}

	return &SessionResult{
		User:              user,
		Session:           session,
		PlainSessionToken: plainSessionToken,
		NewUser:           newUser,
	}, nil
}

func (s *Svc) signupWithUsernamePassword(ctx context.Context, username, password string) (*SessionResult, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, cerrors.New(err, "failed to hash password", nil)
	}

	user := &User{
		UUID:      uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Username:  &username,
		Password:  hashedPassword,
	}

	err = s.queries.InsertUser(ctx, user)
	if err != nil {
		return nil, cerrors.New(err, "failed to save user", nil)
	}

	session, plainSessionToken, err := s.createSession(ctx, user.UUID)
	if err != nil {
		return nil, cerrors.New(err, "failed to create a session", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	return &SessionResult{
		User:              user,
		Session:           session,
		PlainSessionToken: plainSessionToken,
		NewUser:           true,
	}, nil
}

// Login logs in an existing user with the given credentials. If the login succeeds, it creates a new session
// and returns it.
func (s *Svc) Login(ctx context.Context, p LoginParams) (*SessionResult, error) {
	if p.Username != nil && p.Password != nil {
		return s.loginWithUsernamePassword(ctx, *p.Username, *p.Password)
	}

	if p.Email != nil && p.Password != nil {
		return s.loginWithEmailPassword(ctx, *p.Email, *p.Password)
	}

	if p.Email != nil && p.VerificationCode != nil {
		return s.loginWithEmailVerificationCode(ctx, *p.Email, *p.VerificationCode)
	}

	return nil, cerrors.New(nil, "invalid login params", nil)
}

// VerifyEmail verifies the email of a user with the given verification code. If the verification succeeds,
// it updates the user's email verification status and returns the user.
func (s *Svc) VerifyEmail(ctx context.Context, p VerifyEmailParams) (*User, error) {
	user, err := s.queries.GetUserByEmail(ctx, p.Email)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, cerrors.New(err, "failed to get user by email", map[string]interface{}{
			"email": p.Email,
		})
	}

	err = bcrypt.CompareHashAndPassword(user.EmailVerificationCode, []byte(p.VerificationCode))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user.EmailVerified = true

	err = s.queries.UpdateUser(ctx, user)
	if err != nil {
		return nil, cerrors.New(err, "failed to update user", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	return user, nil
}

func (s *Svc) loginWithEmailVerificationCode(ctx context.Context, email, code string) (*SessionResult, error) {
	user, err := s.VerifyEmail(ctx, VerifyEmailParams{
		Email:            email,
		VerificationCode: code,
	})
	if err != nil {
		return nil, cerrors.New(err, "failed to verify email", map[string]interface{}{
			"email": email,
		})
	}

	if len(user.Password) > 0 {
		return nil, cerrors.New(nil, "user cannot login with verification code because they have a password", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	session, plainSessionToken, err := s.createSession(ctx, user.UUID)
	if err != nil {
		return nil, cerrors.New(err, "failed to create session", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	return &SessionResult{
		User:              user,
		Session:           session,
		PlainSessionToken: plainSessionToken,
	}, nil

}

func (s *Svc) loginWithEmailPassword(ctx context.Context, email, password string) (*SessionResult, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, cerrors.New(err, "failed to get user by email", map[string]interface{}{
			"email": email,
		})
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	session, plainSessionToken, err := s.createSession(ctx, user.UUID)
	if err != nil {
		return nil, cerrors.New(err, "failed to create session", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	return &SessionResult{
		User:              user,
		Session:           session,
		PlainSessionToken: plainSessionToken,
	}, nil
}

func (s *Svc) loginWithUsernamePassword(ctx context.Context, username, password string) (*SessionResult, error) {
	user, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, cerrors.New(err, "failed to get user by username", map[string]interface{}{
			"username": username,
		})
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	session, plainSessionToken, err := s.createSession(ctx, user.UUID)
	if err != nil {
		return nil, cerrors.New(err, "failed to create session", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	return &SessionResult{
		User:              user,
		Session:           session,
		PlainSessionToken: plainSessionToken,
	}, nil
}

func (s *Svc) createSession(ctx context.Context, userUUID string) (*Session, string, error) {
	const tokenLen = 72

	plainToken := crandom.GenerateRandomString(tokenLen)

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", cerrors.New(err, "failed to hash session token", nil)
	}

	session := &Session{
		UUID:      uuid.New().String(),
		CreatedAt: time.Now(),
		UserUUID:  userUUID,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	err = s.queries.InsertSession(ctx, session)
	if err != nil {
		return nil, "", cerrors.New(err, "failed to create a new session", nil)
	}

	return session, plainToken, nil
}

// ValidateSession validates whether the provided plainToken is valid for the session identified by the given
// sessionUUID.
func (s *Svc) ValidateSession(ctx context.Context, sessionUUID, plainToken string) (bool, *Session, error) {
	session, err := s.queries.GetSession(ctx, sessionUUID)
	if err != nil && errors.Is(err, ErrNotFound) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, cerrors.New(err, "failed to get session", map[string]interface{}{
			"sessionUUID": sessionUUID,
		})
	}

	err = bcrypt.CompareHashAndPassword(session.Token, []byte(plainToken))
	if err != nil {
		return false, nil, nil
	}

	return true, session, nil
}

// GetUserByUUID returns the user identified by the given userUUID.
func (s *Svc) GetUserByUUID(ctx context.Context, userUUID string) (*User, error) {
	return s.queries.GetUserByUUID(ctx, userUUID)
}

// Logout invalidates the session identified by the given sessionUUID.
func (s *Svc) Logout(ctx context.Context, sessionUUID string) error {
	session, err := s.queries.GetSession(ctx, sessionUUID)
	if err != nil {
		return cerrors.New(err, "failed to get session", map[string]interface{}{
			"sessionUUID": sessionUUID,
		})
	}

	session.ExpiresAt = time.Now()

	err = s.queries.UpdateSession(ctx, session)
	if err != nil {
		return cerrors.New(err, "failed to save session", map[string]interface{}{
			"sessionUUID": sessionUUID,
		})
	}

	return nil
}
