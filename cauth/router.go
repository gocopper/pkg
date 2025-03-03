package cauth

import (
	"errors"
	"net/http"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/copper/chttp"
	"github.com/gocopper/copper/clogger"
)

// NewRouterParams holds the dependencies to create a new Router.
type NewRouterParams struct {
	Auth      *Svc
	SessionMW *VerifySessionMiddleware
	JSON      *chttp.JSONReaderWriter
	HTML      *chttp.HTMLReaderWriter
	Logger    clogger.Logger
}

// NewRouter instantiates and returns a new Router.
func NewRouter(p NewRouterParams) *Router {
	return &Router{
		svc:       p.Auth,
		json:      p.JSON,
		html:      p.HTML,
		sessionMW: p.SessionMW,
		logger:    p.Logger,
	}
}

// Router handles incoming HTTP requests related the cauth package.
type Router struct {
	svc       *Svc
	sessionMW chttp.Middleware
	json      *chttp.JSONReaderWriter
	html      *chttp.HTMLReaderWriter
	logger    clogger.Logger
}

// Routes returns the routes managed by this router.
func (ro *Router) Routes() []chttp.Route {
	return []chttp.Route{
		{
			Path:    "/api/auth/signup",
			Methods: []string{http.MethodPost},
			Handler: ro.HandleSignup,
		},
		{
			Path:    "/api/auth/verify-email",
			Methods: []string{http.MethodPost},
			Handler: ro.HandleVerifyEmail,
		},
		{
			Path:    "/api/auth/login",
			Methods: []string{http.MethodPost},
			Handler: ro.HandleLogin,
		},
		{
			Middlewares: []chttp.Middleware{ro.sessionMW},
			Path:        "/api/auth/logout",
			Methods:     []string{http.MethodPost},
			Handler:     ro.HandleLogout,
		},
	}
}

// HandleSignup handles a user signup request.
func (ro *Router) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var params SignupParams

	if !ro.json.ReadJSON(w, r, &params) {
		return
	}

	sessionResult, err := ro.svc.Signup(r.Context(), params)
	if err != nil && errors.Is(err, ErrUserAlreadyExists) {
		ro.json.WriteJSON(w, chttp.WriteJSONParams{
			StatusCode: http.StatusBadRequest,
			Data:       map[string]string{"error": "user already exists"},
		})
	} else if err != nil {
		ro.logger.Error("Failed to signup", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	ro.json.WriteJSON(w, chttp.WriteJSONParams{
		Data: sessionResult,
	})
}

// HandleVerifyEmail handles a user email verification request.
func (ro *Router) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var params VerifyEmailParams

	if !ro.json.ReadJSON(w, r, &params) {
		return
	}

	_, err := ro.svc.VerifyEmail(r.Context(), params)
	if err != nil && errors.Is(err, ErrInvalidCredentials) {
		ro.html.Unauthorized(w, r)
		return
	} else if err != nil {
		ro.html.WriteHTMLError(w, r, cerrors.New(err, "failed to verify email", map[string]interface{}{
			"email": params.Email,
		}))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleLogin handles a user login request.
func (ro *Router) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var params LoginParams

	if !ro.json.ReadJSON(w, r, &params) {
		return
	}

	sessionResult, err := ro.svc.Login(r.Context(), params)
	if err != nil && errors.Is(err, ErrInvalidCredentials) {
		ro.html.Unauthorized(w, r)
		return
	} else if err != nil {
		ro.html.WriteHTMLError(w, r, cerrors.New(err, "failed to login", map[string]interface{}{
			"email": params.Email,
		}))
		return
	}

	ro.json.WriteJSON(w, chttp.WriteJSONParams{
		Data: sessionResult,
	})
}

// HandleLogout handles a user logout request.
func (ro *Router) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		session = GetCurrentSession(ctx)
	)

	err := ro.svc.Logout(ctx, session.UUID)
	if err != nil {
		ro.html.WriteHTMLError(w, r, cerrors.New(err, "failed to logout", map[string]interface{}{
			"session": session.UUID,
		}))
		return
	}
}
