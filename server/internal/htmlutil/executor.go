// Package htmlutil renders the frontend's HTML template, sourcing it from the
// development server or from the build output depending on the environment.
package htmlutil

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/String-sg/teacher-workspace/server/internal/httputil"
)

// devServerTimeout bounds a template fetch, so a stalled development server
// fails the request rather than holding it until the server's write timeout.
const devServerTimeout = 10 * time.Second

// TemplateExecutor renders an HTML template with the given data.
type TemplateExecutor interface {
	Execute(ctx context.Context, w io.Writer, data any) error
}

// DevelopmentTemplateExecutor fetches an HTML template from a URL on every call
// to Execute, so that changes to the template are picked up without restarting
// the server.
type DevelopmentTemplateExecutor struct {
	client *http.Client
	url    string
}

// NewDevelopmentTemplateExecutor returns a [DevelopmentTemplateExecutor] that
// fetches the template from the given URL.
func NewDevelopmentTemplateExecutor(url string) *DevelopmentTemplateExecutor {
	return &DevelopmentTemplateExecutor{
		client: &http.Client{Timeout: devServerTimeout},
		url:    url,
	}
}

// Execute fetches the template, parses it, and renders it with the given data.
// It reports an error rather than writing a partial page when the template
// cannot be fetched or parsed.
func (e *DevelopmentTemplateExecutor) Execute(ctx context.Context, w io.Writer, data any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	// The rsbuild development server serves the template only to a request that
	// accepts HTML, and Go sends no Accept header of its own.
	req.Header.Set("Accept", httputil.MIMETextHTML)

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server responded %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	tmpl, err := template.New("index.html").Parse(string(respBody))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	return tmpl.Execute(w, data)
}

// ProductionTemplateExecutor parses an HTML template once and reuses it for
// every call to Execute. It is safe for concurrent use.
type ProductionTemplateExecutor struct {
	tmpl *template.Template
}

// NewProductionTemplateExecutor parses the template file at the given path,
// reporting an error when it is missing or malformed.
func NewProductionTemplateExecutor(path string) (*ProductionTemplateExecutor, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &ProductionTemplateExecutor{tmpl: tmpl}, nil
}

// Execute renders the parsed template with the given data.
func (e *ProductionTemplateExecutor) Execute(_ context.Context, w io.Writer, data any) error {
	return e.tmpl.Execute(w, data)
}
