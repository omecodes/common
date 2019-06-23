package errors

var (
	HttpForbidden      = &Error{Code: 403, Message: "HTTP forbidden"}
	HttpUnauthorized   = &Error{Code: 401, Message: "HTTP not allowed"}
	HttpNotFound       = &Error{Code: 404, Message: "HTTP not found"}
	HttpInternal       = &Error{Code: 500, Message: "HTTP internal service error"}
	HttpBadRequest     = &Error{Code: 400, Message: "HTTP bad request"}
	HttpNotImplemented = &Error{Code: 501, Message: "HTTP not implemented"}
	HttpProcessing     = &Error{Code: 502, Message: "HTTP processing"}
	HttpTimeOut        = &Error{Code: 408, Message: "HTTP timeout"}
	NotFound           = &Error{Code: 1, Message: "not found"}
	Unauthorized       = &Error{Code: 2, Message: "unauthorized"}
	NotSupported       = &Error{Code: 3, Message: "not supported"}
	Duplicate          = &Error{Code: 4, Message: "duplicate key"}
	Unexpected         = &Error{Code: 5, Message: "unexpected"}
	WrongContent       = &Error{Code: 6, Message: "wrong content"}
	BadInput           = &Error{Code: 7, Message: "bad input"}
	TimeOut            = &Error{Code: 8, Message: "time out"}
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func New(message string) error {
	return &Error{
		Code:    HttpInternal.Code,
		Message: message,
	}
}

func Parse(e error) *Error {
	return ParseString(e.Error())
}

func ParseString(str string) *Error {
	switch str {
	case HttpForbidden.Message:
		return HttpForbidden
	case HttpUnauthorized.Message:
		return HttpUnauthorized
	case HttpNotFound.Message:
		return HttpNotFound
	case HttpBadRequest.Message:
		return HttpBadRequest
	case HttpNotImplemented.Message:
		return HttpNotImplemented
	case HttpProcessing.Message:
		return HttpProcessing
	case HttpTimeOut.Message:
		return HttpTimeOut
	case NotFound.Message:
		return NotFound
	case Unauthorized.Message:
		return NotFound
	case Unexpected.Message:
		return NotFound
	case Duplicate.Message:
		return Duplicate
	case NotSupported.Message:
		return NotSupported
	case WrongContent.Message:
		return WrongContent
	case TimeOut.Message:
		return TimeOut
	case BadInput.Message:
		return BadInput
	default:
		return HttpInternal
	}
}

func IsHttpNotFound(e error) bool {
	return Parse(e) == HttpNotFound
}

func IsHttpUnauthorized(e error) bool {
	return Parse(e) == HttpUnauthorized
}

func IsNotFound(e error) bool {
	return Parse(e) == NotFound
}

func IsTimeOut(e error) bool {
	return Parse(e) == TimeOut
}

func IsDuplicate(e error) bool {
	return Parse(e) == Duplicate
}
