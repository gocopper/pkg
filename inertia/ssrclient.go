package inertia

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gocopper/copper/cerrors"
	"io"
	"net/http"
)

func NewSSRClient(config Config) *SSRClient {
	return &SSRClient{
		config: config,
		http:   http.DefaultClient,
	}
}

type SSRClient struct {
	config Config
	http   *http.Client
}

func (c *SSRClient) Render(ctx context.Context, page *Page) (*SSRRenderResponse, error) {
	var (
		url = c.config.SSRServer + "/render"

		renderResponse SSRRenderResponse
	)

	reqBody, err := json.Marshal(page)
	if err != nil {
		return nil, cerrors.New(err, "failed to marshal page", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, cerrors.New(err, "failed to create request", nil)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, cerrors.New(err, "failed to make request", nil)
	}
	defer resp.Body.Close()

	rawRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, cerrors.New(err, "failed to read response body", nil)
	}

	if err := json.Unmarshal(rawRespBody, &renderResponse); err != nil {
		return nil, cerrors.New(err, "failed to decode response", map[string]any{
			"body": string(rawRespBody),
		})
	}

	return &renderResponse, nil
}
