package api

import (
	"net/http"
)

type Decision string

const (
	Allow  Decision = "ALLOW"
	Deny            = "DENY"
	Ignore          = "IGNORE"
)

type Authenticator interface {
	Authenticate(r *http.Request) (*Identity, Decision)
}
