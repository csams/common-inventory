package stdout

import (
	"context"
	"encoding/json"
	"os"

	"github.com/csams/common-inventory/pkg/eventing/api"
)

type StdOutProducer struct {
	Encoder *json.Encoder
}

func New() api.Producer {
	return &StdOutProducer{
		Encoder: json.NewEncoder(os.Stdout),
	}
}

func (p *StdOutProducer) Produce(ctx context.Context, event *api.Event) error {
	return p.Encoder.Encode(event)
}
