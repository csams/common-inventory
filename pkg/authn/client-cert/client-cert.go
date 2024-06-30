package clientcert

import (
	"log/slog"
	"net/http"

	"github.com/csams/common-inventory/pkg/authn/api"
)

type ClientCertAuthenticator struct{}

func New(log *slog.Logger) *ClientCertAuthenticator {
	return &ClientCertAuthenticator{}
}

func (a *ClientCertAuthenticator) Authenticate(r *http.Request) (*api.Identity, api.Decision) {
	if r.TLS == nil {
		return nil, api.Ignore
	}

	if len(r.TLS.PeerCertificates) == 0 {
		return nil, api.Ignore
	}

	cert := r.TLS.PeerCertificates[0]

	// TODO: What do we do about tenant id here?
	// TODO: Should we say all reporters will authenticate with client certificates?
	return &api.Identity{
		Principcal: cert.Subject.CommonName,
		Groups:     cert.Subject.Organization,
		IsReporter: true,
	}, api.Allow
}