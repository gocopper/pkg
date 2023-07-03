package cauth

import (
	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/cerrors"
)

// Config configures the cauth module
type Config struct {
	VerificationCodeLen       uint   `toml:"verification_code_len"`
	VerificationEmailSubject  string `toml:"verification_email_subject"`
	VerificationEmailFrom     string `toml:"verification_email_from"`
	VerificationEmailBodyHTML string `toml:"verification_email_body_html"`
}

// LoadConfig loads the config for cauth module
func LoadConfig(loader cconfig.Loader) (Config, error) {
	var config = Config{
		VerificationCodeLen:       4,
		VerificationEmailSubject:  "Your Verification Code",
		VerificationEmailFrom:     "webmaster@example.com",
		VerificationEmailBodyHTML: `Your verification code is <b>{{.VerificationCode}}</b>`,
	}

	err := loader.Load("cauth", &config)
	if err != nil {
		return Config{}, cerrors.New(err, "failed to load cauth config", nil)
	}

	return config, nil
}
