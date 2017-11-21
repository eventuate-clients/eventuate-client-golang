package eventuate

import "reflect"

type EntityMetadata struct {
	EntityTypeName string             `json:"typeName"`
	EntityId       Int128             `json:"entityId"`
	EntityVersion  Int128             `json:"version"`
	HasEntity      bool               `json:"-"`
	EntityInstance interface{}        `json:"-"`
	metadata       *AggregateMetadata `json:"-"`
}

func (entity *EntityMetadata) ApplyEvent(event Event) (*EntityMetadata, error) {
	//eventNamePattern := regexp.MustCompile(`^(\w+)Event$`)

	if !entity.HasEntity {
		return nil, AppError("ApplyEvent: cannot apply events to un-synced entity")
	}

	aggregate := entity.EntityInstance
	meta := entity.metadata

	eventData := getUnderlyingValue(event)
	eventType := reflect.TypeOf(eventData)

	eventMethodName := "ApplyEvent"
	if eventMethod, specificMethodExists := meta.eventMethodsMap[eventType]; specificMethodExists {
		eventMethodName = eventMethod.Name
	}

	values, callErr := callMethod(aggregate, eventMethodName, event)
	if callErr != nil {
		return entity, callErr
	}
	// len(values) == 1 is ensured in the CreateAggregateMetadata(..)

	nextInstanceInterface := values[0].Interface()
	if !checkUnderlyingType(nextInstanceInterface, meta.UnderlyingType) {
		return entity, AppError("Signature mismatch. Method: %s is either missing or returned unexpected type: %T, not %v", eventMethodName, nextInstanceInterface, meta.UnderlyingType)
	}

	return &EntityMetadata{
		EntityTypeName: entity.EntityTypeName,
		HasEntity:      true,
		EntityInstance: nextInstanceInterface,
		metadata:       entity.metadata}, nil

}

func (entity *EntityMetadata) applyEvents(events []interface{}) (*EntityMetadata, error) {
	result := entity
	for _, event := range events {
		nextEntity, err := entity.ApplyEvent(event)
		if err != nil {
			return nil, AppError("Event Application Error: %v", err)
		}
		result = nextEntity
	}
	return result, nil
}

func (entity *EntityMetadata) ProcessCommand(command Command) ([]Event, error) {

	if !entity.HasEntity {
		return nil, AppError("ProcessCommand: cannot process commands against an un-synced entity")
	}

	commandType := reflect.TypeOf(command)
	underlyingType := getUnderlyingType(commandType)

	commandMethodName := "ProcessCommand"
	commandMethod, doesMethodExist := entity.metadata.commandMethodsMap[underlyingType.Name()]
	if doesMethodExist {
		commandMethodName = commandMethod.Name
	} else {
		_, doesMethodExist = entity.metadata.commandMethodsMap[""]
	}

	if !doesMethodExist {
		return nil, AppError("ProcessCommand: command argument cannot be processed, type: %v", underlyingType)
	}

	aggregate := entity.EntityInstance

	values, callErr := callMethod(aggregate, commandMethodName, command)
	if callErr != nil {
		return nil, callErr
	}

	result1 := values[0].Interface()

	switch t := result1.(type) {
	case *[]Event:
		{
			return *t, nil
		}
	case []Event:
		{
			return t, nil
		}
	case []interface{}:
		{
			result := make([]Event, len(t))
			for idx, evt := range t {
				if castEvt, castOk := evt.(Event); castOk {
					result[idx] = castEvt
				}
			}
			return result, nil
		}
	default:
		{
			return nil,
				AppError("ProcessCommand: unexpected type %T in method: %s",
					t, commandMethodName)

		}
	}
}
