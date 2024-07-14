package eventing

import (
	"context"

	authnapi "github.com/csams/common-inventory/pkg/authn/api"
	"github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/eventing/stdout"
	"github.com/csams/common-inventory/pkg/models"
)

type EventingManager struct {
	Producer api.Producer
	Errors   chan error
}

func New() *EventingManager {
	return &EventingManager{
		Producer: stdout.New(),
		Errors:   make(chan error),
	}
}

func (m *EventingManager) Lookup(identity *authnapi.Identity, model *models.Resource) (api.Producer, error) {
	return m.Producer, nil
}

func (m *EventingManager) Errs() <-chan error {
	return m.Errors
}

func (m *EventingManager) Shutdown(ctx context.Context) error {
	return nil
}
