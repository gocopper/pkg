package vitejs

import (
	"net/url"

	"github.com/gocopper/copper/cerrors"
)

func urlMustParse(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(cerrors.New(err, "failed to parse raw url", map[string]interface{}{
			"rawURL": rawURL,
		}))
	}

	return u
}
