package authz

import (
	"context"
	"fmt"

	"github.com/csams/common-inventory/pkg/authz/allow"
	"github.com/csams/common-inventory/pkg/authz/api"
	"github.com/csams/common-inventory/pkg/authz/kessel"
)

func New(ctx context.Context, config CompletedConfig) (api.Authorizer, error) {

	switch config.Authz {
	case "allow-all":
		return allow.New(), nil
	case "kessel":
		return kessel.New(ctx, config.Kessel)
	default:
		return nil, fmt.Errorf("Unrecognized authz.impl: %s", config.Authz)
	}
}
