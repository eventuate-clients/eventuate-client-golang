package eventuate

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type EventMetadata struct {
	Id           Int128
	EntityId     Int128
	EntityType   string
	EventType    string
	SwimLane     int
	Offset       int
	EventContext EventContext
}

func (meta *EventMetadata) String() string {
	var tmp struct {
		Id         Int128 `json:"id"`
		EventType  string `json:"eventType"`
		EntityId   Int128 `json:"entityId"`
		EntityType string `json:"entityType"`
	} = struct {
		Id         Int128 `json:"id"`
		EventType  string `json:"eventType"`
		EntityId   Int128 `json:"entityId"`
		EntityType string `json:"entityType"`
	}{
		Id:         meta.Id,
		EventType:  meta.EventType,
		EntityId:   meta.EntityId,
		EntityType: meta.EntityType}
	b, _ := json.Marshal(tmp)
	return fmt.Sprintf("<StompEvent: %v/>", string(b))
}


func NewEventMetadataFromStomp(evt *StompEvent, hintsMap TypeHintMapper) (interface{}, *EventMetadata) {

	var evtData interface{}

	if hintsMap.HasEventType(evt.EventType) {
		source := []byte(evt.EventData)
		destination := reflect.New(getUnderlyingType(hintsMap.GetEventType(evt.EventType))).Interface()
		evtDataErr := json.Unmarshal(source, destination)
		if evtDataErr == nil {
			evtData = destination
		}
	}

	entityTypeParts := strings.Split(evt.EntityType, "/")
	return evtData, &EventMetadata{
		Id:           evt.Id,
		EntityId:     evt.EntityId,
		EntityType:   entityTypeParts[len(entityTypeParts)-1],
		EventType:    evt.EventType,
		SwimLane:     evt.Swimlane,
		Offset:       evt.Offset,
		EventContext: EventContext(evt.EventToken)}
}

// StompEvent is the struct for stomp Event
type StompEvent struct {
	Id         Int128 `json:"id"`
	EventType  string `json:"eventType"`
	EventData  string `json:"eventData"`
	EntityId   Int128 `json:"entityId"`
	EntityType string `json:"entityType"`
	EventToken string `json:"eventToken"`
	Swimlane   int    `json:"swimlane"`
	Offset     int    `json:"offset"`
}

func (stompEvent *StompEvent) String() string {
	var tmp struct {
		Id         Int128 `json:"id"`
		EventType  string `json:"eventType"`
		EventData  string `json:"eventData"`
		EntityId   Int128 `json:"entityId"`
		EntityType string `json:"entityType"`
	} = struct {
		Id         Int128 `json:"id"`
		EventType  string `json:"eventType"`
		EventData  string `json:"eventData"`
		EntityId   Int128 `json:"entityId"`
		EntityType string `json:"entityType"`
	}{
		Id:         stompEvent.Id,
		EventType:  stompEvent.EventType,
		EventData:  stompEvent.EventData,
		EntityId:   stompEvent.EntityId,
		EntityType: stompEvent.EntityType}
	b, _ := json.Marshal(tmp)
	return fmt.Sprintf("<StompEvent: %v/>", string(b))
}

type EntityIdAndType struct {
	EntityType string
	EntityId   Int128
}

type EntityIdAndVersion struct {
	EntityId      Int128
	EntityVersion Int128
}

type AggregateCrudSaveOptions struct {
	EventMetadata   *string
	TriggeringEvent *EventContext
	EntityId        Int128
}

type AggregateCrudFindOptions struct {
	TriggeringEvent *EventContext
}

type AggregateCrudUpdateOptions struct {
	TriggeringEvent    *EventContext
	EventMetadata      *string
	SerializedSnapshot *SerializedSnapshot
}

// TODO: implement
type Snapshot interface{}

// TODO: implement
type Aggregate interface{}

// TODO: implement
type EntityWithMetadata interface{}

// TODO: implement
type EventWithMetadata interface{}

type EventContext string

func (ctx EventContext) String() string {
	return string(ctx)
}

type SubscriberOptions struct {
	Durability            SubscriberDurability      // DURABLE
	ReadFrom              SubscriberInitialPosition // BEGINNING
	ProgressNotifications bool                      // false
}

type SubscriberDurability int
type SubscriberInitialPosition int

const (
	DURABLE SubscriberDurability = iota
	TRANSIENT
)

const (
	BEGINNING SubscriberInitialPosition = iota
	END
)

type SerializedSnapshotWithVersion struct {
	SerializedSnapshot *SerializedSnapshot
	EntityVersion      Int128
}

type SerializedSnapshot struct {
	SnapshotType string
	Json         string
}

// EventIdTypeAndData is the struct for EventID, EventType and EventData
type EventIdTypeAndData struct {
	EventId Int128 `json:"id"`
	EventTypeAndData
}

func (evt EventIdTypeAndData) String() string {
	b, _ := json.Marshal(evt)
	return fmt.Sprintf("<EventIdTypeAndData: %v />", string(b))
}

func (rcv *EventIdTypeAndData) ToEventTypeAndData() EventTypeAndData {
	return EventTypeAndData{
		EventType: rcv.EventType,
		EventData: rcv.EventData}
}

// EventTypeAndData is the struct for EventType and EventData
type EventTypeAndData struct {
	EventType string `json:"eventType"`
	EventData string `json:"eventData"`
}

// CreateEntityRequest is the struct for create entity request
type CreateEntityRequest struct {
	EntityTypeName string             `json:"entityTypeName"`
	Events         []EventTypeAndData `json:"events"`
}

// UpdateEntityRequest is the struct for entity request
type UpdateEntityRequest struct {
	Events        []EventTypeAndData `json:"events"`
	EntityVersion string             `json:"entityVersion"`
}

type LoadedEvents struct {
	Events   []EventIdTypeAndData           `json:"events"`
	Snapshot *SerializedSnapshotWithVersion `json:"snapshot"`
}

type EntityIdVersionAndEventIds struct {
	EntityId      Int128   `json:"entityId"`
	EntityVersion Int128   `json:"entityVersion"`
	EventIds      []Int128 `json:"eventIds"`
}

// SubscriptionRequest is the struct for subscription request
type SubscriptionRequest struct {
	EntityTypesAndEvents interface{} `json:"entityTypesAndEvents"`
	SubscriberID         string      `json:"subscriberId"`
	Space                string      `json:"space"`
}

// CreateResponse is the struct for create response
type CreateResponse EntityIdVersionAndEventIds

// GetResponse is the struct for response to GET requests
type GetResponse LoadedEvents

// UpdateResponse is the struct for update response
type UpdateResponse EntityIdVersionAndEventIds

// https://github.com/eventuate-clients/eventuate-client-java/blob/d75fc7f4237bcb6d8f6572dd280804cf0ff08a94/eventuate-client-java/src/main/java/io/eventuate/DispatchedEvent.java
type DispatchedEvent struct {
	//String entityId, Int128 eventId, T Event, Integer swimlane, Long offset, EventContext eventContext
	EntityId     string
	EventId      Int128
	Event        *Event
	SwimLane     int
	Offset       int
	EventContext EventContext
}

type Command interface{}

type Event interface{}
