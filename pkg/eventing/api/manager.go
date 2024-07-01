package api

import (
	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	"github.com/csams/common-inventory/pkg/models"
)

type Manager interface {
    Lookup(identity *authnapi.Identity, resourceType string, resourceId models.IDType) (Producer, error)
}
