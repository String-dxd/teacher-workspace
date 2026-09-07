package oidc_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"

	"github.com/String-sg/teacher-workspace/server/internal/oidc"
)

func newTestOIDCServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		base := srv.URL
		doc := map[string]any{
			"issuer":                                base,
			"authorization_endpoint":                base + "/authorize",
			"token_endpoint":                        base + "/token",
			"jwks_uri":                              base + "/jwks",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc) //nolint:errcheck
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"keys":[]}`)) //nolint:errcheck
	})

	return srv
}

func TestNew(t *testing.T) {
	t.Run("discovers the provider and returns a configured RelyingParty", func(t *testing.T) {
		srv := newTestOIDCServer(t)

		rp, err := oidc.New(t.Context(), srv.URL, "test-client-id", "test-client-secret", "http://localhost/callback")
		if err != nil {
			t.Fatalf("oidc.New: %v", err)
		}

		if want, got := "test-client-id", rp.OAuth2.ClientID; want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := "http://localhost/callback", rp.OAuth2.RedirectURL; want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := oauth2.AuthStyleInParams, rp.OAuth2.Endpoint.AuthStyle; want != got {
			t.Errorf("want: %v; got: %v", want, got)
		}
		if got := rp.Verifier; got == nil {
			t.Error("want: non-nil; got: nil")
		}
	})

	t.Run("returns an error when the issuer is unreachable", func(t *testing.T) {
		srv := newTestOIDCServer(t)
		srv.Close()

		_, err := oidc.New(t.Context(), srv.URL, "test-client-id", "test-client-secret", "http://localhost/callback")
		if err == nil {
			t.Fatal("New() returned nil error, want non-nil")
		}
	})
}
