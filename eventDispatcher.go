package eventuate

import "github.com/eventuate-clients/eventuate-client-golang/future"
import loglib "github.com/eventuate-clients/eventuate-client-golang/logger"

type EventDispatcher struct {
	handlers *EventResultHandlerMap
	lg       loglib.Logger
}

func NewEventDispatcher(eventHandlers *EventResultHandlerMap) (Dispatcher, error) {
	result := &EventDispatcher{
		handlers: eventHandlers,
		lg:       loglib.NewNilLogger()}
	return result, nil
}

func (dsp *EventDispatcher) Dispatch(data interface{}, evt *EventMetadata) future.Settler {
	handler, handlerErr := dsp.handlers.GetHandler(evt.EntityType, evt.EventType)
	if handlerErr != nil {
		return future.NewFailure(handlerErr)
	}

	tmp := *handler
	return tmp(data, evt)
}
