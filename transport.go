package novada

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// retryBaseDelay is the base backoff between retry attempts; it is multiplied
// by the attempt index to produce a simple linear backoff.
const retryBaseDelay = 200 * time.Millisecond

// DoMultipart sends a multipart/form-data POST to baseURL+path, the encoding
// used by every /v1/* management endpoint, and decodes the response envelope
// into out. Fields with empty values are omitted. baseURL is typically
// c.BaseURL().
func (c *Client) DoMultipart(ctx context.Context, baseURL, path string, fields map[string]string, out any) error {
	body, contentType, err := buildMultipart(fields)
	if err != nil {
		return err
	}
	return c.do(ctx, joinURL(baseURL, path), contentType, body, out)
}

// DoMultipartRaw is like DoMultipart but returns the raw response body without
// envelope decoding. It is used by endpoints that return a file stream (e.g.
// the static IP export endpoints) rather than the standard JSON envelope.
func (c *Client) DoMultipartRaw(ctx context.Context, baseURL, path string, fields map[string]string) ([]byte, error) {
	body, contentType, err := buildMultipart(fields)
	if err != nil {
		return nil, err
	}
	respBody, _, err := c.doRaw(ctx, joinURL(baseURL, path), contentType, body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

// DoFormURLEncoded sends an application/x-www-form-urlencoded POST to
// baseURL+path, the encoding used by the scraper /request endpoints, and
// decodes the response envelope into out.
func (c *Client) DoFormURLEncoded(ctx context.Context, baseURL, path string, values url.Values, out any) error {
	body := []byte(values.Encode())
	return c.do(ctx, joinURL(baseURL, path), "application/x-www-form-urlencoded", body, out)
}

// DoFormURLEncodedRaw is like DoFormURLEncoded but returns the raw response
// body without envelope decoding. Scraper /request responses are the scraped
// payload itself (JSON, CSV or XLSX) rather than the management envelope.
func (c *Client) DoFormURLEncodedRaw(ctx context.Context, baseURL, path string, values url.Values) ([]byte, error) {
	body := []byte(values.Encode())
	respBody, _, err := c.doRaw(ctx, joinURL(baseURL, path), "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

// buildMultipart encodes fields as a multipart/form-data body, omitting empty
// values, and returns the body bytes together with the content type.
func buildMultipart(fields map[string]string) ([]byte, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if v == "" {
			continue
		}
		if err := w.WriteField(k, v); err != nil {
			return nil, "", err
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

// do performs the HTTP request and hands a successful response to the envelope
// decoder.
func (c *Client) do(ctx context.Context, fullURL, contentType string, body []byte, out any) error {
	respBody, status, err := c.doRaw(ctx, fullURL, contentType, body)
	if err != nil {
		return err
	}
	return decodeEnvelope(status, respBody, out)
}

// doRaw performs the HTTP request with Bearer injection and retry handling and
// returns the raw response body and status. The body bytes are reused across
// retry attempts. A non-2xx response (after retries) is returned as an
// *APIError alongside the body.
func (c *Client) doRaw(ctx context.Context, fullURL, contentType string, body []byte) ([]byte, int, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(retryBaseDelay * time.Duration(attempt)):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(body))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue // network error: retry
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = &APIError{
				HTTPStatus: resp.StatusCode,
				Message:    httpErrorMessage(respBody, resp.Status),
			}
			continue // 429/5xx: retry
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return respBody, resp.StatusCode, &APIError{
				HTTPStatus: resp.StatusCode,
				Message:    httpErrorMessage(respBody, resp.Status),
			}
		}

		return respBody, resp.StatusCode, nil
	}

	return nil, 0, lastErr
}

// httpErrorMessage extracts a human-readable message for a non-2xx response,
// preferring the envelope "msg" field when the body parses, otherwise falling
// back to the HTTP status line.
func httpErrorMessage(body []byte, status string) string {
	var env envelope
	if err := json.Unmarshal(body, &env); err == nil && env.Msg != "" {
		return env.Msg
	}
	return status
}

// joinURL concatenates a base URL and a path, tolerating a trailing slash on
// the base and a missing leading slash on the path.
func joinURL(base, path string) string {
	base = strings.TrimRight(base, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}
