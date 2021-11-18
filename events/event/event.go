package event

import "time"

// Event событие системы
type Event struct {
	Type string
	Name string
	Data string
	Time int64
}

// EventType ....
//type EventType string

// New ...
func New(eventType, eventName string, eventTime int64, eventData string) *Event {

	if eventTime == 0 {
		eventTime = time.Now().UnixNano()
	}

	return &Event{
		Type: eventType,
		Name: eventName,
		Data: eventData,
		Time: eventTime,
	}

}
