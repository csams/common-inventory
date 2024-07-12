package stdout

import (
	"encoding/json"
	"os"
)

type StdOutProducer struct {
	Encoder *json.Encoder
}

func New() *StdOutProducer {
	return &StdOutProducer{
		Encoder: json.NewEncoder(os.Stdout),
	}
}

func (p *StdOutProducer) Produce(e interface{}) error {
	return p.Encoder.Encode(e)
}
