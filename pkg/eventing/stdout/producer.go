package stdout

import "fmt"

type StdOutProducer struct{}

func New() *StdOutProducer {
	return &StdOutProducer{}
}

func (p *StdOutProducer) Produce(e interface{}) error {
	msg := fmt.Sprintf("%v", e)
	println(msg)
	return nil
}
