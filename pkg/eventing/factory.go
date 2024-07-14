package eventing

import (
	"fmt"
	"log/slog"

	"github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/eventing/kafka"
	"github.com/csams/common-inventory/pkg/eventing/stdout"
)

func New(c CompletedConfig, log *slog.Logger) (api.Manager, error) {

	switch c.Eventer {
	case "stdout":
		return stdout.New()
	case "kafka":
		return kafka.New(c.Kafka, log)
	}
	return nil, fmt.Errorf("unrecognized eventer type: %s", c.Eventer)

}
