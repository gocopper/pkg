package cauth

import (
	"context"
	"errors"
	"github.com/gocopper/pkg/cvars"
	"net/http"
	"strconv"
	"strings"
	"text/template"
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

	// ErrVerificationCodeExpired is returned when the verification code has expired.
	ErrVerificationCodeExpired = errors.New("verification code expired")
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
	Email    string  `json:"email"`
	Password *string `json:"password"`
}

// LoginParams hold the params needed to login a user.
type LoginParams struct {
	Email            string  `json:"email"`
	Password         *string `json:"password"`
	VerificationCode *string `json:"verification_code"`
}

// VerifyEmailParams hold the params needed to verify an email.
type VerifyEmailParams struct {
	Email            string `json:"email"`
	VerificationCode string `json:"verification_code"`
}

type ResetPasswordParams struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	VerificationCode string `json:"verification_code"`
}

func (s *Svc) ResendVerificationCode(ctx context.Context, email string) error {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, ErrNotFound) {
		return ErrInvalidCredentials
	} else if err != nil {
		return cerrors.New(err, "failed to get user by email", map[string]interface{}{
			"email": email,
		})
	}

	return s.sendVerificationCodeEmail(ctx, user)
}

func (s *Svc) sendVerificationCodeEmail(ctx context.Context, user *User) error {
	var emailBodySb strings.Builder

	user.UpdatedAt = time.Now()
	user.VerificationCode = cvars.Ptr(strconv.Itoa(int(crandom.GenerateRandomNumericalCode(s.config.VerificationCodeLen))))
	user.VerificationCodeExpiresAt = cvars.Ptr(time.Now().UTC().Add(time.Minute * 10))

	err := s.queries.UpdateUser(ctx, user)
	if err != nil {
		return cerrors.New(err, "failed to update user with new verification code", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	tmpl, err := template.New("email_verification_code").Parse(s.config.VerificationEmailBodyHTML)
	if err != nil {
		return cerrors.New(err, "failed to parse verification code email template", nil)
	}

	err = tmpl.Execute(&emailBodySb, map[string]string{
		"VerificationCode": *user.VerificationCode,
	})
	if err != nil {
		return cerrors.New(err, "failed to execute verification code email template", nil)
	}

	emailBody := emailBodySb.String()

	err = s.mailer.Send(ctx, cmailer.SendParams{
		From:     s.config.VerificationEmailFrom,
		To:       []string{user.Email},
		Subject:  s.config.VerificationEmailSubject,
		HTMLBody: &emailBody,
	})
	if err != nil {
		return cerrors.New(err, "failed to send verification code email", map[string]interface{}{
			"to": user.Email,
		})
	}

	return nil
}

func (s *Svc) ResetPassword(ctx context.Context, p ResetPasswordParams) error {
	user, err := s.queries.GetUserByEmail(ctx, p.Email)
	if err != nil && errors.Is(err, ErrNotFound) {
		return ErrInvalidCredentials
	} else if err != nil {
		return cerrors.New(err, "failed to get user by email", map[string]interface{}{
			"email": p.Email,
		})
	}

	if user.VerificationCodeExpiresAt == nil || time.Now().UTC().After(*user.VerificationCodeExpiresAt) {
		return ErrVerificationCodeExpired
	} else if user.VerificationCode == nil || *user.VerificationCode != p.VerificationCode {
		return ErrInvalidCredentials
	}

	hp, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return cerrors.New(err, "failed to hash password", nil)
	}

	user.UpdatedAt = time.Now()
	user.Password = hp

	err = s.queries.UpdateUser(ctx, user)
	if err != nil {
		return cerrors.New(err, "failed to update user", map[string]interface{}{
			"userUUID": user.UUID,
		})
	}

	return nil
}

// Signup creates a new user. If contact methods such as email or phone are provided, it will send verification
// codes so them. It creates a new session for this newly created user and returns that.
func (s *Svc) Signup(ctx context.Context, p SignupParams) (*SessionResult, error) {
	return s.signupWithEmail(ctx, p.Email, p.Password)
}

func (s *Svc) signupWithEmail(ctx context.Context, email string, password *string) (*SessionResult, error) {
	var newUser = false

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
		user.EmailVerifiedAt = nil

		err = s.queries.UpdateUser(ctx, user)
		if err != nil {
			return nil, cerrors.New(err, "failed to update user", nil)
		}
	} else if errors.Is(err, ErrNotFound) {
		newUser = true
		user = &User{
			UUID:      uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Email:     email,
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

	err = s.sendVerificationCodeEmail(ctx, user)
	if err != nil {
		return nil, cerrors.New(err, "failed to send verification code email", map[string]interface{}{
			"userID": user.UUID,
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

// Login logs in an existing user with the given credentials. If the login succeeds, it creates a new session
// and returns it.
func (s *Svc) Login(ctx context.Context, p LoginParams) (*SessionResult, error) {
	if p.Password != nil {
		return s.loginWithEmailPassword(ctx, p.Email, *p.Password)
	}

	if p.VerificationCode != nil {
		return s.loginWithEmailVerificationCode(ctx, p.Email, *p.VerificationCode)
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

	if user.VerificationCodeExpiresAt == nil || time.Now().UTC().After(*user.VerificationCodeExpiresAt) {
		return nil, ErrVerificationCodeExpired
	} else if user.VerificationCode == nil || *user.VerificationCode != p.VerificationCode {
		return nil, ErrInvalidCredentials
	}

	user.UpdatedAt = time.Now()
	user.EmailVerifiedAt = &user.UpdatedAt
	user.VerificationCodeExpiresAt = &user.UpdatedAt

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

// getSessionAndUserFromHTTPRequest gets the session and user from the request. It checks for the session uuid and
// token in the following places:
//  1. The Authorization header using basic auth where the username is the session uuid
//     and the password is the session token
//  2. SessionUUID and SessionToken cookies
//
// If the validation fails, ErrInvalidCredentials is returned.
func (s *Svc) getSessionAndUserFromHTTPRequest(_ context.Context, r *http.Request) (*Session, *User, error) {
	var (
		sessionUUID string
		plainToken  string
	)

	sessionUUIDCookie, err := r.Cookie("SessionUUID")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, nil, cerrors.New(err, "failed to get session uuid cookie", nil)
	}

	sessionTokenCookie, err := r.Cookie("SessionToken")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, nil, cerrors.New(err, "failed to get session token cookie", nil)
	}

	if sessionTokenCookie != nil && sessionUUIDCookie != nil {
		sessionUUID = sessionUUIDCookie.Value
		plainToken = sessionTokenCookie.Value
	}

	basicAuthUsername, basicAuthPass, ok := r.BasicAuth()
	if ok && basicAuthUsername != "" && basicAuthPass != "" {
		sessionUUID = basicAuthUsername
		plainToken = basicAuthPass
	}

	if sessionUUID == "" || plainToken == "" {
		return nil, nil, ErrInvalidCredentials
	}

	ok, session, err := s.ValidateSession(r.Context(), sessionUUID, plainToken)
	if err != nil {
		return nil, nil, cerrors.New(err, "failed to validate session", map[string]interface{}{
			"sessionUUID": sessionUUID,
		})
	}

	if !ok {
		return nil, nil, ErrInvalidCredentials
	}

	user, err := s.GetUserByUUID(r.Context(), session.UserUUID)
	if err != nil {
		return nil, nil, cerrors.New(err, "failed to get user by uuid", map[string]interface{}{
			"userUUID": session.UserUUID,
		})
	}

	return session, user, nil
}
