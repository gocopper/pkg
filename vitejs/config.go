package vitejs

import (
	"net/url"

	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/cerrors"
)

func LoadConfig(appConfig cconfig.Loader) (Config, error) {
	config := Config{
		Host: "http://localhost:3000",
	}

	err := appConfig.Load("vitejs", &config)
	if err != nil {
		return Config{}, cerrors.New(err, "failed to load vitejs config", nil)
	}

	host, err := url.Parse(config.Host)
	if err != nil {
		return Config{}, cerrors.New(err, "failed to parse host as url", map[string]interface{}{
			"host": config.Host,
		})
	}
	config.hostURL = host

	return config, nil
}

type Config struct {
	DevMode bool   `toml:"dev_mode"`
	Host    string `toml:"host"`

	hostURL *url.URL
}
