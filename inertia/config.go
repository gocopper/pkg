package inertia

import (
	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/cerrors"
)

func LoadConfig(configs cconfig.Loader) (Config, error) {
	var config Config

	err := configs.Load("inertia", &config)
	if err != nil {
		return Config{}, cerrors.New(err, "failed to load inertia config", nil)
	}

	if config.SSRServer == "" {
		config.SSRServer = "http://localhost:13714"
	}

	return config, nil
}

type Config struct {
	SSR       bool   `toml:"ssr"`
	SSRServer string `toml:"ssr_server"`
}
