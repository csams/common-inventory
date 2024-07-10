package api

// this should be based on cloudevents

type Event[T any] struct {
	Headers map[string]string

	EventType    string
	ResourceType string
	Object       T
}
