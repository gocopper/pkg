package vitejs

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/copper/chttp"
)

func NewAssets(staticDir chttp.StaticDir, config Config) *Assets {
	return &Assets{
		staticDir: staticDir,
		config:    config,
	}
}

type Assets struct {
	staticDir chttp.StaticDir
	config    Config
}

func (a *Assets) HTMLRenderFunc() chttp.HTMLRenderFunc {
	return chttp.HTMLRenderFunc{
		Name: "viteAssets",
		Func: a.Assets,
	}
}

func (a *Assets) Assets(req *http.Request) interface{} {
	if a.config.DevMode {
		return a.dev(req)
	}

	return a.prod()
}

func (a *Assets) prod() interface{} {
	return func() (template.HTML, error) {
		var (
			out      strings.Builder
			manifest struct {
				MainJS struct {
					File string   `json:"file"`
					CSS  []string `json:"css"`
				} `json:"src/main.js"`
			}
		)

		manifestFile, err := a.staticDir.Open("static/manifest.json")
		if err != nil {
			return "", cerrors.New(err, "failed to open manifest.json", nil)
		}

		err = json.NewDecoder(manifestFile).Decode(&manifest)
		if err != nil {
			return "", cerrors.New(err, "failed to decode manifest.json", nil)
		}

		if len(manifest.MainJS.CSS) == 1 {
			out.WriteString(fmt.Sprintf("<link rel=\"stylesheet\" href=\"/static/%s\" />\n", manifest.MainJS.CSS[0]))
		}

		out.WriteString(fmt.Sprintf("<script type=\"module\" src=\"/static/%s\"></script>", manifest.MainJS.File))

		//nolint:gosec
		return template.HTML(out.String()), nil
	}
}

func (a *Assets) dev(req *http.Request) interface{} {
	return func() (template.HTML, error) {
		const reactRefreshURL = "http://localhost:3000/@react-refresh"
		var out strings.Builder

		reactReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, reactRefreshURL, nil)
		if err != nil {
			return "", cerrors.New(err, "failed to create request for @react-refresh", nil)
		}

		resp, err := http.DefaultClient.Do(reactReq)
		if err != nil {
			return "", cerrors.New(err, "failed to execute request for @react-refresh", nil)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusOK {
			out.WriteString(`
<script type="module">
  import RefreshRuntime from 'http://localhost:3000/@react-refresh'
  RefreshRuntime.injectIntoGlobalHook(window)
  window.$RefreshReg$ = () => {}
  window.$RefreshSig$ = () => (type) => type
  window.__vite_plugin_react_preamble_installed__ = true
</script>`)
		}

		// note: in dev mode only, the css is not part of the initial page load. since it is loaded async, there is
		// a brief time period where the page has no styles. to avoid this, the following snippet hides the
		// body until the css has been loaded.
		out.WriteString(`
<style type="text/css" id="copper-hide-body">
	body { visibility: hidden; }
</style>
<script type="text/javascript">
	(function() {
		let interval;
	
		function showBodyIfStylesPresent() {
			const styleEls = document.getElementsByTagName('style');
			const copperHideBodyStyleEl = document.getElementById('copper-hide-body');
			
			if (!copperHideBodyStyleEl || styleEls.length === 1) {
				return;
			}
			
			copperHideBodyStyleEl.remove();
			clearInterval(interval);
		}
	
		interval = setInterval(showBodyIfStylesPresent, 100);
	})();
</script>
`)

		out.WriteString(`
<script type="module" src="http://localhost:3000/@vite/client"></script>
<script type="module" src="http://localhost:3000/src/main.js"></script>`)

		// nolint:gosec
		return template.HTML(out.String()), nil
	}
}
