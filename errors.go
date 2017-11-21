package eventuate

import "fmt"
import (
	"net/http"
)

type appErrCode int

const (
	appErrDefault appErrCode = iota
	appErrSignaturesMismatch
	appErrSignaturesMismatchExpRefGotVal
	appErrSignaturesMismatchExpValGotRef
	appErrMethodNotFound
	appErrRestErrors
)

type appError struct {
	code         appErrCode
	shortMessage string
	args         []interface{}
}

type appRestError struct {
	appError
	httpCode int
	conflict string
}

func (e *appError) Error() string {
	//panic(e)
	prefix := "Eventuate General Error"
	switch e.code {
	case appErrSignaturesMismatch:
		{
			prefix = "Signatures Mismatch Error"
		}
	case appErrSignaturesMismatchExpRefGotVal:
		{
			prefix = "Signatures Mismatch Error (expected by-ref, got by-val)"
		}
	case appErrSignaturesMismatchExpValGotRef:
		{
			prefix = "Signatures Mismatch Error (expected by-val, got by-ref)"
		}
	case appErrMethodNotFound:
		{
			prefix = "Method Not Found"
		}
	}
	if len(e.args) == 0 {
		return fmt.Sprintf("%v: %v", prefix, e.shortMessage)
	}
	return fmt.Sprintf("%v: %v", prefix,
		fmt.Sprintf(e.shortMessage, e.args...))
}

func appErrorWithCode(code appErrCode, shortMessage string, args ...interface{}) *appError {
	return &appError{
		code,
		shortMessage, args}
}
func SignatureMismatchError(shortMessage string, args ...interface{}) *appError {
	return appErrorWithCode(appErrSignaturesMismatch, shortMessage, args...)
}
func SignatureMismatchExpRefGotValError(shortMessage string, args ...interface{}) *appError {
	return appErrorWithCode(appErrSignaturesMismatchExpRefGotVal, shortMessage, args...)
}
func SignatureMismatchExpValGotRefError(shortMessage string, args ...interface{}) *appError {
	return appErrorWithCode(appErrSignaturesMismatchExpValGotRef, shortMessage, args...)
}
func MethodNotFoundError(shortMessage string, args ...interface{}) *appError {
	return appErrorWithCode(appErrMethodNotFound, shortMessage, args...)
}

func AppError(shortMessage string, args ...interface{}) *appError {
	return &appError{
		appErrDefault,
		shortMessage, args}
}

func RestError(httpCode int, conflict string, shortMessage string, args ...interface{}) *appRestError {
	return &appRestError{
		appError{
			appErrRestErrors,
			shortMessage, args},
		httpCode,
		conflict}
}

func (e *appRestError) Error() string {
	var msg string

	prefix := "REST API Error: "
	commonErrorBody := fmt.Sprintf(e.shortMessage, e.args...)
	switch e.httpCode {
	case http.StatusInternalServerError:
		msg = fmt.Sprintf("Eventuate Server exception: \n%v",
			commonErrorBody)

	case http.StatusUnauthorized:
		msg = "Not authorized"

	case http.StatusNotFound:
		msg = "Resource is not found"

	case http.StatusServiceUnavailable:
		msg = "Server overloaded try again"

	case http.StatusConflict:
		{
			switch e.conflict {
			case "":
				msg = "Status Conflict general Error: unknown response"
			case "entity_exists":
				msg = "Status Conflict Error: entity already exists"
			case "optimistic_lock_error":
				// todo: add here entityIdAndType, entityVersion
				msg = "Status Conflict Error: optimistic locking"
			case "duplicate_event":
				msg = "Status Conflict Error: duplicate triggering Event"
			case "entity_temporarily_unavailable":
				msg = "Status Conflict Error: entity temporarily unavailable"
			default:
				msg = "Status Conflict general Error: unknown response"
			}
			//err = AppError("An optimistic locking failure or Event has already been processed")
		}
	default:
		msg = fmt.Sprintf("Unrecognized response status: %d", e.httpCode)
	}
	return fmt.Sprintf("%v %v\n%v", prefix, msg, commonErrorBody)
}
