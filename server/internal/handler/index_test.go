package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/String-sg/teacher-workspace/server/internal/config"
	"github.com/String-sg/teacher-workspace/server/internal/httputil"
)

func TestHandler_index(t *testing.T) {
	t.Run("templates the dev server page for a page load in development environment", func(t *testing.T) {
		devServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Like rsbuild, only fall back to the page for a request that
			// accepts HTML.
			if !strings.Contains(r.Header.Get("Accept"), "text/html") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set(httputil.HeaderContentType, httputil.MIMETextHTMLCharsetUTF8)
			_, _ = w.Write([]byte(`<html>` + r.URL.Path + `<script type="application/json" id="runtime-config">{{.}}</script></html>`))
		}))
		t.Cleanup(devServer.Close)

		devServerURL, err := url.Parse(devServer.URL)
		if err != nil {
			t.Fatalf("url.Parse: %v", err)
		}

		h, err := New(&config.Config{
			Env:          config.EnvDevelopment,
			DevServerURL: devServerURL,
			Remote:       config.RemoteConfig{PostsManifestURL: "https://pg.test/mf-manifest.json"},
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		rec := httptest.NewRecorder()

		h.index(rec, req)

		if want, got := http.StatusOK, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
		if want, got := `<html>/dashboard<script type="application/json" id="runtime-config">{"remotes":[{"name":"pg","entry":"https://pg.test/mf-manifest.json"}]}</script></html>`, rec.Body.String(); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := "no-store", rec.Header().Get("Cache-Control"); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
	})

	t.Run("proxies everything but a page load in development environment", func(t *testing.T) {
		devServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("proxied:" + r.URL.Path))
		}))
		t.Cleanup(devServer.Close)

		devServerURL, err := url.Parse(devServer.URL)
		if err != nil {
			t.Fatalf("url.Parse: %v", err)
		}

		h, err := New(&config.Config{
			Env:          config.EnvDevelopment,
			DevServerURL: devServerURL,
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/mf-manifest.json", nil)
		req.Header.Set("Accept", "*/*")
		rec := httptest.NewRecorder()

		h.index(rec, req)

		if want, got := http.StatusOK, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
		if want, got := "proxied:/mf-manifest.json", rec.Body.String(); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
	})

	t.Run("answers 502 when the dev server is unreachable in development environment", func(t *testing.T) {
		h, err := New(&config.Config{
			Env:          config.EnvDevelopment,
			DevServerURL: &url.URL{Scheme: "http", Host: "127.0.0.1:1"},
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept", "text/html")
		rec := httptest.NewRecorder()

		h.index(rec, req)

		if want, got := http.StatusBadGateway, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
	})

	t.Run("serves the page rendered at startup for all routes in production environment", func(t *testing.T) {
		buildDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte(`<html><script type="application/json" id="runtime-config">{{.}}</script></html>`), 0o644); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}

		h, err := New(&config.Config{
			Env:      config.EnvProduction,
			BuildDir: buildDir,
			Remote: config.RemoteConfig{
				PostsManifestURL:           "https://pg.test/mf-manifest.json",
				StudentInsightsManifestURL: "https://si.test/mf-manifest.json",
			},
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		h.index(rec, req)

		if want, got := http.StatusOK, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
		if want, got := `<html><script type="application/json" id="runtime-config">{"remotes":[{"name":"pg","entry":"https://pg.test/mf-manifest.json"},{"name":"si","entry":"https://si.test/mf-manifest.json"}]}</script></html>`, rec.Body.String(); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := httputil.MIMETextHTMLCharsetUTF8, rec.Header().Get(httputil.HeaderContentType); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := "no-store", rec.Header().Get("Cache-Control"); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
	})

	t.Run("embeds an empty array rather than null when no remote is configured", func(t *testing.T) {
		buildDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte(`<script type="application/json" id="runtime-config">{{.}}</script>`), 0o644); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}

		h, err := New(&config.Config{
			Env:      config.EnvProduction,
			BuildDir: buildDir,
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.index(rec, req)

		if want, got := `<script type="application/json" id="runtime-config">{"remotes":[]}</script>`, rec.Body.String(); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
	})

	t.Run("return 404 for an unknown environment", func(t *testing.T) {
		h, err := New(&config.Config{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.index(rec, req)

		if want, got := http.StatusNotFound, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
	})
}

func TestHandler_static(t *testing.T) {
	t.Run("proxy to the dev server in development environment", func(t *testing.T) {
		devServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(httputil.HeaderContentType, httputil.MIMETextHTMLCharsetUTF8)
			_, _ = w.Write([]byte("proxied:" + r.URL.Path))
		}))
		t.Cleanup(devServer.Close)

		devServerURL, err := url.Parse(devServer.URL)
		if err != nil {
			t.Fatalf("url.Parse: %v", err)
		}

		h, err := New(&config.Config{
			Env:          config.EnvDevelopment,
			DevServerURL: devServerURL,
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/static/js/index.js", nil)
		rec := httptest.NewRecorder()

		h.static(rec, req)

		if want, got := http.StatusOK, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
		if want, got := "proxied:/static/js/index.js", rec.Body.String(); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
	})

	t.Run("serve hashed asset in production environment", func(t *testing.T) {
		buildDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>Hello world!</html>"), 0o644); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(buildDir, "static", "js"), 0o755); err != nil {
			t.Fatalf("os.MkdirAll: %v", err)
		}
		if err := os.WriteFile(filepath.Join(buildDir, "static", "js", "index.abc123.js"), []byte("console.log('Hello world!');"), 0o644); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}

		h, err := New(&config.Config{
			Env:      config.EnvProduction,
			BuildDir: buildDir,
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/static/js/index.abc123.js", nil)
		rec := httptest.NewRecorder()

		h.static(rec, req)

		if want, got := http.StatusOK, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
		if want, got := "console.log('Hello world!');", rec.Body.String(); want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
	})

	t.Run("return 404 for an unknown environment", func(t *testing.T) {
		h, err := New(&config.Config{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/static/js/missing.js", nil)
		rec := httptest.NewRecorder()

		h.static(rec, req)

		if want, got := http.StatusNotFound, rec.Code; want != got {
			t.Errorf("want: %d; got: %d", want, got)
		}
	})
}

func TestNewRuntimeConfig(t *testing.T) {
	t.Run("maps posts to pg and student insights to si in order", func(t *testing.T) {
		got := newRuntimeConfig(config.RemoteConfig{
			PostsManifestURL:           "https://pg.test/mf-manifest.json",
			StudentInsightsManifestURL: "https://si.test/mf-manifest.json",
		})

		want := []runtimeRemote{
			{Name: "pg", Entry: "https://pg.test/mf-manifest.json"},
			{Name: "si", Entry: "https://si.test/mf-manifest.json"},
		}
		if !slices.Equal(want, got.Remotes) {
			t.Errorf("want: %v; got: %v", want, got.Remotes)
		}
	})

	t.Run("skips a remote whose url is empty", func(t *testing.T) {
		got := newRuntimeConfig(config.RemoteConfig{
			StudentInsightsManifestURL: "https://si.test/mf-manifest.json",
		})

		want := []runtimeRemote{{Name: "si", Entry: "https://si.test/mf-manifest.json"}}
		if !slices.Equal(want, got.Remotes) {
			t.Errorf("want: %v; got: %v", want, got.Remotes)
		}
	})

	t.Run("returns an empty slice rather than nil when nothing is configured", func(t *testing.T) {
		got := newRuntimeConfig(config.RemoteConfig{})

		if got.Remotes == nil {
			t.Fatal("want: non-nil; got: nil")
		}
		if want := 0; want != len(got.Remotes) {
			t.Errorf("want: %d; got: %d", want, len(got.Remotes))
		}
	})
}
