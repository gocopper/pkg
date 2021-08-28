package cauth

import (
	"errors"
	"net/http"

	"github.com/gocopper/copper/chttp"
	"github.com/gocopper/copper/clogger"
)

// NewRouterParams holds the dependencies to create a new Router.
type NewRouterParams struct {
	Auth      *Svc
	SessionMW *VerifySessionMiddleware
	RW        *chttp.ReaderWriter
	Logger    clogger.Logger
}

// NewRouter instantiates and returns a new Router.
func NewRouter(p NewRouterParams) *Router {
	return &Router{
		svc:       p.Auth,
		rw:        p.RW,
		sessionMW: p.SessionMW,
		logger:    p.Logger,
	}
}

// Router handles incoming HTTP requests related the cauth package.
type Router struct {
	svc       *Svc
	sessionMW chttp.Middleware
	rw        *chttp.ReaderWriter
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

	if !ro.rw.ReadJSON(w, r, &params) {
		return
	}

	sessionResult, err := ro.svc.Signup(r.Context(), params)
	if err != nil {
		ro.logger.Error("Failed to signup", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	ro.rw.WriteJSON(w, chttp.WriteJSONParams{
		Data: sessionResult,
	})
}

// HandleLogin handles a user login request.
func (ro *Router) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var params LoginParams

	if !ro.rw.ReadJSON(w, r, &params) {
		return
	}

	sessionResult, err := ro.svc.Login(r.Context(), params)
	if err != nil && errors.Is(err, ErrInvalidCredentials) {
		w.WriteHeader(http.StatusUnauthorized)

		return
	} else if err != nil {
		ro.logger.Error("Failed to login", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	ro.rw.WriteJSON(w, chttp.WriteJSONParams{
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
		ro.logger.Error("Failed to logout", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
