package eventuate

import "github.com/shopcookeat/eventuate-client-golang/future"

type EventResultHandler func(interface{}, *EventMetadata) future.Settler
type EventResultHandlerMap map[string]map[string]EventResultHandler
type DispatcherMaker func(eventHandlers *EventResultHandlerMap) (Dispatcher, error)

func NewEventResultHandlerMap() *EventResultHandlerMap {
	tmp := EventResultHandlerMap(make(map[string]map[string]EventResultHandler))
	return &tmp
}

func (handlersMap *EventResultHandlerMap) AddHandler(aggType, eventType string, handler EventResultHandler) *EventResultHandlerMap {
	aggregate, hasAggregate := (*handlersMap)[aggType]
	if !hasAggregate {
		aggregate = make(map[string]EventResultHandler)
		(*handlersMap)[aggType] = aggregate
	}
	aggregate[eventType] = handler
	return handlersMap
}

func (handlersMap *EventResultHandlerMap) GetHandler(aggType, eventType string) (*EventResultHandler, error) {
	_, hasEntityType := (*handlersMap)[aggType]
	if !hasEntityType {
		return nil, AppError("Event handler for entity type %s is not registered", aggType)
	}

	handler, hasEventType := (*handlersMap)[aggType][eventType]
	if !hasEventType {
		return nil, AppError("Event handler for entity type/event type %s / %s is not registered",
			aggType, eventType)
	}

	return &handler, nil
}

func  (handlersMap *EventResultHandlerMap) transformToEntityEventGroups() map[string][]string {
	result := make(map[string][]string)
	for key1, val1 := range *handlersMap {
		if _, hasKey := result[key1]; !hasKey {
			result[key1] = []string{}
		}
		for key2 := range val1 {
			result[key1] = append(result[key1], key2)
		}
	}
	return result
}

