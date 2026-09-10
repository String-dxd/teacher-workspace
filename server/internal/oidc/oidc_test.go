package oidc_test

import (
	"testing"

	"golang.org/x/oauth2"

	"github.com/String-sg/teacher-workspace/server/internal/oidc"
)

func TestNew(t *testing.T) {
	t.Run("returns a configured RelyingParty", func(t *testing.T) {
		rp := oidc.New(
			"https://issuer.example.com",
			"test-client-id",
			"test-client-secret",
			"http://localhost/callback",
			"https://issuer.example.com/authorize",
			"https://issuer.example.com/token",
			"https://issuer.example.com/jwks",
		)

		if want, got := "test-client-id", rp.OAuth2.ClientID; want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := "http://localhost/callback", rp.OAuth2.RedirectURL; want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := oauth2.AuthStyleInParams, rp.OAuth2.Endpoint.AuthStyle; want != got {
			t.Errorf("want: %v; got: %v", want, got)
		}
		if want, got := "https://issuer.example.com/authorize", rp.OAuth2.Endpoint.AuthURL; want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if want, got := "https://issuer.example.com/token", rp.OAuth2.Endpoint.TokenURL; want != got {
			t.Errorf("want: %q; got: %q", want, got)
		}
		if got := rp.Verifier; got == nil {
			t.Error("want: non-nil; got: nil")
		}
	})
}
