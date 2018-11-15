package proto2

type EventSender interface {
	Send(eventType EventType)
}
