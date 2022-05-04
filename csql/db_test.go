package csql_test

import (
	"path"
	"testing"

	"github.com/gocopper/copper"
	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/cconfig/cconfigtest"
	"github.com/gocopper/copper/clogger"
	"github.com/gocopper/pkg/csql"
	"github.com/stretchr/testify/assert"
)

func TestNewDBConnection(t *testing.T) {
	t.Parallel()

	logger := clogger.New()
	lc := copper.NewLifecycle(logger)

	configDir := cconfigtest.SetupDirWithConfigs(t, map[string]string{"test.toml": `
[csql]
dsn = ":memory:"
`})

	config, err := cconfig.New(cconfig.Path(path.Join(configDir, "test.toml")))
	assert.NoError(t, err)

	db, err := csql.NewDBConnection(lc, config, logger)
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)

	assert.NoError(t, sqlDB.Ping())

	lc.Stop()

	assert.Error(t, sqlDB.Ping())
}
