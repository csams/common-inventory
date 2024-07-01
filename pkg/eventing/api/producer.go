package api

type Producer interface {
    Produce(event interface{}) error
}
