package eventuate

import (
	"github.com/eventuate-clients/eventuate-client-golang/future"
)

type subscriptionManager struct {
	*StompClient
}

func newSubscriptionManager(client *StompClient) *subscriptionManager {
	return &subscriptionManager{
		StompClient: client}
}

func (mgr *subscriptionManager) Subscribe(
	subscriberId string,
	eventHandlers *EventResultHandlerMap,
	subscriberOptions *SubscriberOptions,
	useSwimlane bool) (*DispatchingSubscription, error) {

	dspMaker := DispatcherMaker(NewEventDispatcher)
	if useSwimlane {
		dspMaker = DispatcherMaker(NewEventTypeSwimlaneDispatcher)
	}
	return mgr.subscribeForStrategy(
		subscriberId,
		eventHandlers,
		subscriberOptions,
		dspMaker)
}

func (mgr *subscriptionManager) subscribeForStrategy(
	subscriberId string,
	eventHandlers *EventResultHandlerMap,
	subscriberOptions *SubscriberOptions,
	dispatcherMaker DispatcherMaker) (*DispatchingSubscription, error) {

	dispatcher, err := dispatcherMaker(eventHandlers)
	if err != nil {
		return nil, err
	}

	evtHandler := EventResultHandler(func(data interface{}, meta *EventMetadata) future.Settler {
		mgr.lg.Printf("SUBS_MANAGER: handling event: %v, (%#v)\n", meta.Id, meta)
		defer mgr.lg.Printf("SUBS_MANAGER: handling event (finished): %v\n", meta.Id)

		return dispatcher.Dispatch(data, meta)
	})

	entityEventsGroup := eventHandlers.transformToEntityEventGroups()

	sub, subErr := mgr.StompClient.Subscribe(subscriberId, entityEventsGroup, subscriberOptions, &evtHandler)
	if subErr != nil {
		return nil, subErr
	}

	msub := &DispatchingSubscription{
		Subscription:  sub,
		eventHandlers: eventHandlers,
		typeHints:     mgr.typeHints}

	go func(sub *Subscription) {
		for evt := range sub.incomingEvent {
			msub.lg.Printf("Received event via STOMP chan: %v (before dispatching)\n", evt.String())
			msub.dispatchEvent(evt)
		}
	}(sub)

	return msub, nil
}
