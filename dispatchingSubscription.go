package eventuate

import (
	"github.com/eventuate-clients/eventuate-client-golang/future"
)

type DispatchingSubscription struct {
	*Subscription
	eventHandlers *EventResultHandlerMap
	//subscription  *Subscription
	typeHints TypeHintMapper
	//mgr *subscriptionManager
	//ll            loglib.LogLevelEnum
	//lg            loglib.Logger
}

func (sub *DispatchingSubscription) dispatchEvent(evt StompEvent) {
	//sub := sub.subscription
	defer func() {
		if r := recover(); r != nil {
			err := AppError(
				"Recovered in event handler: %#v", r)
			sub.handleEventHandlerResults(&evt, nil, err)
		}
	}()

	var (
		result     future.Settler
		evtHandler EventResultHandler
	)

	evtHandler = *sub.eventHandler
	result = evtHandler(NewEventMetadataFromStomp(&evt, sub.typeHints))

	if result.IsSettled() {
		val, err := result.GetValue()
		sub.handleEventHandlerResults(&evt, val, err)

	} else {
		go func() {
			val, err := result.GetValue() // blocks
			sub.handleEventHandlerResults(&evt, val, err)
		}()
	}
}

func (sub *DispatchingSubscription) handleEventHandlerResults(evt *StompEvent, val interface{}, err error) {
	sub.lg.Printf("handleEventHandlerResults. subscription #%s, event: %v\nValue: %v\nError: %#v",
		sub.Id, evt, val, err)

	if err != nil {
		sub.lg.Printf("Failing event: %v\n", evt)

		sub.subscriptionErrors <- AppError(
			"Failed handler for subscription #%s for event: %#v\nError: %#v",
			sub.Id, evt, err)
		return
	}
	sub.lg.Printf("Acknowledging event: %v\n", evt)
	sub.AcknowledgeEvent(evt)
}
