package eventuate

import (
	"fmt"
	loglib "github.com/eventuate-clients/eventuate-client-golang/logger"
	"log"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

type AggregateMetadata struct {
	Type              reflect.Type
	UnderlyingType    reflect.Type
	EntityTypeName    string
	EventMethods      []string
	CommandMethods    []string
	ll                loglib.LogLevelEnum
	lg                loglib.Logger
	lmu               sync.Mutex
	commandMethodsMap map[string]reflect.Method
	eventMethodsMap   map[reflect.Type]reflect.Method
	//eventTypesMap     TypeHintMapper
	newInstance func() (*EntityMetadata, error)
}

func (meta *AggregateMetadata) String() string {
	return fmt.Sprintf("EntityTypeName: %s, Type: %s, Event methods: [%v], Command methods: [%v]",
		meta.EntityTypeName,
		meta.Type.Name(),
		meta.EventMethods,
		meta.CommandMethods)
}

func CreateAggregateMetadata(newInstance interface{}, entityTypeName string) (*AggregateMetadata, error) {
	eventNamePattern := regexp.MustCompile(`^Apply(\w+)?Event$`)
	commandNamePattern := regexp.MustCompile(`^Process(\w+)?Command`)

	defer func() {
		if r := recover(); r != nil {
			err := AppError("Recovered (%v)", r)
			log.Fatal(err)
		}
	}()

	funcValue := reflect.ValueOf(newInstance)
	funcType := funcValue.Type()
	if funcType.Kind() != reflect.Func || funcType.NumIn() != 0 || funcType.NumOut() != 1 {
		return nil, AppError("%s (%s)",
			"`newInstance` signature mismatch",
			"newInstance must be a function that takes no arguments and produces exactly one value")
	}

	funcTypeOut := funcType.Out(0)

	underlyingFuncTypeOut := funcTypeOut
	if funcTypeOut.Kind() == reflect.Ptr {
		underlyingFuncTypeOut = funcTypeOut.Elem()
	}

	methodsCount := funcTypeOut.NumMethod()
	methodsMap := make(map[string]reflect.Method)
	methodNames := make([]string, methodsCount)

	methods := make([]reflect.Method, methodsCount)
	commandMethods := make(map[string]reflect.Method)
	eventMethods := make(map[reflect.Type]reflect.Method)
	//eventTypes := make(map[string]reflect.Type)

	for i := 0; i < methodsCount; i++ {
		method := funcTypeOut.Method(i)

		methods = append(methods, method)

		methodName := method.Name
		methodNames = append(methodNames, methodName)

		methodsMap[methodName] = method

		var commandType reflect.Type
		if commandNamePattern.MatchString(methodName) {
			// dealing with Command

			commandCoreName := commandNamePattern.ReplaceAllString(methodName, `$1`)

			isInParamsOk := method.Type.NumIn() == 2
			isOutParamsOk := method.Type.NumOut() == 1
			isGenericName := len(commandCoreName) == 0
			isMethodOk := isInParamsOk && isOutParamsOk

			isValidCommand := false
			commandKey := ""

			if isMethodOk {
				if isGenericName {
					isValidCommand = true
				} else {
					commandType = method.Type.In(1)
					underlyingType := getUnderlyingType(commandType)
					isValidCommand = strings.HasSuffix(underlyingType.Name(), "Command")
					commandKey = underlyingType.Name()
				}
			}

			_, hasDuplicateKey := commandMethods[commandKey]

			if isMethodOk && isValidCommand && !hasDuplicateKey {
				commandMethods[commandKey] = method
			} else {
				if hasDuplicateKey {
					return nil, AppError("Same command in several methods. func (recv %s) %s(command %s).",
						funcTypeOut,
						methodName,
						commandType)
				}
				return nil, AppError("Signature mismatch. func (recv %s) %s(command %s). Ensure the argument `command` ends with 'Command', and the method returns a single value",
					funcTypeOut,
					methodName,
					commandType)
			}
		}

		if eventNamePattern.MatchString(methodName) {
			// dealing with Event
			eventCoreName := eventNamePattern.ReplaceAllString(methodName, `$1`)

			classDef, classOk := getVerifiedArgumentType(method, eventCoreName)
			//if classOk && len(eventCoreName) > 0 {
			//	//underlyingType := getUnderlyingType(classDef)
			//	//classOk = hasProp(underlyingType, "EventType", "eventType") &&
			//	//	hasProp(underlyingType, "EventData", "eventData")
			//}
			classOk = classOk && (method.Type.NumOut() == 1)
			if classOk {
				eventMethods[classDef] = method
				//eventTypes[eventCoreName] = classDef
			} else {
				return nil, AppError("Signature mismatch. func (recv %s) %s(Event %v). Ensure the argument `Event` is a structure whose name ends with 'Event', and the method returns a single value",
					funcTypeOut,
					methodName,
					classDef)
			}
		}

	}

	meta := &AggregateMetadata{
		Type:              funcTypeOut,
		UnderlyingType:    underlyingFuncTypeOut,
		EntityTypeName:    entityTypeName,
		EventMethods:      Filter(methodNames, eventNamePattern.MatchString),
		CommandMethods:    Filter(methodNames, commandNamePattern.MatchString),
		eventMethodsMap:   eventMethods,
		//eventTypesMap:     TypeHintMapper,
		commandMethodsMap: commandMethods,
		newInstance:       nil}

	meta.newInstance = createConstructor(funcValue, meta)

	return meta, nil
}

func createConstructor(funcValue reflect.Value, meta *AggregateMetadata) func() (*EntityMetadata, error) {
	return func() (*EntityMetadata, error) {
		results := funcValue.Call([]reflect.Value{})
		if len(results) == 2 {
			maybeError := results[1].Interface()
			if !reflect.DeepEqual(maybeError, reflect.Zero(results[1].Type()).Interface()) {
				return nil, AppError("Constructor for type %v returned with Error. (%v)",
					meta.Type,
					maybeError)
			}
		}
		tmp := results[0].Interface()
		return &EntityMetadata{
			EntityTypeName: meta.EntityTypeName,
			HasEntity:      true,
			EntityInstance: tmp,
			metadata:       meta}, nil
	}
}
