package eventuate_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	//"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/eventuate-clients/eventuate-client-golang"
)

const NAMESPACE = "ns12345678"

const ENTITY_TYPE = "net.chrisrichardson.eventstore.example.MyEntity-12345"
const ENTITY_ID = "0000015a822f7197-0242ac1100db0000"
const EVENT_CREATED = "net.chrisrichardson.eventstore.example.MyEntityWasCreated"
const EVENT_CHANGED = "net.chrisrichardson.eventstore.example.MyEntityNameChanged"

const EVENT_ID_1 = "000001588bce5f44-0242ac1101020002"
const EVENT_DATA_1 = `{"name":"Arthur Dent"}`
const EVENT_DATA_1_ESC = `{\"name\":\"Arthur Dent\"}`

const EVENT_ID_2 = "0000015a8a9019ce-0242ac1101020002"
const EVENT_DATA_2 = `{"name":"Zaphod Beeblebrox"}`
const EVENT_DATA_2_ESC = `{\"name\":\"Zaphod Beeblebrox\"}`

var (
	hasUpdated bool = false
)

func ExampleNewRESTClient_fromEnv() {
	client, clientErr := eventuate.ClientBuilder().WithCredentials("123", "456").BuildREST()
	if clientErr != nil {
		log.Fatal(clientErr)
	}

	fmt.Println(client.Url.String(), client.Credentials.Space)

	// Output: https://api.eventuate.io default
}

func ExampleNewRESTClient_allSpecified() {
	client, clientErr := eventuate.ClientBuilder().WithCredentials("123", "456").WithSpace("789").BuildREST()
	if clientErr != nil {
		log.Fatal(clientErr)
	}

	fmt.Println(client.Url.String(), client.Credentials.Space)

	// Output: https://api.eventuate.io 789
}

func ExampleNewRESTClient_usualWay() {
	client, clientErr := eventuate.ClientBuilder().WithCredentials("123", "456").BuildREST()

	fmt.Println(clientErr, client.Url.String(), client.Credentials.Space)

	// Output: <nil> https://api.eventuate.io default
}

func TestNewRESTClient(t *testing.T) {
	_, clientErr := eventuate.ClientBuilder().WithCredentials("123", "456").BuildREST()
	if clientErr != nil {
		t.Fail()
		return
	}
}

func TestRESTClient_Save(t *testing.T) {

	// https://api.eventuate.io/entity/ns12345678/

	//{"entityTypeName":"net.chrisrichardson.eventstore.example.MyEntity-12345","events":[{
	//"eventType":  "net.chrisrichardson.eventstore.example.MyEntityWasCreated",
	// "eventData": "{\"name\":\"Fred\"}"
	//}]}

	// {"EntityId":"0000015a822f7197-0242ac1100db0000",
	// "entityVersion":"0000015a822f7198-0242ac1101020002","eventIds":["0000015a822f7198-0242ac1101020002"]}

	_, server := getTestServerWithRoutes(t)
	defer server.Close()

	client := getRestClient(t, server.URL, NAMESPACE)
	if client == nil {
		t.Fatal("REST Client is not initialized")
	}

	createResponse, createResponseError := client.Save(
		ENTITY_TYPE,
		[]eventuate.EventTypeAndData{
			{
				EventType: EVENT_CREATED,
				EventData: EVENT_DATA_1}},
		nil)

	assertNoError(t, createResponseError)

	tmp := eventuate.EntityIdVersionAndEventIds(eventuate.CreateResponse{
		EntityId:      eventuate.Int128FromString(ENTITY_ID),
		EntityVersion: eventuate.Int128FromString(EVENT_ID_1),
		EventIds:      []eventuate.Int128{eventuate.Int128FromString(EVENT_ID_1)}})
	expected := &tmp

	assert.Equal(t, expected, createResponse)
}

func TestRESTClient_Find(t *testing.T) {

	// https://api.eventuate.io/entity/ns12345678/net.chrisrichardson.eventstore.example.MyEntity-12345/0000015a822f7197-0242ac1100db0000
	//httpmock.RegisterResponder("GET",
	//	fmt.Sprintf(`https://api.eventuate.io/entity/%s/%s/%s`, NAMESPACE, EntityType, EntityId),
	//	httpmock.NewStringResponder(200, `{231"events":[{"id":"0000015a822f7198-0242ac1101020002","eventType":"net.chrisrichardson.eventstore.example.MyEntityWasCreated","eventData":"{\"name\":\"Fred\"}"}],"triggeringEvents":[]}`))

	_, server := getTestServerWithRoutes(t)
	defer server.Close()

	client := getRestClient(t, server.URL, NAMESPACE)
	if client == nil {
		t.Fatal("REST Client is not initialized")
	}

	events, err := client.Find(ENTITY_TYPE, eventuate.Int128FromString(ENTITY_ID), nil)
	assertNoError(t, err)

	if events == nil {
		t.Fatal("events is nil")
	}

	if len(events.Events) != 1 {
		t.Error("events.Events, expecting array of one")
	}

	getResponseExpected := &eventuate.LoadedEvents{
		Events: []eventuate.EventIdTypeAndData{
			{
				EventId: eventuate.Int128FromString(EVENT_ID_1),
				EventTypeAndData: eventuate.EventTypeAndData{
					EventType: EVENT_CREATED,
					EventData: EVENT_DATA_1}}}}

	assert.Equal(t, getResponseExpected, events)
}

func TestRESTClient_Update(t *testing.T) {

	//https://api.eventuate.io/entity/ns12345678/net.chrisrichardson.eventstore.example.MyEntity-12345/0000015a822f7197-0242ac1100db0000

	//{"entityVersion":"0000015a822f7197-0242ac1100db0000","events":[{
	//"eventType":  "net.chrisrichardson.eventstore.example.MyEntityWasChanged", "eventData": "{\"name\":\"John\"}"
	//}]}

	_, server := getTestServerWithRoutes(t)
	defer server.Close()

	client := getRestClient(t, server.URL, NAMESPACE)
	if client == nil {
		t.Fatal("REST Client is not initialized")
	}

	updateResponse, err := client.Update(eventuate.EntityIdAndType{
		ENTITY_TYPE,
		eventuate.Int128FromString(ENTITY_ID)},
		eventuate.Int128FromString(EVENT_ID_1),
		[]eventuate.EventTypeAndData{
			{
				EventType: EVENT_CHANGED,
				EventData: EVENT_DATA_2}},
		nil)
	assertNoError(t, err)

	tmp := eventuate.EntityIdVersionAndEventIds(eventuate.UpdateResponse{
		EntityId:      eventuate.Int128FromString(ENTITY_ID),
		EntityVersion: eventuate.Int128FromString(EVENT_ID_2),
		EventIds:      []eventuate.Int128{eventuate.Int128FromString(EVENT_ID_2)}})
	expected := &tmp

	assert.Equal(t, expected, updateResponse)

	hasUpdated = true

	events, err := client.Find(ENTITY_TYPE, eventuate.Int128FromString(ENTITY_ID), nil)
	assertNoError(t, err)

	getEventsAfterUpdateExpected := &eventuate.LoadedEvents{
		Events: []eventuate.EventIdTypeAndData{
			{
				EventId: eventuate.Int128FromString(EVENT_ID_1),
				EventTypeAndData: eventuate.EventTypeAndData{
					EventType: EVENT_CREATED,
					EventData: EVENT_DATA_1}},
			{
				EventId: eventuate.Int128FromString(EVENT_ID_2),
				EventTypeAndData: eventuate.EventTypeAndData{
					EventType: EVENT_CHANGED,
					EventData: EVENT_DATA_2,
				}}}}

	assert.Equal(t, getEventsAfterUpdateExpected, events)

}

func getRestClient(t *testing.T, url, space string) *eventuate.RESTClient {

	client, clientErr := eventuate.ClientBuilder().
		WithUrl(url).
		WithSpace(space).
		WithCredentials("123", "456").
		BuildREST()

	if clientErr != nil {
		t.Error(clientErr)
		return nil
	}

	client.SetInsecureSkipVerify(true)

	return client
}

func testServer() (*http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()

	server := httptest.NewUnstartedServer(mux)

	server.StartTLS()

	return mux, server
}

func getTestServerWithRoutes(t *testing.T) (*http.ServeMux, *httptest.Server) {
	mux, server := testServer()

	// https://api.eventuate.io/entity/ns12345678/
	mux.HandleFunc(fmt.Sprintf("/entity/%s", NAMESPACE),
		func(writer http.ResponseWriter, request *http.Request) {
			assertMethod(t, "POST", request)

			writer.Header().Set("Content-Type", "application/json")

			// {"entityId":"0000015a8a45d0bf-0242ac1100db0000",
			// "entityVersion":"0000015a8a45d0c4-0242ac1101020002",
			// "eventIds":["0000015a8a45d0c4-0242ac1101020002"]}

			fmt.Fprintf(writer,
				`{"entityId":"%s","entityVersion":"%s","eventIds":["%s"]}`,
				ENTITY_ID,
				eventuate.Int128FromString(EVENT_ID_1),
				EVENT_ID_1)
		})

	mux.HandleFunc(fmt.Sprintf("/entity/%s/%s/%s", NAMESPACE, ENTITY_TYPE, ENTITY_ID),
		func(w http.ResponseWriter, request *http.Request) {

			switch request.Method {
			case "GET":
				{
					w.Header().Set("Content-Type", "application/json")

					if !hasUpdated {
						fmt.Fprintf(w, `{"events":[{"id":"%s","eventType":"%s","eventData":"%s"}],"triggeringEvents":[]}`,
							eventuate.Int128FromString(EVENT_ID_1),
							EVENT_CREATED,
							EVENT_DATA_1_ESC)
					} else {
						// {"events":[{"id":"0000015a822f7198-0242ac1101020002","eventType":"net.chrisrichardson.eventstore.example.MyEntityWasCreated","eventData":"{\"name\":\"Fred\"}"},{"id":"0000015a8a9019ce-0242ac1101020002","eventType":"net.chrisrichardson.eventstore.example.MyEntityWasChanged","eventData":"{\"name\":\"Johnny\"}"}],"triggeringEvents":[]}
						fmt.Fprintf(w, `{"events":[{"id":"%s","eventType":"%s",
				"eventData":"%s"},{"id":"%s","eventType":"%s","eventData":"%s"}],"triggeringEvents":[]}`,
							eventuate.Int128FromString(EVENT_ID_1),
							EVENT_CREATED,
							EVENT_DATA_1_ESC,
							eventuate.Int128FromString(EVENT_ID_2),
							EVENT_CHANGED,
							EVENT_DATA_2_ESC)
					}
				}
			case "POST":
				{
					w.Header().Set("Content-Type", "application/json")

					// {"entityId":"0000015a822f7197-0242ac1100db0000","entityVersion":"0000015a8a9019ce-0242ac1101020002","eventIds":["0000015a8a9019ce-0242ac1101020002"]}
					fmt.Fprintf(w, `{"entityId":"%s","entityVersion":"%s","eventIds":["%s"]}`,
						ENTITY_ID,
						eventuate.Int128FromString(EVENT_ID_2),
						EVENT_ID_2)
				}
			default:
				{
					t.Error(request.URL.String())
				}
			}

		})

	return mux, server
}
