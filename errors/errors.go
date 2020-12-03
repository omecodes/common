package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var New = errors.New
var Errorf = fmt.Errorf

type Error uint32

const (
	Internal            = Error(1)
	NotFound            = Error(2)
	Unavailable         = Error(3)
	Forbidden           = Error(4)
	Unauthorized        = Error(5)
	Duplicate           = Error(6)
	BadInput            = Error(7)
	NotSupported        = Error(8)
	NotImplemented      = Error(9)
	ServiceNotAvailable = Error(10)
)

func (e Error) Error() string {
	switch e {
	case NotFound:
		return "not found"

	case Duplicate:
		return "duplicate"

	case Forbidden:
		return "forbidden"

	case Unauthorized:
		return "unauthorized"

	case Unavailable:
		return "unavailable"

	case BadInput:
		return "bad input"

	case NotSupported:
		return "not supported"

	case NotImplemented:
		return "not implemented"

	case ServiceNotAvailable:
		return "service not available"

	default:
		return "internal"
	}
}

func (e Error) HttpCode() int {
	return HttpStatus(e)
}

func HttpStatus(e error) int {
	str := e.Error()
	switch str {
	case "not found":
		return http.StatusNotFound
	case "duplicate":
		return http.StatusConflict
	case "forbidden":
		return http.StatusForbidden
	case "unauthorized":
		return http.StatusUnauthorized
	case "unavailable":
		return http.StatusServiceUnavailable
	case "bad input":
		return http.StatusBadRequest
	case "not supported":
		return http.StatusHTTPVersionNotSupported
	case "not implemented":
		return http.StatusNotImplemented
	case "service not available":
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func IsNotFound(e error) bool {
	return errors.Is(e, NotFound) || e.Error() == "not found"
}

func IsForbidden(e error) bool {
	return errors.Is(e, Forbidden) || e.Error() == "forbidden"
}
