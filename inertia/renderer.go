package inertia

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/copper/clogger"
	"github.com/google/uuid"

	"github.com/gocopper/copper/chttp"
)

type Renderer struct {
	json   *chttp.JSONReaderWriter
	html   *chttp.HTMLReaderWriter
	ssr    *SSRClient
	config Config
	logger clogger.Logger

	component      *string
	layoutTemplate *string
	basePath       *string

	sharedPropsByRequestID *sync.Map
	flashPropsBySessionID  *sync.Map
}

const inertiaSessionCookieName = "inertia_session"

type NewRendererParams struct {
	HTMLReaderWriter *chttp.HTMLReaderWriter
	JSONReaderWriter *chttp.JSONReaderWriter
	SSRClient        *SSRClient
	Config           Config
	Logger           clogger.Logger
}

func NewRenderer(p NewRendererParams) *Renderer {
	return &Renderer{
		html:                   p.HTMLReaderWriter,
		json:                   p.JSONReaderWriter,
		ssr:                    p.SSRClient,
		config:                 p.Config,
		logger:                 p.Logger,
		sharedPropsByRequestID: &sync.Map{},
		flashPropsBySessionID:  &sync.Map{},
	}
}

func (r *Renderer) copy() *Renderer {
	return &Renderer{
		html:                   r.html,
		json:                   r.json,
		ssr:                    r.ssr,
		config:                 r.config,
		logger:                 r.logger,
		component:              r.component,
		layoutTemplate:         r.layoutTemplate,
		basePath:               r.basePath,
		sharedPropsByRequestID: r.sharedPropsByRequestID,
		flashPropsBySessionID:  r.flashPropsBySessionID,
	}
}

func (r *Renderer) WithComponent(component string) *Renderer {
	c := r.copy()
	c.component = &component
	return c
}

func (r *Renderer) WithLayoutTemplate(layoutTemplate string) *Renderer {
	c := r.copy()
	c.layoutTemplate = &layoutTemplate
	return c
}

func (r *Renderer) WithBasePath(basePath string) *Renderer {
	c := r.copy()
	c.basePath = &basePath
	return c
}

func (r *Renderer) ShareProps(ctx context.Context, props map[string]any) {
	var requestID = chttp.GetRequestID(ctx)

	if requestID == "" {
		r.logger.Warn("[Inertia] Tried to share props without request id", nil)
		return
	}

	// consider merging props with existing props (if any)
	r.sharedPropsByRequestID.Store(requestID, props)
}

func (r *Renderer) FlashProps(req *http.Request, props map[string]any) {
	cookie, err := req.Cookie(inertiaSessionCookieName)
	if err != nil || cookie.Value == "" {
		r.logger.Warn("[Inertia] Tried to flash props without a session", nil)
		return
	}

	// consider merging props with existing props (if any)
	r.flashPropsBySessionID.Store(cookie.Value, props)
}

func (r *Renderer) ReadForm(w http.ResponseWriter, req *http.Request, form any) bool {
	err := json.NewDecoder(req.Body).Decode(form)
	if err != nil {
		r.html.WriteHTMLError(w, req, cerrors.New(err, "failed to read json form", nil))
		return false
	}

	ok, err := govalidator.ValidateStruct(form)
	if !ok || err != nil {
		r.logger.Warn("[Inertia] Form validation failed", err)
		r.FlashProps(req, map[string]any{
			"validationError": err.Error(),
		})
		return false
	}

	return true
}

type RenderParams struct {
	Component  string
	Props      map[string]any
	MergeProps []string

	LayoutTemplate *string
	BasePath       *string
}

func (r *Renderer) Render(w http.ResponseWriter, req *http.Request, p RenderParams) {
	var (
		reqCtx = req.Context()

		isInertiaRequest = req.Header.Get("x-inertia") == "true"
		partialData      = req.Header.Get("x-inertia-partial-data")

		layoutTemplate = "main.html"

		page = Page{
			Component:  p.Component,
			Props:      p.Props,
			MergeProps: p.MergeProps,
			URL:        req.URL.String(),
			Version:    "2",
		}
	)

	if page.Props == nil {
		page.Props = make(map[string]any)
	}

	// Get or create Inertia session cookie
	var sessionID string
	cookie, err := req.Cookie(inertiaSessionCookieName)
	if err == nil && cookie.Value != "" {
		sessionID = cookie.Value
	} else {
		sessionID = uuid.New().String()
		http.SetCookie(w, &http.Cookie{
			Name:     inertiaSessionCookieName,
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   req.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})
	}

	if flashProps, ok := r.flashPropsBySessionID.Load(sessionID); ok {
		// consider merging with existing flash props
		page.Props["flash"] = flashProps

		r.flashPropsBySessionID.Delete(sessionID)
	}

	requestID := chttp.GetRequestID(reqCtx)
	if requestID != "" {
		if sharedProps, ok := r.sharedPropsByRequestID.Load(requestID); ok {
			for k, v := range sharedProps.(map[string]any) {
				page.Props[k] = v
			}

			r.sharedPropsByRequestID.Delete(requestID)
		}
	}

	if page.Component == "" && r.component != nil {
		page.Component = *r.component
	}

	if p.LayoutTemplate != nil {
		layoutTemplate = *p.LayoutTemplate
	} else if r.layoutTemplate != nil {
		layoutTemplate = *r.layoutTemplate
	}

	if page.Props == nil {
		page.Props = map[string]any{}
	}

	if p.BasePath != nil {
		page.URL = strings.TrimPrefix(page.URL, *p.BasePath)
	} else if r.basePath != nil {
		page.URL = strings.TrimPrefix(page.URL, *r.basePath)
	}

	if partialData != "" {
		propsToIncludeSet := map[string]bool{}
		for _, prop := range strings.Split(partialData, ",") {
			propsToIncludeSet[prop] = true
		}

		for key := range page.Props {
			if !propsToIncludeSet[key] {
				delete(page.Props, key)
			}
		}
	}

	if isInertiaRequest {
		w.Header().Set("x-inertia", "true")
		r.json.WriteJSON(w, chttp.WriteJSONParams{
			Data: page,
		})
		return
	}

	if !r.config.SSR {
		r.html.WriteHTML(w, req, chttp.WriteHTMLParams{
			LayoutTemplate: layoutTemplate,
			PageTemplate:   "react-inertia.gohtml",
			Data: map[string]any{
				"Page": page,
			},
		})
		return
	}

	ssrResponse, err := r.ssr.Render(req.Context(), &page)
	if err != nil {
		r.logger.WithTags(map[string]any{
			"url": req.URL.String(),
		}).Warn("Failed to render page with SSR", err)

		r.html.WriteHTML(w, req, chttp.WriteHTMLParams{
			LayoutTemplate: layoutTemplate,
			PageTemplate:   "react-inertia.gohtml",
			Data: map[string]any{
				"Page": page,
			},
		})
		return
	}

	r.html.WriteHTML(w, req, chttp.WriteHTMLParams{
		LayoutTemplate: layoutTemplate,
		PageTemplate:   "react-inertia.gohtml",
		Data: map[string]any{
			"Body": ssrResponse.Body,
		},
	})
}
