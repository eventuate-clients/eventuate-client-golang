package eventuate

import (
	//"log"
	"fmt"
	"reflect"
)

func callMethod(i interface{}, methodName string, args ...interface{}) ([]reflect.Value, error) {
	var ptr reflect.Value
	var value reflect.Value
	var finalMethod reflect.Value

	value = reflect.ValueOf(i)

	// if we start with a pointer, we need to get value pointed to
	// if we start with a value, we need to get a pointer to that value
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(value)
	}

	// check for method on value
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}

	if finalMethod.IsValid() && finalMethod.Type().NumIn() == len(args) {

		methodType := finalMethod.Type()

		params := make([]reflect.Value, len(args))
		for idx, param := range args {
			var ptr1 reflect.Value
			val := reflect.ValueOf(param)

			if val.Type().Kind() == reflect.Interface {
				params[idx] = val.Elem()
				continue
			}

			if val.Type().Kind() == reflect.Ptr {
				ptr1 = val
				val = ptr1.Elem()
			} else {
				ptr1 = reflect.New(reflect.TypeOf(param))
				temp := ptr1.Elem()
				temp.Set(val)
			}
			expectedArg := methodType.In(idx)

			//log.Printf("before calling coerceTo(%#v, %v)", param, expectedArg)
			tmpVal, tmpErr := coerceTo(ptr1, expectedArg)
			if tmpErr != nil {
				return nil, tmpErr
			}
			params[idx] = tmpVal
		}

		return finalMethod.Call(params), nil
	}

	//return or panic, method not found of either type
	return nil, MethodNotFoundError("Method: %v, Receiver: %v", methodName, i)
}

func coerceTo(of reflect.Value, i reflect.Type) (reflect.Value, error) {
	//log.Printf("inside coerceTo(%#v %T, %v)", of.Interface(), of.Interface(), i)
	if of.Type() == i {
		return of, nil
	}
	if i.Kind() == reflect.Interface {
		return of, nil
	}
	if i.Kind() == reflect.Ptr && of.Kind() == reflect.Struct {
		return reflect.Value{},
			SignatureMismatchExpRefGotValError("Actual: %T, Aggregate Expects: %v",
				of.Interface(), i)
	}
	if of.Kind() == reflect.Ptr || of.Kind() == reflect.Interface {

		tmp := of.Elem()
		if tmp == of {
			return of, nil
		}
		return coerceTo(tmp, i)
	}

	return of, nil

}

func checkUnderlyingType(i interface{}, checkTo reflect.Type) bool {
	value := reflect.ValueOf(i)
	return getUnderlyingType(value.Type()) == getUnderlyingType(checkTo)
}

func getUnderlyingType(inputType reflect.Type) reflect.Type {
	if inputType.Kind() == reflect.Ptr {
		return inputType.Elem()
	}
	return inputType
}

func getUnderlyingValue(input interface{}) interface{} {
	value := reflect.ValueOf(input)
	inputType := value.Type()
	if inputType.Kind() == reflect.Ptr {
		tmp := value.Elem().Interface()
		if tmp != input {
			return getUnderlyingValue(tmp)
		}
	}
	return input
}

func hasProp(input reflect.Type, name, tag string) bool {
	if input.Kind() != reflect.Struct {
		return false
	}

	field, fieldOk := input.FieldByName(name)
	if fieldOk && len(tag) > 0 {
		_, tagOk := field.Tag.Lookup("json")
		fieldOk = tagOk
	}

	return fieldOk
}

func readString(input interface{}, propName string) string {
	//log.Print(reflect.TypeOf(input).Kind(), reflect.Ptr)
	switch reflect.TypeOf(input).Kind() {
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		{
			return readString(reflect.ValueOf(input).Elem().Interface(), propName)
		}
	}
	return reflect.ValueOf(input).FieldByName(propName).String()
}

func readInterface(input interface{}, propName string) interface{} {
	if reflect.TypeOf(input).Kind() == reflect.Ptr {
		if len(propName) == 0 {
			return readInterface(reflect.ValueOf(input).Elem().Interface(), propName)
		}
		return reflect.ValueOf(input).Elem().FieldByName(propName).Interface()
	}
	if len(propName) == 0 {
		return input
	}
	return reflect.ValueOf(input).FieldByName(propName).Interface()
}

func setProp(ctx interface{}, propName string, value interface{}) {
	contextObj := reflect.ValueOf(ctx)
	if contextObj.Type().Kind() == reflect.Ptr {
		contextObj = contextObj.Elem()
	}
	fldToAssignTo := contextObj.FieldByName(propName)

	newValue := reflect.ValueOf(value)
	if newValue.Type().Kind() == reflect.Ptr {
		newValue = newValue.Elem()
	}

	if fldToAssignTo.IsValid() {
		if fldToAssignTo.CanSet() {
			fldToAssignTo.Set(newValue)
		}
	}
	//if fldToAssignTo.
	//.Set(reflect.ValueOf(value))
}

func getVerifiedArgumentType(method reflect.Method, eventName string) (reflect.Type, bool) {
	if method.Type.NumIn() != 2 { // receiver (0) + Event (1)
		return nil, false
	}

	inEventStructType := method.Type.In(1) // Event struct
	underlyingType := getUnderlyingType(inEventStructType)

	if len(eventName) != 0 && underlyingType.Name() != fmt.Sprintf("%vEvent", eventName) {
		return nil, false
	}

	return underlyingType, true
}
