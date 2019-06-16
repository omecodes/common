package errors

import (
	"fmt"
	"strings"
)

var Forbidden = &Error{Code: 403, Message: "forbidden"}
var Unauthorized = &Error{Code: 401, Message: "not allowed"}
var NotFound = &Error{Code: 404, Message: "not found"}
var Internal = &Error{Code: 500, Message: "internal service error"}
var BadRequest = &Error{Code: 400, Message: "bad request"}
var NotImplemented = &Error{Code: 501, Message: "not implemented"}
var Processing = &Error{Code: 502, Message: "processing"}
var TimeOut = &Error{Code: 408, Message: "timeout"}

var Duplicate = &Error{Code: 100, Message: "duplicate key"}
var Unexpected = &Error{Code: 101, Message: "unexpected"}
var WrongContent = &Error{Code: 102, Message: "wrong content"}
var BadInput = &Error{Code: 103, Message: "bad input"}

type Error struct {
	Code    int
	Message string
	Details string
}

func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s:%s", e.Message, e.Details)
	} else {
		return e.Message
	}
}

func Detailed(e *Error, details string) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

func New(message string) error {
	return &Error{
		Code:    Internal.Code,
		Message: Internal.Message,
		Details: message,
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
		return Detailed(Forbidden, details)
	case Unauthorized.Message:
		return Detailed(Unauthorized, details)
	case NotFound.Message:
		return Detailed(NotFound, details)
	case BadRequest.Message:
		return Detailed(BadRequest, details)
	case NotImplemented.Message:
		return Detailed(NotImplemented, details)
	case Processing.Message:
		return Detailed(Processing, details)
	case TimeOut.Message:
		return Detailed(TimeOut, details)
	default:
		return Detailed(Internal, details)
	}
}

func Is(err error, e *Error) bool {
	er := Parse(err.Error())
	return er.Code == e.Code
}
