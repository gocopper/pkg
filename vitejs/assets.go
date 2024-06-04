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

	devModeEntryPointURL *string
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
		type ManifestEntry struct {
			File    string   `json:"file"`
			IsEntry bool     `json:"isEntry"`
			Src     string   `json:"src"`
			CSS     []string `json:"css"`
		}

		var (
			out strings.Builder

			manifest map[string]*ManifestEntry
		)

		manifestFile, err := a.staticDir.Open("static/manifest.json")
		if err != nil {
			return "", cerrors.New(err, "failed to open manifest.json", nil)
		}

		err = json.NewDecoder(manifestFile).Decode(&manifest)
		if err != nil {
			return "", cerrors.New(err, "failed to decode manifest.json", nil)
		}

		var entryPoint *ManifestEntry
		for _, entry := range manifest {
			if entry.IsEntry {
				entryPoint = entry
				break
			}
		}
		if entryPoint == nil {
			return "", cerrors.New(nil, "no entry point found in manifest.json", nil)
		}

		if len(entryPoint.CSS) == 1 {
			out.WriteString(fmt.Sprintf("<link rel=\"stylesheet\" href=\"/static/%s\" />\n", entryPoint.CSS[0]))
		}

		out.WriteString(fmt.Sprintf("<script type=\"module\" src=\"/static/%s\"></script>", entryPoint.File))

		//nolint:gosec
		return template.HTML(out.String()), nil
	}
}

func (a *Assets) dev(req *http.Request) interface{} {
	if a.devModeEntryPointURL == nil {
		for _, ext := range []string{".js", ".ts", ".jsx", ".tsx"} {
			var url = a.config.hostURL.ResolveReference(urlMustParse("/src/main" + ext)).String()

			entryPointReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, url, nil)
			if err != nil {
				return cerrors.New(err, "failed to create request for entrypoint", nil)
			}

			resp, err := http.DefaultClient.Do(entryPointReq)
			if err != nil {
				return cerrors.New(err, "failed to execute request for entrypoint", nil)
			}

			if resp.StatusCode == http.StatusOK {
				a.devModeEntryPointURL = &url
			}
		}
	}

	return func() (template.HTML, error) {
		var (
			reactRefreshURL = a.config.hostURL.ResolveReference(urlMustParse("/@react-refresh")).String()
			viteClientURL   = a.config.hostURL.ResolveReference(urlMustParse("/@vite/client")).String()

			out strings.Builder
		)

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
			out.WriteString(fmt.Sprintf(`
<script type="module">
  import RefreshRuntime from '%s'
  RefreshRuntime.injectIntoGlobalHook(window)
  window.$RefreshReg$ = () => {}
  window.$RefreshSig$ = () => (type) => type
  window.__vite_plugin_react_preamble_installed__ = true
</script>`, reactRefreshURL))
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

		out.WriteString(fmt.Sprintf(`
<script type="module" src="%s"></script>
<script type="module" src="%s"></script>`, viteClientURL, *a.devModeEntryPointURL))

		// nolint:gosec
		return template.HTML(out.String()), nil
	}
}
