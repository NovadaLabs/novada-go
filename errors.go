package novada

import (
	"errors"
	"fmt"
)

// APIError represents a failed Novada API call. It is returned in two
// situations:
//
//   - the HTTP response status was outside the 2xx range, in which case
//     HTTPStatus is set and Code is 0; or
//   - the HTTP status was 2xx but the response envelope carried a non-zero
//     business code, in which case both HTTPStatus and Code are set.
//
// Callers should use the IsAuthError, IsRateLimited and CodeOf helpers rather
// than inspecting the fields directly.
type APIError struct {
	// HTTPStatus is the HTTP status code of the response.
	HTTPStatus int
	// Code is the Novada business code carried in the response envelope.
	// A value of 0 means success; any other value is a failure. It is 0 when
	// the failure occurred at the HTTP layer (e.g. a 5xx without a parseable
	// envelope).
	Code int
	// Message is the human-readable error message (the envelope "msg" field
	// when available).
	Message string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("novada: api error (code=%d, http=%d): %s", e.Code, e.HTTPStatus, e.Message)
	}
	return fmt.Sprintf("novada: http error (status=%d): %s", e.HTTPStatus, e.Message)
}

// IsAuthError reports whether err is an *APIError caused by an authentication
// or authorization failure (HTTP 401 or 403).
func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.HTTPStatus == 401 || apiErr.HTTPStatus == 403
	}
	return false
}

// IsRateLimited reports whether err is an *APIError caused by rate limiting
// (HTTP 429).
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.HTTPStatus == 429
	}
	return false
}

// CodeOf returns the Novada business code carried by err. The boolean result
// is false when err is not an *APIError.
func CodeOf(err error) (int, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code, true
	}
	return 0, false
}
