package api

import (
	"context"

	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	"github.com/csams/common-inventory/pkg/models"
)

type Manager interface {
	Lookup(identity *authnapi.Identity, resource *models.Resource) (Producer, error)
	Errs() <-chan error
	Shutdown(ctx context.Context) error
}
