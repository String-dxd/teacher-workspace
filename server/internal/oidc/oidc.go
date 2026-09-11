package oidc

import (
	"context"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// RelyingParty holds the OAuth2 config and ID-token verifier needed for the
// authorization code flow with PKCE.
type RelyingParty struct {
	OAuth2   oauth2.Config
	Verifier *coreoidc.IDTokenVerifier
}

// New constructs a RelyingParty from explicit endpoint URLs. No network call
// is made; JWKS keys are fetched lazily on the first Verify() call.
func New(issuerURL, clientID, clientSecret, redirectURL, authURL, tokenURL, jwksURI string) *RelyingParty {
	keySet := coreoidc.NewRemoteKeySet(context.Background(), jwksURI)
	verifier := coreoidc.NewVerifier(issuerURL, keySet, &coreoidc.Config{ClientID: clientID})

	cfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{coreoidc.ScopeOpenID},
	}

	return &RelyingParty{
		OAuth2:   cfg,
		Verifier: verifier,
	}
}
