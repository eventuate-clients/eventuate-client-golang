package eventuate

import (
	"reflect"

	"github.com/gmallard/stompngo"
	"github.com/eventuate-clients/eventuate-client-golang/future"
)

type Crud interface {
	Find(
		aggregateType string,
		entityId Int128,
		findOptions *AggregateCrudFindOptions) (*LoadedEvents, error)
	Save(
		aggregateType string,
		events []EventTypeAndData,
		saveOptions *AggregateCrudSaveOptions) (*EntityIdVersionAndEventIds, error)
	Update(
		entityIdAndType EntityIdAndType,
		entityVersion Int128,
		events []EventTypeAndData,
		updateOptions *AggregateCrudUpdateOptions) (*EntityIdVersionAndEventIds, error)
}

type Repository interface {
	Save(cmd Command) (*EntityMetadata, error)
	Update(entityId Int128, cmd Command) (*EntityMetadata, error)
	Find(entityId Int128) (*EntityMetadata, error)
}

type Dispatcher interface {
	Dispatch(interface{}, *EventMetadata) future.Settler
}

type Subscriber interface {
	Subscribe(
		subscriberId string,
		aggregatesAndEvents map[string][]string,
		subscriberOptions *SubscriberOptions,
		handler *EventResultHandler) (*Subscription, error)

	SubscribeAndDispatch(
		subscriberId string,
		eventHandlers *EventResultHandlerMap,
		subscriberOptions *SubscriberOptions,
		useSwimlane bool) (*DispatchingSubscription, error)
}

type SubscriberDispatcher interface {
	Subscribe(
		subscriberId string,
		eventHandlers *EventResultHandlerMap,
		subscriberOptions *SubscriberOptions,
		useSwimlane bool) (*DispatchingSubscription, error)
}

type AggregateStore interface {
	Save(
		class reflect.Type,
		events []interface{},
		saveOptions *interface{}) (*EntityIdAndVersion, error)
	Find(
		class reflect.Type,
		entityId string,
		findOptions *interface{}) (*EntityWithMetadata, error)
	Update(
		class reflect.Type,
		entityIdAndVersion EntityIdAndType,
		events []interface{},
		updateOptions *interface{}) (*EntityIdAndVersion, error)
	Subscribe(
		subscriberId string,
		aggregatesAndEvents map[string]interface{},
		subscriberOptions interface{},
		dispatch func() interface{})
	MaybeSnapshot(
		aggregate *Aggregate,
		snapshotVersion *Int128,
		oldEvents []EventWithMetadata,
		newEvents []interface{})
	FromSnapshot(
		class reflect.Type,
		snapshot Snapshot) *Aggregate
}

type Acker interface {
	Ack(stompngo.Headers) error
}

type TypeHintMapper interface {
	//MakeCopy() TypeHintMapper
	HasEventType(name string) bool
	GetEventType(name string) reflect.Type
	GetTypeByKeyName(name string) (bool, reflect.Type)
	GetTypeByTypeName(typeName string) (bool, string)
}
type Unsubscriber interface {
	Unsubscribe() error
}

type TypeHintRegisterer interface {
	RegisterEventType(name string, typeInstance interface{}) error
}

type CommandProcessor interface {
	ProcessCommand(command Command) []Event
}

type EventApplier interface {
	ApplyEvent(evt Event) EventApplier
}
