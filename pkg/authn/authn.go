package authn

import (
	"github.com/csams/common-inventory/pkg/authn/api"
	"github.com/csams/common-inventory/pkg/authn/clientcert"
	"github.com/csams/common-inventory/pkg/authn/delegator"
	"github.com/csams/common-inventory/pkg/authn/guest"
	"github.com/csams/common-inventory/pkg/authn/oidc"
	"github.com/csams/common-inventory/pkg/authn/psk"
)

func New(config CompletedConfig) (api.Authenticator, error) {
	d := delegator.New()

    // client certs authn
	d.Add(clientcert.New())

    // pre shared key authn
	if config.PreSharedKeys != nil {
        a := psk.New(*config.PreSharedKeys)
		d.Add(a)
	}

    // oidc tokens
	if config.Oidc != nil {
		if a, err := oidc.New(*config.Oidc); err == nil {
			d.Add(a)
		} else {
			return nil, err
		}
	}

    // unauthenticated
    // TODO: make it configurable whether we allow unauthenticated access
    d.Add(guest.New())

	return d, nil
}
