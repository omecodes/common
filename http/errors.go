package http

import "encoding/json"

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	data, _ := json.Marshal(e)
	return string(data)
}

var (
	NotFoundError     = &Error{Code: 404, Message: "not found"}
	BadRequestError   = &Error{Code: 400, Message: "bad request"}
	InternalError     = &Error{Code: 500, Message: "internal error"}
	ForbiddenError    = &Error{Code: 401, Message: "forbidden"}
	UnAuthorizedError = &Error{Code: 403, Message: "unauthorize"}
)

func IsNotFound(e *Error) bool {
	return e == NotFoundError
}

func Parse(e error) *Error {
	if ce, ok := e.(*Error); ok {
		return ce
	}

	return &Error{Code: 500, Message: "internal"}
}
