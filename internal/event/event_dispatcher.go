package event

type EventListener interface {
	OnEventDispatched(eventID string, request EventDispatchRequest)
	OnEventBound(binding EventBinding)
	OnEventUnbound(bindingID string)
}
