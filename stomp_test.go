package eventuate_test

import (
	"fmt"
	"os"
	//"sync"
	"testing"

	//"github.com/gmallard/stompngo"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
	"github.com/shopcookeat/eventuate-client-golang"
)

func TestStompJsonSchemas(t *testing.T) {
	t.Parallel()

	workDirectory, workDirectoryErr := os.Getwd()
	if workDirectoryErr != nil {
		t.Fatal()
	}

	loader := gojsonschema.NewGoLoader(
		eventuate.SubscriptionRequest{
			SubscriberID:         "stompSubscriberID",
			EntityTypesAndEvents: make(map[string]string),
		},
	)

	schemaLoader := gojsonschema.NewReferenceLoader(
		fmt.Sprintf("file://%s/json_schema_test_suite/stomp/destination-header.json", workDirectory))

	var destinationHeaderError error

	checkDestinationHeader, destinationHeaderError := gojsonschema.Validate(schemaLoader, loader)

	assertNoError(t, destinationHeaderError)

	assert.Equal(t, true, checkDestinationHeader.Valid())

	loader = gojsonschema.NewGoLoader(
		eventuate.StompEvent{
			Id:         eventuate.Int128FromString("000001588bce5f44-0242ac1101020002"),
			EventType:  "net.chrisrichardson.eventstore.example.MyEntity",
			EventData:  `{"name": "John"}`,
			EntityId:   eventuate.Int128FromString("000001588bce5f47-0242ac1100dc0000"),
			EntityType: "net.chrisrichardson.eventstore.example.MyEntityUpdate",
			EventToken: "JNKkjf23enjdsv",
			Swimlane:   1,
			Offset:     1,
		},
	)

	schemaLoader = gojsonschema.NewReferenceLoader(
		fmt.Sprintf("file://%s/json_schema_test_suite/stomp/stomp-Event.json", workDirectory))

	var stompEventErr error

	checkStompEvent, stompEventErr := gojsonschema.Validate(schemaLoader, loader)

	assertNoError(t, stompEventErr)

	assert.Equal(t, true, checkStompEvent.Valid())

}

//func TestRealServerStomp(t *testing.T) {
//	stomp := ClientFromEnv()
//
//	conn, err := stomp.ConnectToStompServer()
//	assertNoError(t, err)
//
//	assert.Equal(t, true, conn.Connected())
//
//	uuid := stompngo.Uuid()
//
//	entityTypeName := fmt.Sprintf("net.chrisrichardson.eventstore.example.MyEntity-%s", uuid)
//
//	createEventType := "net.chrisrichardson.eventstore.example.MyEntityWasCreated"
//
//	createResponse, createEntityErr := stomp.CreateEntity(CreateEntityRequest{
//		Events: []EventTypeAndData{
//			{
//				EventType: createEventType,
//				EventData: `{"name":"Fred"}`}},
//		EntityTypeName: entityTypeName})
//	assertNoError(t, createEntityErr)
//
//	updateEventType := "net.chrisrichardson.eventstore.example.MyEntityWasUpdated"
//
//	_, updateEntityErr := stomp.UpdateEntity(
//		entityTypeName,
//		createResponse.EntityId,
//		UpdateEntityRequest{
//			Events: []EventTypeAndData{
//				{
//					EventType: updateEventType,
//					EventData: `{"name": "John"}`}},
//			EntityVersion: createResponse.EntityVersion})
//	assertNoError(t, updateEntityErr)
//
//	entityTypesAndEvents := make(map[string][]string)
//
//	entityTypesAndEvents[entityTypeName] = []string{createEventType, updateEventType}
//
//	subscription, err := stomp.Subscribe(&SubscriptionRequest{
//		SubscriberID:         fmt.Sprintf("subscriber-%s", uuid),
//		EntityTypesAndEvents: entityTypesAndEvents})
//	assertNoError(t, err)
//
//	eventFirst, err := subscription.ReadEvent()
//	assertNoError(t, err)
//
//	assert.Equal(t, "net.chrisrichardson.eventstore.example.MyEntityWasCreated", eventFirst.EventType)
//
//	assert.Equal(t, `{"name":"Fred"}`, eventFirst.EventData)
//
//	eventSecond, err := subscription.ReadEvent()
//	assertNoError(t, err)
//
//	assert.Equal(t, `{"name": "John"}`, eventSecond.EventData)
//
//	assert.Equal(t, "net.chrisrichardson.eventstore.example.MyEntityWasUpdated", eventSecond.EventType)
//
//	subscription.AcknowledgeEvent(eventSecond)
//
//	assert.Equal(t, 2, subscription.FetchPendingsCount())
//
//	subscription.AcknowledgeEvent(eventFirst)
//
//	assert.Equal(t, 0, subscription.FetchPendingsCount())
//}
//
//func TestRealServerStompParallelAck(t *testing.T) {
//	stomp := ClientFromEnv()
//
//	stomp.ConnectToStompServer()
//
//	uuid := stompngo.Uuid()
//
//	entityTypeName := fmt.Sprintf("net.chrisrichardson.eventstore.example.MyEntity-%s", uuid)
//
//	createEventType := "net.chrisrichardson.eventstore.example.MyEntityWasCreated"
//
//	const number_of_events = 6
//
//	for number := 0; number < number_of_events; number++ {
//		stomp.CreateEntity(CreateEntityRequest{
//			Events: []EventTypeAndData{
//				{
//					EventType: createEventType,
//					EventData: `{"name":"Fred"}`}},
//			EntityTypeName: entityTypeName})
//	}
//
//	entityTypesAndEvents := make(map[string][]string)
//
//	entityTypesAndEvents[entityTypeName] = []string{createEventType}
//
//	subscription, subscribeErr := stomp.Subscribe(&SubscriptionRequest{
//		SubscriberID:         fmt.Sprintf("subscriber-%s", uuid),
//		EntityTypesAndEvents: entityTypesAndEvents,
//	})
//
//	assertNoError(t, subscribeErr)
//
//	eventsList := subscription.readNEventsSequentially(number_of_events)
//
//	assert.Equal(t, number_of_events, subscription.FetchPendingsCount())
//
//	var wg sync.WaitGroup
//
//	for _, evt := range eventsList {
//		wg.Add(1)
//		go func(evt *StompEvent) {
//			defer wg.Done()
//			subscription.AcknowledgeEvent(evt)
//		}(evt)
//	}
//
//	wg.Wait()
//
//	assert.Equal(t, 0, subscription.FetchPendingsCount())
//}
//
//func TestRealServerStompParallelAckAndRead(t *testing.T) {
//	stomp := ClientFromEnv()
//
//	stomp.ConnectToStompServer()
//
//	uuid := stompngo.Uuid()
//
//	entityTypeName := fmt.Sprintf("net.chrisrichardson.eventstore.example.MyEntity-%s", uuid)
//
//	createEventType := "net.chrisrichardson.eventstore.example.MyEntityWasCreated"
//
//	const number_of_threads = 6
//
//	for i := 0; i < number_of_threads; i++ {
//		stomp.CreateEntity(CreateEntityRequest{
//			Events: []EventTypeAndData{
//				{EventType: createEventType, EventData: `{"name":"Fred"}`}},
//			EntityTypeName: entityTypeName})
//	}
//
//	entityTypesAndEvents := make(map[string][]string)
//
//	entityTypesAndEvents[entityTypeName] = []string{createEventType}
//
//	subscription, subscribeErr := stomp.Subscribe(&SubscriptionRequest{
//		SubscriberID:         fmt.Sprintf("subscriber-%s", uuid),
//		EntityTypesAndEvents: entityTypesAndEvents,
//	})
//
//	assertNoError(t, subscribeErr)
//
//	var wg sync.WaitGroup
//
//	for i := 0; i < number_of_threads; i++ {
//		wg.Add(1)
//
//		go func(sub *Subscription) {
//
//			defer wg.Done()
//			evt, readErr := sub.ReadEvent()
//			assertNoError(t, readErr)
//
//			sub.AcknowledgeEvent(evt)
//
//		}(subscription)
//	}
//
//	wg.Wait()
//
//	assert.Equal(t, 0, subscription.FetchPendingsCount())
//}
//
//
//func TestpendingAcknowledgesBAC(t *testing.T) {
//	t.Parallel()
//
//	eventA, eventB, eventC, pendings := CreateThreeEventsAndPendings()
//
//	assert.Equal(t, []pendingAcknowledge{}, pendings.ack(eventB))
//
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "a", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "b", Acked: true, AckHeader: ""},
//			pendingAcknowledge{EventID: "c", Acked: false, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "a", Acked: true, AckHeader: ""},
//		pendingAcknowledge{EventID: "b", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventA))
//
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "c", Acked: false, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventC))
//
//	assert.Equal(t, pendingAcknowledges{queue: []pendingAcknowledge{}}, pendings)
//}
//
//func TestpendingAcknowledgesABC(t *testing.T) {
//	t.Parallel()
//
//	eventA, eventB, eventC, pendings := CreateThreeEventsAndPendings()
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "a", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventA))
//
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "b", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "c", Acked: false, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "b", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventB))
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "c", Acked: false, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventC))
//	assert.Equal(t, pendingAcknowledges{queue: []pendingAcknowledge{}}, pendings)
//}
//
//func TestpendingAcknowledgesCBA(t *testing.T) {
//	t.Parallel()
//
//	eventA, eventB, eventC, pendings := CreateThreeEventsAndPendings()
//
//	assert.Equal(t, []pendingAcknowledge{}, pendings.ack(eventC))
//
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "a", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "b", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{}, pendings.ack(eventB))
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "a", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "b", Acked: true, AckHeader: ""},
//			pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "a", Acked: true, AckHeader: ""},
//		pendingAcknowledge{EventID: "b", Acked: true, AckHeader: ""},
//		pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventA))
//	assert.Equal(t, pendingAcknowledges{queue: []pendingAcknowledge{}}, pendings)
//}
//
//func TestpendingAcknowledgesCAB(t *testing.T) {
//	t.Parallel()
//
//	eventA, eventB, eventC, pendings := CreateThreeEventsAndPendings()
//
//	assert.Equal(t, []pendingAcknowledge{}, pendings.ack(eventC))
//
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "a", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "b", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "a", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventA))
//	assert.Equal(t, pendingAcknowledges{
//		queue: []pendingAcknowledge{
//			pendingAcknowledge{EventID: "b", Acked: false, AckHeader: ""},
//			pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//		},
//	}, pendings)
//
//	assert.Equal(t, []pendingAcknowledge{
//		pendingAcknowledge{EventID: "b", Acked: true, AckHeader: ""},
//		pendingAcknowledge{EventID: "c", Acked: true, AckHeader: ""},
//	}, pendings.ack(eventB))
//	assert.Equal(t, pendingAcknowledges{queue: []pendingAcknowledge{}}, pendings)
//}
