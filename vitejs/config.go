package vitejs

import (
	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/cerrors"
)

func LoadConfig(appConfig cconfig.Loader) (Config, error) {
	var config Config

	err := appConfig.Load("vitejs", &config)
	if err != nil {
		return Config{}, cerrors.New(err, "failed to load vitejs config", nil)
	}

	return config, nil
}

type Config struct {
	DevMode bool `toml:"dev_mode"`
}
