package cauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/copper/chttp"
	"github.com/gocopper/copper/clogger"
)

type ctxKey string

const (
	ctxKeySession = ctxKey("cauth/session")
	ctxKeyUser    = ctxKey("cauth/user")
)

// NewVerifySessionMiddleware instantiates and creates a new VerifySessionMiddleware
func NewVerifySessionMiddleware(auth *Svc, rw *chttp.HTMLReaderWriter, logger clogger.Logger) *VerifySessionMiddleware {
	return &VerifySessionMiddleware{
		auth:   auth,
		rw:     rw,
		logger: logger,
	}
}

func NewSetSessionIfAnyMiddleware(auth *Svc, rw *chttp.HTMLReaderWriter, logger clogger.Logger) *SetSessionIfAnyMiddleware {
	return &SetSessionIfAnyMiddleware{
		auth:   auth,
		rw:     rw,
		logger: logger,
	}
}

// VerifySessionMiddleware is a middleware that checks for a valid session uuid and token in:
//  1. The Authorization header using basic auth where the username is the session uuid
//     and the password is the session token
//  2. SessionUUID and SessionToken cookies
//
// If the session is present, it is validated, saved in the request ctx along with the user,
// and the next handler is called. If the session is invalid, an unauthorized response is sent
// back.
// To ensure verified session, use in conjunction with VerifySessionMiddleware.
type VerifySessionMiddleware struct {
	auth   *Svc
	rw     *chttp.HTMLReaderWriter
	logger clogger.Logger
}

// SetSessionIfAnyMiddleware works the same way as VerifySessionMiddleware, but does not return an
// error if no session is found. Instead, it just calls the next handler.
// This is useful in conjunction with the HasVerifiedSession function.
type SetSessionIfAnyMiddleware struct {
	auth   *Svc
	rw     *chttp.HTMLReaderWriter
	logger clogger.Logger
}

// Handle implements the middleware for VerifySessionMiddleware.
func (mw *VerifySessionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, user, err := mw.auth.getSessionAndUserFromHTTPRequest(r.Context(), r)
		if err != nil && errors.Is(err, ErrInvalidCredentials) {
			mw.rw.Unauthorized(w, r)
			return
		} else if err != nil {
			mw.rw.WriteHTMLError(w, r, cerrors.New(err, "failed to get session and user from http request", nil))
			return
		}

		ctxWithUser := context.WithValue(r.Context(), ctxKeyUser, user)
		ctxWithUserAndSession := context.WithValue(ctxWithUser, ctxKeySession, session)

		next.ServeHTTP(w, r.WithContext(ctxWithUserAndSession))
	})
}

func (mw *SetSessionIfAnyMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, user, err := mw.auth.getSessionAndUserFromHTTPRequest(r.Context(), r)
		if err != nil && errors.Is(err, ErrInvalidCredentials) {
			next.ServeHTTP(w, r)

			return
		} else if err != nil {
			mw.rw.WriteHTMLError(w, r, cerrors.New(err, "failed to get session and user from http request", nil))
			return
		}

		ctxWithUser := context.WithValue(r.Context(), ctxKeyUser, user)
		ctxWithUserAndSession := context.WithValue(ctxWithUser, ctxKeySession, session)

		next.ServeHTTP(w, r.WithContext(ctxWithUserAndSession))
	})
}

// GetCurrentSession returns the session in the HTTP request context. It should only be used in HTTP request
// handlers that have the VerifySessionMiddleware on them. If a session is not found, this method will panic. To avoid
// panics, verify that a session exists either with the VerifySessionMiddleware or the HasVerifiedSession function.
func GetCurrentSession(ctx context.Context) *Session {
	session, ok := ctx.Value(ctxKeySession).(*Session)
	if !ok || session == nil {
		panic("session not found in context")
	}

	return session
}

// GetCurrentUser returns the user in the HTTP request context. It should only be used in HTTP request
// handlers that have the VerifySessionMiddleware on them. If a user is not found, this method will panic. To avoid
// panics, verify that a user exists either with the VerifySessionMiddleware or the HasVerifiedSession function.
func GetCurrentUser(ctx context.Context) *User {
	user, ok := ctx.Value(ctxKeyUser).(*User)
	if !ok || user == nil {
		panic("user not found in context")
	}

	return user
}

// HasVerifiedSession checks if the context has a valid session
func HasVerifiedSession(ctx context.Context) bool {
	session, ok := ctx.Value(ctxKeySession).(*Session)

	return ok && session != nil
}
