package oauth2

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/csams/common-inventory/pkg/authn/api"
	"github.com/csams/common-inventory/pkg/authn/util"
)

type OAuth2Authenticator struct {
	CompletedConfig

	ClientContext context.Context
	Verifier      *oidc.IDTokenVerifier
}

func New(c CompletedConfig) (*OAuth2Authenticator, error) {
	ctx := context.Background()
	ctx = oidc.ClientContext(ctx, c.Client)

	oidcConfig := &oidc.Config{ClientID: c.ClientId}

	provider, err := oidc.NewProvider(ctx, c.AuthorizationServerURL)
	if err != nil {
		return nil, err
	}

	return &OAuth2Authenticator{
		CompletedConfig: c,
		ClientContext:   ctx,
		Verifier:        provider.Verifier(oidcConfig),
	}, nil

}

func (o *OAuth2Authenticator) Authenticate(r *http.Request) (*api.Identity, api.Decision) {
	// get the token from the request
	rawToken := util.GetBearerToken(r)

	// ensure we got one
	if rawToken == "" {
		return nil, api.Ignore
	}

	// verify and parse it
	tok, err := o.Verify(rawToken)
	if err != nil {
		return nil, api.Deny
	}

	// extract the claims we care about
	u := &Claims{}
	tok.Claims(u)
	if u.Id == "" {
		return nil, api.Deny
	}

	if u.Audience != o.CompletedConfig.ClientId {
		return nil, api.Deny
	}

    // TODO: What are the tenant and group claims?
    return &api.Identity{ Principcal: u.Id }, api.Allow
}

// Claims holds the values we want to extract from the JWT.
// TODO: make JWT claim fields configurable
type Claims struct {
	Id       string `json:"preferred_username"`
	Audience string `json:"aud"`
}

func (l *OAuth2Authenticator) Verify(token string) (*oidc.IDToken, error) {
	return l.Verifier.Verify(l.ClientContext, token)
}
