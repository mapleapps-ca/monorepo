package httperror

// This package introduces a new `error` type that combines an HTTP status code and a message.

import (
	"encoding/json"
	"errors"
	"net/http"
)

// HTTPError represents an http error that occurred while handling a request
type HTTPError struct {
	Code   int                `json:"-"` // HTTP Status code. We use `-` to skip json marshaling.
	Errors *map[string]string `json:"-"` // The original error. Same reason as above.
}

// New creates a new HTTPError instance with a multi-field errors.
func New(statusCode int, errorsMap *map[string]string) error {
	return HTTPError{
		Code:   statusCode,
		Errors: errorsMap,
	}
}

// NewForSingleField create a new HTTPError instance for a single field. This is a convinience constructor.
func NewForSingleField(statusCode int, field string, message string) error {
	return HTTPError{
		Code:   statusCode,
		Errors: &map[string]string{field: message},
	}
}

// NewForBadRequest create a new HTTPError instance pertaining to 403 bad requests with the multi-errors. This is a convinience constructor.
func NewForBadRequest(err *map[string]string) error {
	return HTTPError{
		Code:   http.StatusBadRequest,
		Errors: err,
	}
}

// NewForBadRequestWithSingleField create a new HTTPError instance pertaining to 403 bad requests for a single field. This is a convinience constructor.
func NewForBadRequestWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusBadRequest,
		Errors: &map[string]string{field: message},
	}
}

func NewForInternalServerErrorWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusInternalServerError,
		Errors: &map[string]string{field: message},
	}
}

// NewForNotFoundWithSingleField create a new HTTPError instance pertaining to 404 not found for a single field. This is a convinience constructor.
func NewForNotFoundWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusNotFound,
		Errors: &map[string]string{field: message},
	}
}

// NewForServiceUnavailableWithSingleField create a new HTTPError instance pertaining service unavailable for a single field. This is a convinience constructor.
func NewForServiceUnavailableWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusServiceUnavailable,
		Errors: &map[string]string{field: message},
	}
}

// NewForLockedWithSingleField create a new HTTPError instance pertaining to 424 locked for a single field. This is a convinience constructor.
func NewForLockedWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusLocked,
		Errors: &map[string]string{field: message},
	}
}

// NewForForbiddenWithSingleField create a new HTTPError instance pertaining to 403 bad requests for a single field. This is a convinience constructor.
func NewForForbiddenWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusForbidden,
		Errors: &map[string]string{field: message},
	}
}

// NewForUnauthorizedWithSingleField create a new HTTPError instance pertaining to 401 unauthorized for a single field. This is a convinience constructor.
func NewForUnauthorizedWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusUnauthorized,
		Errors: &map[string]string{field: message},
	}
}

// NewForGoneWithSingleField create a new HTTPError instance pertaining to 410 gone for a single field. This is a convinience constructor.
func NewForGoneWithSingleField(field string, message string) error {
	return HTTPError{
		Code:   http.StatusGone,
		Errors: &map[string]string{field: message},
	}
}

// Error function used to implement the `error` interface for returning errors.
func (err HTTPError) Error() string {
	b, e := json.Marshal(err.Errors)
	if e != nil { // Defensive code
		return e.Error()
	}
	return string(b)
}

// ResponseError function returns the HTTP error response based on the httpcode used.
func ResponseError(rw http.ResponseWriter, err error) {
	// Copied from:
	// https://dev.to/tigorlazuardi/go-creating-custom-error-wrapper-and-do-proper-error-equality-check-11k7

	rw.Header().Set("Content-Type", "Application/json")

	//
	// CASE 1 OF 2: Handle API Errors.
	//

	var ew HTTPError
	if errors.As(err, &ew) {
		rw.WriteHeader(ew.Code)
		_ = json.NewEncoder(rw).Encode(ew.Errors)
		return
	}

	//
	// CASE 2 OF 2: Handle non ErrorWrapper types.
	//

	rw.WriteHeader(http.StatusInternalServerError)

	_ = json.NewEncoder(rw).Encode(err.Error())
}

// NewForInternalServerError create a new HTTPError instance pertaining to 500 internal server error with the multi-errors. This is a convinience constructor.
func NewForInternalServerError(err string) error {
	return HTTPError{
		Code:   http.StatusInternalServerError,
		Errors: &map[string]string{"message": err},
	}
}
