package cauthtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/cconfig/cconfigtest"
	"github.com/gocopper/copper/csql"

	"github.com/gocopper/pkg/cmailer"

	"github.com/gocopper/copper/chttp"

	"github.com/gocopper/copper/chttp/chttptest"

	"github.com/gocopper/copper/clogger"
	"github.com/gocopper/pkg/cauth"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// NewHandler instantiates and returns a http.Handler with auth router and middlewares suited for testing.
func NewHandler(t *testing.T) http.Handler {
	t.Helper()

	const (
		dbDialect = "sqlite3"
		dbDSN     = ":memory:"
	)

	var (
		logger = clogger.New()
		jsonRW = chttptest.NewJSONReaderWriter(t)
		htmlRW = chttptest.NewHTMLReaderWriter(t)
	)

	db, err := sql.Open(dbDialect, dbDSN)
	assert.NoError(t, err)

	csqlConfig := csql.Config{
		Dialect: dbDialect,
		DSN:     dbDSN,
		Migrations: csql.ConfigMigrations{
			Direction: "up",
		},
	}

	err = csql.NewMigrator(csql.NewMigratorParams{
		DB:         db,
		Migrations: csql.Migrations(cauth.SQLiteMigrations),
		Config:     csqlConfig,
		Logger:     logger,
	}).Run()
	assert.NoError(t, err)

	configDir := cconfigtest.SetupDirWithConfigs(t, map[string]string{"test.toml": ""})

	configLoader, err := cconfig.New(cconfig.Path(path.Join(configDir, "test.toml")), "")
	assert.NoError(t, err)

	config, err := cauth.LoadConfig(configLoader)
	assert.NoError(t, err)

	svc, err := cauth.NewSvc(
		cauth.NewQueries(csql.NewQuerier(db, csqlConfig, logger)),
		cmailer.NewLogMailer(logger),
		config,
	)
	assert.NoError(t, err)

	verifySessionMW := cauth.NewVerifySessionMiddleware(svc, htmlRW, logger)
	dbTxMW := csql.NewTxMiddleware(db, csqlConfig, logger)

	router := cauth.NewRouter(cauth.NewRouterParams{
		Auth:      svc,
		JSON:      jsonRW,
		HTML:      htmlRW,
		SessionMW: verifySessionMW,
		Logger:    logger,
	})

	handler := chttp.NewHandler(chttp.NewHandlerParams{
		Routers:           []chttp.Router{router},
		GlobalMiddlewares: []chttp.Middleware{dbTxMW},
		Logger:            logger,
	})

	return handler
}

// CreateNewUserSession creates a new user using the given router and returns the session created by it.
func CreateNewUserSession(t *testing.T, server *httptest.Server) *cauth.SessionResult {
	t.Helper()

	var session cauth.SessionResult

	reqBody := strings.NewReader(`{
		"username": "test-user",
		"password": "test-pass"
	}`)

	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost,
		server.URL+"/api/auth/signup",
		reqBody,
	)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBodyJ, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(respBodyJ, &session)
	assert.NoError(t, err)

	return &session
}
