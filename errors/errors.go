package errors2

import (
	"fmt"
	"net/http"
	"strings"
)

var Forbidden = &Error{Code: http.StatusForbidden, Message: "forbidden"}
var Unauthorized = &Error{Code: http.StatusUnauthorized, Message: "not allowed"}
var NotFound = &Error{Code: http.StatusNotFound, Message: "not found"}
var Internal = &Error{Code: http.StatusInternalServerError, Message: "internal service error"}
var BadRequest = &Error{Code: http.StatusBadRequest, Message: "bad request"}
var NotImplemented = &Error{Code: http.StatusNotImplemented, Message: "not implemented"}
var Processing = &Error{Code: http.StatusProcessing, Message: "processing"}
var TimeOut = &Error{Code: http.StatusRequestTimeout, Message: "timeout"}

var Duplicate = &Error{Code: 100, Message: "duplicate key"}
var Unexpected = &Error{Code: 101, Message: "unexpected"}
var WrongContent = &Error{Code: 102, Message: "wrong content"}

type Error struct {
	Code    int
	Message string
	Details string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%s", e.Message, e.Details)
}

func New(e *Error, details string) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

func GetErrorCode(e error) int {
	eObject, ok := e.(*Error)
	if ok {
		return eObject.Code
	}

	msg := e.Error()

	switch msg {
	case Forbidden.Message:
		return Forbidden.Code
	case Unauthorized.Message:
		return Unauthorized.Code
	case NotFound.Message:
		return NotFound.Code
	case BadRequest.Message:
		return BadRequest.Code
	case NotImplemented.Message:
		return NotImplemented.Code
	case Processing.Message:
		return Processing.Code
	case TimeOut.Message:
		return TimeOut.Code
	default:
		return Internal.Code
	}
}

func Parse(str string) *Error {
	parts := strings.Split(str, ":")
	head := parts[0]
	details := ""
	if len(parts) > 1 {
		details = parts[1]
	}

	switch head {
	case Forbidden.Message:
		return New(Forbidden, details)
	case Unauthorized.Message:
		return New(Unauthorized, details)
	case NotFound.Message:
		return New(NotFound, details)
	case BadRequest.Message:
		return New(BadRequest, details)
	case NotImplemented.Message:
		return New(NotImplemented, details)
	case Processing.Message:
		return New(Processing, details)
	case TimeOut.Message:
		return New(TimeOut, details)
	default:
		return New(Internal, details)
	}
}

func Is(err error, e *Error) bool {
	er := Parse(err.Error())
	return er.Code == e.Code
}
