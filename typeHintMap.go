package eventuate

import (
	"reflect"
	"strings"
)

type typeHintsMap map[string]reflect.Type

func NewTypeHintsMap() typeHintsMap {
	return make(map[string]reflect.Type)
}

func (hints typeHintsMap) Merge(newHints typeHintsMap) {
	for k, v := range newHints {
		hints[k] = v
	}
}

func (hints typeHintsMap) MakeCopy() typeHintsMap {
	result := make(map[string]reflect.Type)
	for k, v := range hints {
		result[k] = v
	}
	return result
}

func (hints typeHintsMap) HasEventType(name string) bool {
	_, hasType := hints[name]
	return hasType
}

func (hints typeHintsMap) GetEventType(name string) reflect.Type {
	result, hasType := hints[name]
	if hasType {
		return result
	}
	return nil
}

func (hints typeHintsMap) GetTypeByKeyName(name string) (bool, reflect.Type) {
	result, hasIt := hints[name]
	return hasIt, result
}

func (hints typeHintsMap) GetTypeByTypeName(typeName string) (bool, string) {
	for mk, mv := range hints {
		if mv.Name() == typeName {
			return true, mk
		}
		//log.Printf("Type %s (%s): nope", mv.Name(), getUnderlyingType(mv).Name())
		str := getUnderlyingType(mv).Name()
		if lastPart(typeName, ".") == mv.Name() {
			return true, mk
		}
		if lastPart(typeName, ".") == str {
			return true, mk
		}
	}
	return false, ""
}

func (hints typeHintsMap) RegisterEventType(name string, typeInstance interface{}) error {
	argType := reflect.TypeOf(typeInstance)
	underlyingType := getUnderlyingType(argType)
	if underlyingType.Kind() != reflect.Struct {
		return AppError("Type hint `%s` requires a type (`%T`) which is not ultimately a structure",
			name,
			underlyingType)
	}

	classOk := strings.HasSuffix(underlyingType.Name(), "Event")

	if !classOk {
		return AppError("Type hint `%s` requires a structure (`%T`) named with ending 'Event'",
			name,
			underlyingType)
	}
	hints[name] = argType
	return nil
}
