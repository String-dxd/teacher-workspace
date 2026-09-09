package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/String-sg/teacher-workspace/server/internal/config"
	"github.com/String-sg/teacher-workspace/server/internal/httputil"
	"github.com/String-sg/teacher-workspace/server/internal/middleware"
)

// runtimeRemote is a Module Federation remote the frontend registers at startup.
type runtimeRemote struct {
	Name  string `json:"name"`
	Entry string `json:"entry"`
}

// runtimeConfig is the document embedded in index.html for the frontend to read at startup.
type runtimeConfig struct {
	Remotes []runtimeRemote `json:"remotes"`
}

// newRuntimeConfig maps the configured manifest URLs to the remote names the
// frontend loads modules from, leaving out the remotes that are not set.
func newRuntimeConfig(cfg config.RemoteConfig) runtimeConfig {
	// An empty array rather than null, so the client never has to guard
	// against a missing list.
	remotes := []runtimeRemote{}

	// The pg remote exposes both the Posts and the Groups module.
	if cfg.PostsManifestURL != "" {
		remotes = append(remotes, runtimeRemote{Name: "pg", Entry: cfg.PostsManifestURL})
	}
	if cfg.StudentInsightsManifestURL != "" {
		remotes = append(remotes, runtimeRemote{Name: "si", Entry: cfg.StudentInsightsManifestURL})
	}

	return runtimeConfig{Remotes: remotes}
}

// index serves the frontend's application shell with the runtime config
// embedded. In development it templates the page from the rsbuild dev server
// and proxies every other request; in production it serves the page rendered
// at startup for all routes so client-side routing works.
func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	logger := middleware.LoggerFromContext(r.Context())

	var page []byte
	switch h.cfg.Env {
	case config.EnvDevelopment:
		// Scripts, assets and the hot reload websocket must reach the dev
		// server untouched; only a page load wants the templated shell.
		if !strings.Contains(r.Header.Get("Accept"), httputil.MIMETextHTML) {
			h.devProxy.ServeHTTP(w, r)
			return
		}

		source, err := h.fetchDevIndex(r)
		if err != nil {
			logger.Error("failed to fetch index.html from the dev server", "err", err)
			httputil.RenderPlain(w, logger, http.StatusBadGateway)
			return
		}
		page, err = renderIndex(source, h.runtime)
		if err != nil {
			logger.Error("failed to render index.html", "err", err)
			httputil.RenderPlain(w, logger, http.StatusInternalServerError)
			return
		}
	case config.EnvProduction:
		page = h.indexPage
	default:
		httputil.RenderPlain(w, logger, http.StatusNotFound)
		return
	}

	// A cached page would keep the remote list of a previous run.
	w.Header().Set("Cache-Control", "no-store")
	httputil.RenderHTML(w, logger, http.StatusOK, page)
}

// static serves the frontend's hashed static assets. In development it proxies
// to the rsbuild dev server; in production it serves files from the build
// directory.
func (h *Handler) static(w http.ResponseWriter, r *http.Request) {
	switch h.cfg.Env {
	case config.EnvDevelopment:
		h.devProxy.ServeHTTP(w, r)
	case config.EnvProduction:
		h.assets.ServeHTTP(w, r)
	default:
		logger := middleware.LoggerFromContext(r.Context())
		httputil.RenderPlain(w, logger, http.StatusNotFound)
	}
}

// fetchDevIndex fetches the page the dev server would serve for the request.
func (h *Handler) fetchDevIndex(r *http.Request) ([]byte, error) {
	target := h.cfg.DevServerURL.ResolveReference(&url.URL{Path: r.URL.Path, RawQuery: r.URL.RawQuery})
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	// The dev server falls back to index.html only for a request that accepts HTML.
	req.Header.Set("Accept", r.Header.Get("Accept"))

	resp, err := h.devClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dev server responded %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// renderIndex executes the index.html template with the runtime config.
func renderIndex(source []byte, runtime runtimeConfig) ([]byte, error) {
	tmpl, err := template.New("index.html").Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("parse index.html: %w", err)
	}

	var page bytes.Buffer
	if err := tmpl.Execute(&page, runtime); err != nil {
		return nil, fmt.Errorf("render index.html: %w", err)
	}

	return page.Bytes(), nil
}
