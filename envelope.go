package novada

import (
	"encoding/json"
	"fmt"
)

// envelope is the uniform wrapper returned by every /v1 management endpoint:
// {"code":0,"data":{...},"msg":"success","timestamp":1732084616}. A code of 0
// indicates success; any other value indicates a business-level failure.
type envelope struct {
	Code      int             `json:"code"`
	Msg       string          `json:"msg"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

// List is the common shape of the "data" field on list endpoints, e.g.
// {"list":[...],"total":1}. Decode into List[T] to unwrap the items.
type List[T any] struct {
	List  []T `json:"list"`
	Total int `json:"total"`
}

// decodeEnvelope parses a 2xx management response body, enforces the business
// code, and unmarshals the inner data into out. It returns an *APIError when
// the envelope carries a non-zero code, so callers must never treat HTTP 200
// as success on its own.
func decodeEnvelope(httpStatus int, body []byte, out any) error {
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return &APIError{
			HTTPStatus: httpStatus,
			Message:    fmt.Sprintf("invalid response body: %v", err),
		}
	}
	if env.Code != 0 {
		return &APIError{
			HTTPStatus: httpStatus,
			Code:       env.Code,
			Message:    env.Msg,
		}
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("novada: decode data: %w", err)
		}
	}
	return nil
}
