package eventuate

import (
	"github.com/eventuate-clients/eventuate-client-golang/future"
	loglib "github.com/eventuate-clients/eventuate-client-golang/logger"
)

type eventPack struct {
	eventData interface{}
	eventMeta *EventMetadata
	fr        *future.Result
}

type eventLane chan eventPack

type EventTypeSwimlaneDispatcher struct {
	EventDispatcher
	queues map[string]map[int]eventLane
	ll     loglib.LogLevelEnum
	lg     loglib.Logger
}

func NewEventTypeSwimlaneDispatcher(eventHandlers *EventResultHandlerMap) (Dispatcher, error) {
	result := &EventTypeSwimlaneDispatcher{
		EventDispatcher: EventDispatcher{
			handlers: eventHandlers,
		},
		queues: make(map[string]map[int]eventLane),
		lg:     loglib.NewNilLogger()}
	return result, nil
}

func (dsp *EventTypeSwimlaneDispatcher) SetLogLevel(level loglib.LogLevelEnum) *EventTypeSwimlaneDispatcher {
	dsp.ll = level
	dsp.lg = loglib.NewLogger(level)
	return dsp
}

func (dsp *EventTypeSwimlaneDispatcher) Dispatch(data interface{}, evt *EventMetadata) future.Settler {
	swimlaneQ, haveSwimalneQ := dsp.queues[evt.EntityType]
	if !haveSwimalneQ {
		swimlaneQ = make(map[int]eventLane)
		dsp.queues[evt.EntityType] = swimlaneQ
	}

	q, haveQ := dsp.queues[evt.EntityType][evt.SwimLane]
	if !haveQ {
		q = make(eventLane, 16)
		dsp.queues[evt.EntityType][evt.SwimLane] = q

		go func() {
			defer func() {
				close(q)
				delete(dsp.queues[evt.EntityType], evt.SwimLane)
			}()
			for {
				pack := <-q
				fr := pack.fr
				meta := pack.eventMeta
				data := pack.eventData
				if fr == nil || evt == nil {
					return
				}

				dsp.lg.Println("go EventTypeSwimlaneDispatcher.Dispatch (go). Before calling a handler for EntityType-EventType", evt.EntityType, evt.EventType)
				rslt := dsp.EventDispatcher.Dispatch(data, meta)
				dsp.lg.Println("go EventTypeSwimlaneDispatcher.Dispatch (go). FR await ", rslt, " => ", *fr)
				value, err := rslt.GetValue() // blocking here
				dsp.lg.Println("go EventTypeSwimlaneDispatcher.Dispatch (go). Result ", rslt, " received ")
				(*fr).Settle(value, err)
				dsp.lg.Println("go EventTypeSwimlaneDispatcher.Dispatch (go). resuming loop")
			}
		}()
	}

	result := future.NewResult()
	result.SetLogLevel(dsp.ll)
	dsp.lg.Println("EventTypeSwimlaneDispatcher.Dispatch. New: ", result)

	q <- eventPack{
		eventData:data,
		eventMeta: evt,
		fr:    result}

	return future.Settler(result)
}
