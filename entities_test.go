package eventuate_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
	"github.com/shopcookeat/eventuate-client-golang"
)

const APPLICATION_NAMESPACE = "net.chrisrichardson.eventstore.example.todo"

func TestRESTClient_JsonSchemas(t *testing.T) {

	jsonSchemas := map[string]string{
		"Find Response": "json_schemas/rest/get-response.json",
		//"Create Request":  "json_schemas/rest/create-request.json",
		"Create Response": "json_schemas/rest/create-response.json",
		//"Update Request":  "json_schemas/rest/update-request.json",
		"Update Response": "json_schemas/rest/update-response.json"}

	factualStructures := map[string]interface{}{
		"Find Response": eventuate.LoadedEvents{
			Events: []eventuate.EventIdTypeAndData{
				{
					EventId: eventuate.Int128FromString(EVENT_ID_1),
					EventTypeAndData: eventuate.EventTypeAndData{
						EventType: EVENT_CREATED,
						EventData: EVENT_DATA_1}}}},
		"Create Response": eventuate.EntityIdVersionAndEventIds(eventuate.CreateResponse{
			EntityId:      eventuate.Int128FromString(ENTITY_ID),
			EntityVersion: eventuate.Int128FromString(EVENT_ID_1),
			EventIds:      []eventuate.Int128{eventuate.Int128FromString(EVENT_ID_1)}}),
		"Update Response": eventuate.UpdateResponse{
			EntityId:      eventuate.Int128FromString(ENTITY_ID),
			EntityVersion: eventuate.Int128FromString(EVENT_ID_2),
			EventIds:      []eventuate.Int128{eventuate.Int128FromString(EVENT_ID_2)}}}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for name, sPath := range jsonSchemas {
		expectedSchemaPath := fmt.Sprintf("file://%s/%s", pwd, sPath)
		expectedSchema := gojsonschema.NewReferenceLoader(expectedSchemaPath)
		factualStructure, ok := factualStructures[name]
		if !ok {
			t.Errorf("Structure is missing for: %s", name)
			continue
		}

		factualSchema := gojsonschema.NewGoLoader(factualStructure)

		validationResult, err := gojsonschema.Validate(expectedSchema, factualSchema)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, true, validationResult.Valid())

		if !validationResult.Valid() {
			t.Error(validationResult.Errors())
		}
	}
}

//
//func TestJsonSchemas(t *testing.T) {
//	t.Skip()
//	return
//	t.Parallel()
//
//	workDirectory, workDirectoryErr := os.Getwd()
//	logErrorAndQuit(workDirectoryErr)
//
//	loader := gojsonschema.NewGoLoader(
//		CreateEntityRequest{
//			Events: []EventTypeAndData{
//				{
//					EventType: "net.chrisrichardson.eventstore.example.MyEntity",
//					EventData: `"name":"Fred"`}},
//			EntityTypeName: "net.chrisrichardson.eventstore.example.MyEntity"})
//
//	schemaLoader := gojsonschema.NewReferenceLoader(
//		fmt.Sprintf("file://%s/json_schema_test_suite/create-request.json", workDirectory))
//
//	checkCreateRequest, createRequestError := gojsonschema.Validate(schemaLoader, loader)
//	assertNoError(t, createRequestError)
//
//	assert.Equal(t, true, checkCreateRequest.Valid())
//
//	loader = gojsonschema.NewGoLoader(
//		CreateResponse{
//			EntityId:      "000001588bce5f47-0242ac1100dc0000",
//			EntityVersion: "000001588bce5f44-0242ac1101020002",
//			EventIds:      []string{"000001588bce5f44-0242ac1101020002"}})
//
//	schemaLoader = gojsonschema.NewReferenceLoader(
//		fmt.Sprintf("file://%s/json_schema_test_suite/create-response.json", workDirectory))
//
//	checkCreateResponse, createResponseError := gojsonschema.Validate(schemaLoader, loader)
//	assertNoError(t, createResponseError)
//
//	assert.Equal(t, true, checkCreateResponse.Valid())
//
//	loader = gojsonschema.NewGoLoader(
//		UpdateEntityRequest{
//			Events: []EventTypeAndData{
//				{
//					EventType: "net.chrisrichardson.eventstore.example.MyEntityWasUpdated",
//					EventData: `"{name": "John"}`}},
//			EntityVersion: "000001588bce5f44-0242ac1101020002"})
//
//	schemaLoader = gojsonschema.NewReferenceLoader(
//		fmt.Sprintf("file://%s/json_schema_test_suite/update-request.json", workDirectory))
//
//	checkUpdateRequest, updateRequestError := gojsonschema.Validate(schemaLoader, loader)
//	assertNoError(t, updateRequestError)
//
//	assert.Equal(t, true, checkUpdateRequest.Valid())
//
//	loader = gojsonschema.NewGoLoader(
//		&UpdateResponse{
//			EntityId:      "000001588bce5f47-0242ac1100dc0000",
//			EntityVersion: "000001588bce5f44-0242ac1101020002",
//			EventIds:      []string{"000001588bce5f44-0242ac1101020002"}})
//
//	schemaLoader = gojsonschema.NewReferenceLoader(
//		fmt.Sprintf("file://%s/json_schema_test_suite/update-response.json", workDirectory))
//
//	checkUpdateResponse, updateResponseError := gojsonschema.Validate(schemaLoader, loader)
//	assertNoError(t, updateResponseError)
//
//	assert.Equal(t, true, checkUpdateResponse.Valid())
//
//	loader = gojsonschema.NewGoLoader(
//		&GetResponse{
//			Events: []EventIdTypeAndData{
//				{
//					EventType: "net.chrisrichardson.eventstore.example.MyEntityWasCreated",
//					EventData: `{"name": "Fred"}`,
//					EntityId:  "000001588bce5f44-0242ac1101020002"}}})
//
//	schemaLoader = gojsonschema.NewReferenceLoader(
//		fmt.Sprintf("file://%s/json_schema_test_suite/get-response.json", workDirectory))
//
//	checkGetResponse, getResponseError := gojsonschema.Validate(schemaLoader, loader)
//	assertNoError(t, getResponseError)
//
//	assert.Equal(t, true, checkGetResponse.Valid())
//}
