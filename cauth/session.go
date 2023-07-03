package cauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gocopper/copper/chttp"
	"github.com/gocopper/copper/clogger"
)

type ctxKey string

const (
	ctxKeySession = ctxKey("cauth/session")
	ctxKeyUser    = ctxKey("cauth/user")
)

// NewVerifySessionMiddleware instantiates and creates a new VerifySessionMiddleware
func NewVerifySessionMiddleware(auth *Svc, rw *chttp.ReaderWriter, logger clogger.Logger) *VerifySessionMiddleware {
	return &VerifySessionMiddleware{
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
	rw     *chttp.ReaderWriter
	logger clogger.Logger
}

// Handle implements the middleware for VerifySessionMiddleware.
func (mw *VerifySessionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			sessionUUID string
			plainToken  string
		)

		sessionUUIDCookie, err := r.Cookie("SessionUUID")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			mw.logger.Error("Failed to read session uuid cookie", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		sessionTokenCookie, err := r.Cookie("SessionToken")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			mw.logger.Error("Failed to read session token cookie", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
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
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ok, session, err := mw.auth.ValidateSession(r.Context(), sessionUUID, plainToken)
		if err != nil {
			mw.logger.WithTags(map[string]interface{}{
				"sessionUUID": sessionUUID,
			}).Error("Failed to verify session token", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, err := mw.auth.GetUserByUUID(r.Context(), session.UserUUID)
		if err != nil {
			mw.logger.WithTags(map[string]interface{}{
				"userUUID": session.UserUUID,
			}).Error("Failed to get user by uuid", err)
			w.WriteHeader(http.StatusInternalServerError)

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
