package eventuate

import (
	"encoding/json"
	"fmt"
	loglib "github.com/shopcookeat/eventuate-client-golang/logger"
	"reflect"
	"regexp"
)

type AggregateRepository struct {
	Client    Crud
	ll        loglib.LogLevelEnum
	lg        loglib.Logger
	typeHints typeHintsMap
	meta      *AggregateMetadata
}

func (repo *AggregateRepository) RegisterEventType(name string, typeInstance interface{}) error {
	return repo.typeHints.RegisterEventType(name, typeInstance)
}

func NewAggregateRepository(client Crud, meta *AggregateMetadata) *AggregateRepository {
	var (
		ll        loglib.LogLevelEnum = loglib.Silent
		lg        loglib.Logger       = loglib.NewNilLogger()
		typeHints typeHintsMap        = NewTypeHintsMap()
	)
	maybeRestClient, maybeRestClientOk := client.(*RESTClient)
	if maybeRestClientOk {
		ll = maybeRestClient.ll
		lg = maybeRestClient.lg
		typeHints = maybeRestClient.typeHints
	}
	return &AggregateRepository{
		Client:    client,
		ll:        ll,
		lg:        lg,
		typeHints: typeHints,
		meta:      meta}
}

func (repo *AggregateRepository) SetLogLevel(level loglib.LogLevelEnum) *AggregateRepository {
	repo.ll = level
	repo.lg = loglib.NewLogger(level)
	return repo
}

func (repo *AggregateRepository) Save(cmd Command) (*EntityMetadata, error) {

	var meta *AggregateMetadata = repo.meta

	meta.ll = repo.ll
	meta.lg = repo.lg

	entity, ctorErr := meta.newInstance()
	if ctorErr != nil {
		return nil, ctorErr
	}

	events, processErr := entity.ProcessCommand(cmd)
	if processErr != nil {
		return nil, processErr
	}

	if len(events) == 0 {
		return &EntityMetadata{
			EntityTypeName: meta.EntityTypeName,
			EntityId:       Int128Nil,
			EntityVersion:  Int128Nil,
			HasEntity:      false,
			EntityInstance: nil,
			metadata:       meta}, nil
	}

	mappedEvents, errMapping := prepareEventsForEventuate(events, repo.typeHints)
	if errMapping != nil {
		return nil, errMapping
	}

	options := &AggregateCrudSaveOptions{}

	evEntity, saveErr := repo.Client.Save(meta.EntityTypeName, mappedEvents, options)
	if saveErr != nil {
		return nil, AppError("Repository persist exception (Save): %v", saveErr)
	}

	return &EntityMetadata{
		EntityTypeName: meta.EntityTypeName,
		EntityId:       evEntity.EntityId,
		EntityVersion:  evEntity.EntityVersion,
		HasEntity:      false,
		EntityInstance: nil,
		metadata:       meta}, nil
}

func (repo *AggregateRepository) Update(entityId Int128, cmd Command) (*EntityMetadata, error) {

	var meta *AggregateMetadata = repo.meta

	meta.ll = repo.ll
	meta.lg = repo.lg

	entity, findErr := repo.Find(entityId)
	if findErr != nil {
		return nil, findErr
	}

	repo.lg.Printf("Entity, after Find(), before Update(): %#v\n", entity)

	events, processErr := entity.ProcessCommand(cmd)
	if processErr != nil {
		return nil, processErr
	}

	if len(events) == 0 {
		return &EntityMetadata{
			EntityTypeName: meta.EntityTypeName,
			EntityId:       entity.EntityId,
			EntityVersion:  entity.EntityVersion,
			HasEntity:      false,
			EntityInstance: nil,
			metadata:       meta}, nil
	}

	mappedEvents, errMapping := prepareEventsForEventuate(events, repo.typeHints)
	if errMapping != nil {
		return nil, errMapping
	}

	options := &AggregateCrudUpdateOptions{}

	evEntity, updErr := repo.Client.Update(EntityIdAndType{
		EntityType: entity.EntityTypeName,
		EntityId:   entityId}, entity.EntityVersion, mappedEvents, options)

	if updErr != nil {
		return nil, AppError("Repository persist exception (Update): %v", updErr)
	}

	return &EntityMetadata{
		EntityTypeName: meta.EntityTypeName,
		EntityId:       evEntity.EntityId,
		EntityVersion:  evEntity.EntityVersion,
		HasEntity:      false,
		EntityInstance: nil,
		metadata:       meta}, nil
}

func (repo *AggregateRepository) Find(entityId Int128) (*EntityMetadata, error) {

	var meta *AggregateMetadata = repo.meta

	meta.lmu.Lock()
	meta.ll = repo.ll
	meta.lg = repo.lg
	meta.lmu.Unlock()

	entity, ctorErr := meta.newInstance()
	if ctorErr != nil {
		return nil, ctorErr
	}

	options := &AggregateCrudFindOptions{}
	loadedEvents, findErr := repo.Client.Find(
		meta.EntityTypeName,
		entityId,
		options)

	if findErr != nil {
		return nil, AppError("Repository search exception (Find): %v", findErr)
	}

	eventuateEvents := loadedEvents.Events
	//snapshot := loadedEvents.Snapshot // ???

	events, entityVersion, deserializationErr := materializeEventsFromEventuate(meta, eventuateEvents, repo.typeHints)
	if deserializationErr != nil {
		return nil, deserializationErr
	}

	nextEntity, applyErr := entity.applyEvents(events)
	if applyErr != nil {
		return nil, applyErr
	}

	return &EntityMetadata{
		EntityTypeName: meta.EntityTypeName,
		EntityId:       entityId,
		EntityVersion:  entityVersion,
		HasEntity:      true,
		EntityInstance: nextEntity.EntityInstance,
		metadata:       meta}, nil
}

func prepareEventsForEventuate(events []Event, eventTypesMap TypeHintMapper) ([]EventTypeAndData, error) {
	mappedEvents := make([]EventTypeAndData, len(events))
	for idx, event := range events {

		val := getUnderlyingValue(event)
		serializedEvent, err := json.Marshal(val)
		if err != nil {
			return nil,
				AppError("Cannot serialize Event (%v), json Error: %v", event, err)
		}

		evtTypeName := getUnderlyingType(reflect.TypeOf(event)).Name()
		hasEventName, eventName := eventTypesMap.GetTypeByTypeName(evtTypeName)

		if !hasEventName {
			panic(fmt.Sprintf("Cannot serialize EventType (%v), register its type first", event))
		}

		mappedEvents[idx] = EventTypeAndData{
			EventType: eventName,
			EventData: string(serializedEvent)}
	}
	return mappedEvents, nil
}

func materializeEventsFromEventuate(meta *AggregateMetadata, events []EventIdTypeAndData, typeHints TypeHintMapper) ([]interface{}, Int128, error) {
	eventNamePattern := regexp.MustCompile(`^(\w+)Event$`)
	var entityVersion Int128

	for i, ev := range events {
		meta.lg.Printf("%d: %v", i, ev)
	}

	mappedEvents := make([]interface{}, len(events))

	for idx, event := range events {

		entityVersion = event.EventId

		isRegistered := typeHints.HasEventType(event.EventType)
		registeredType := typeHints.GetEventType(event.EventType)
		if isRegistered {
			meta.lg.Printf("event.EventType = %v, registeredType = %v, Is Interface? %v; Is Ptr? %v; Is Struct? %v\n",
				event.EventType, registeredType, registeredType.Kind() == reflect.Interface, registeredType.Kind() == reflect.Ptr, registeredType.Kind() == reflect.Struct)
			newVal := reflect.New(registeredType).Interface()
			err := json.Unmarshal([]byte(event.EventData), newVal)
			//_, err := reJson(event.EventData, newVal)
			if err != nil {
				return nil, Int128Nil, AppError("Cannot deserialize Event of type `%s` into a registered type (%v) for data: %v",
					event.EventType,
					registeredType,
					event)
			}
			//mappedEvents[idx] = Event(result)
			mappedEvents[idx] = getUnderlyingValue(newVal)
			continue
		}

		eventMostSpecificPart := lastPart(event.EventType, "")

		eventCoreName := ""
		if eventNamePattern.MatchString(eventMostSpecificPart) {
			eventCoreName = eventNamePattern.ReplaceAllString(eventMostSpecificPart, `$1`)
		}

		typeExists := typeHints.HasEventType(eventCoreName)
		eventType := typeHints.GetEventType(eventCoreName)

		if typeExists {
			newVal := reflect.New(eventType).Interface()
			_, err := reJson(event.EventData, newVal)
			if err != nil {
				return nil, Int128Nil, AppError("Cannot deserialize Event of type `%s` into a reflected type (%v) for data: %v",
					event.EventType,
					eventType,
					event)
			}
			mappedEvents[idx] = getUnderlyingValue(newVal)
		} else {
			var container interface{}
			jsonData := []byte(event.EventData)
			err := json.Unmarshal(jsonData, container)
			if err != nil {
				return nil, Int128Nil, AppError("Cannot deserialize Event of type `%s` into a general container for data: %v",
					event.EventType,
					event)

			}
			mappedEvents[idx] = container
		}

	}
	return mappedEvents, entityVersion, nil
}

func reJson(evtData string, propResult interface{}) (interface{}, error) {
	unmarshalErr := json.Unmarshal([]byte(evtData), propResult)
	return propResult, unmarshalErr
}
