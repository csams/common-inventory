package api

// this should be based on cloudevents

type Event struct {
	EventType    string
	ResourceType string
	Object       interface{}
}
