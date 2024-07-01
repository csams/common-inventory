package api

// this should be based on cloudevents

type Event[T any] struct {
    EventType string
    ResourceType string
    Object T
}
