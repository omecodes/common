package errors

import (
	"fmt"
	"strings"
)

var (
	HttpForbidden      = &Error{Code: 403, Message: "forbidden"}
	HttpUnauthorized   = &Error{Code: 401, Message: "not allowed"}
	HttpNotFound       = &Error{Code: 404, Message: "not found"}
	HttpInternal       = &Error{Code: 500, Message: "internal service error"}
	HttpBadRequest     = &Error{Code: 400, Message: "bad request"}
	HttpNotImplemented = &Error{Code: 501, Message: "not implemented"}
	HttpProcessing     = &Error{Code: 502, Message: "processing"}
	HttpTimeOut        = &Error{Code: 408, Message: "timeout"}
	NotFound           = &Error{Code: 1, Message: "not found"}
	Unauthorized       = &Error{Code: 2, Message: "unauthorized"}
	NotSupported       = &Error{Code: 3, Message: "not supported"}
	Duplicate          = &Error{Code: 4, Message: "duplicate key"}
	Unexpected         = &Error{Code: 5, Message: "unexpected"}
	WrongContent       = &Error{Code: 6, Message: "wrong content"}
	BadInput           = &Error{Code: 7, Message: "bad input"}
)

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
		Code:    HttpInternal.Code,
		Message: HttpInternal.Message,
		Details: message,
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
	case HttpForbidden.Message:
		return Detailed(HttpForbidden, details)
	case HttpUnauthorized.Message:
		return Detailed(HttpUnauthorized, details)
	case HttpNotFound.Message:
		return Detailed(HttpNotFound, details)
	case HttpBadRequest.Message:
		return Detailed(HttpBadRequest, details)
	case HttpNotImplemented.Message:
		return Detailed(HttpNotImplemented, details)
	case HttpProcessing.Message:
		return Detailed(HttpProcessing, details)
	case HttpTimeOut.Message:
		return Detailed(HttpTimeOut, details)
	default:
		return Detailed(HttpInternal, details)
	}
}

func IsHttpNotFound(e error) bool {
	return Parse(e.Error()).Code == HttpNotFound.Code
}

func IsHttpUnauthorized(e error) bool {
	return Parse(e.Error()).Code == Unauthorized.Code
}

func IsNotFound(e error) bool {
	return Parse(e.Error()).Code == NotFound.Code
}
