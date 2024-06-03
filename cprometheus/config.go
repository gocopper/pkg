package cprometheus

import (
	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/cerrors"
)

type Config struct {
	HTTPEnabled bool   `toml:"http_enabled"`
	HTTPPath    string `toml:"http_path"`
}

func LoadConfig(loader cconfig.Loader) (Config, error) {
	var config = Config{
		HTTPEnabled: false,
		HTTPPath:    "/internal/metrics",
	}

	err := loader.Load("cprometheus", &config)
	if err != nil {
		return Config{}, cerrors.New(err, "failed to load cprometheus config", nil)
	}

	return config, nil
}
