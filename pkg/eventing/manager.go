package eventing

import (
	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	"github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/eventing/stdout"
	"github.com/csams/common-inventory/pkg/models"
)

type EventingManager struct {
	Producer api.Producer
}

func New() *EventingManager {
	return &EventingManager{
		Producer: stdout.New(),
	}
}

func (m *EventingManager) Lookup(identity *authnapi.Identity, resourceType string, resourceId models.IDType) (api.Producer, error) {
	return m.Producer, nil
}
