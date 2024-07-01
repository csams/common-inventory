package middleware

import (
	"context"
	"net/http"

	authnapi "github.com/csams/common-inventory/pkg/authn/api"
)

func Authentication(authenticator authnapi.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, decision := authenticator.Authenticate(r)
			if decision != authnapi.Allow {
				http.Error(w, "Not Authenticated", http.StatusUnauthorized)
                return
			}

			ctx := context.WithValue(r.Context(), IdentityRequestKey, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

var (
	IdentityRequestKey = &contextKey{"authnapi.Identity"}
	GetIdentity        = GetFromContext[authnapi.Identity](IdentityRequestKey)
)
