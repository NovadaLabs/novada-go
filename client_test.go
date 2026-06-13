package novada

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// whitelistItem mirrors the shape of an item in the white_list/list response
// example from the OpenAPI spec. It lives in the test only; the real typed
// model will live in the proxy package.
type whitelistItem struct {
	ID     int    `json:"id"`
	UID    int    `json:"uid"`
	MarkIP string `json:"mark_ip"`
	Status int    `json:"status"`
	Lock   bool   `json:"lock"`
	Mark   string `json:"mark"`
}

// newTestClient returns a Client pointed at srv with retries disabled so tests
// fail fast.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient("test-key",
		WithBaseURL(srv.URL),
		WithWebUnblockerURL(srv.URL),
		WithScraperURL(srv.URL),
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestDoMultipart_WhitelistList_Success(t *testing.T) {
	const body = `{
		"code": 0,
		"data": {
			"list": [
				{"created_at":1732084553,"id":12,"uid":81,"mark_ip":"10.10.10.1",
				 "status":1,"lock":false,"mark":"test"}
			],
			"total": 1
		},
		"msg": "success",
		"timestamp": 1732084616
	}`

	var (
		gotAuth        string
		gotContentType string
		gotProduct     string
		gotMethod      string
		gotPath        string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		gotProduct = r.FormValue("product")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)

	var out List[whitelistItem]
	err := c.DoMultipart(context.Background(), c.BaseURL(), "/v1/white_list/list",
		map[string]string{"product": "1"}, &out)
	if err != nil {
		t.Fatalf("DoMultipart: %v", err)
	}

	// Transport-level assertions.
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/v1/white_list/list" {
		t.Errorf("path = %q, want /v1/white_list/list", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("Content-Type = %q, want multipart/form-data", gotContentType)
	}
	if gotProduct != "1" {
		t.Errorf("product field = %q, want 1", gotProduct)
	}

	// Envelope/list unwrapping assertions.
	if out.Total != 1 {
		t.Errorf("total = %d, want 1", out.Total)
	}
	if len(out.List) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(out.List))
	}
	item := out.List[0]
	if item.ID != 12 || item.MarkIP != "10.10.10.1" || item.Mark != "test" {
		t.Errorf("item = %+v, unexpected values", item)
	}
}

func TestDoMultipart_BusinessCodeNonZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1001,"data":null,"msg":"invalid product","timestamp":1}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)

	var out List[whitelistItem]
	err := c.DoMultipart(context.Background(), c.BaseURL(), "/v1/white_list/list",
		map[string]string{"product": "99"}, &out)
	if err == nil {
		t.Fatal("expected error for code != 0, got nil")
	}
	code, ok := CodeOf(err)
	if !ok {
		t.Fatalf("CodeOf returned ok=false for %v", err)
	}
	if code != 1001 {
		t.Errorf("code = %d, want 1001", code)
	}
	if IsAuthError(err) || IsRateLimited(err) {
		t.Errorf("unexpected auth/ratelimit classification for %v", err)
	}
}

func TestDo_HTTPErrorStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		wantAuth    bool
		wantLimited bool
	}{
		{"unauthorized", http.StatusUnauthorized, true, false},
		{"forbidden", http.StatusForbidden, true, false},
		{"rate limited", http.StatusTooManyRequests, false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"code":-1,"msg":"denied","timestamp":1}`))
			}))
			defer srv.Close()

			c := newTestClient(t, srv)
			err := c.DoMultipart(context.Background(), c.BaseURL(), "/v1/white_list/list",
				map[string]string{"product": "1"}, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := IsAuthError(err); got != tc.wantAuth {
				t.Errorf("IsAuthError = %v, want %v", got, tc.wantAuth)
			}
			if got := IsRateLimited(err); got != tc.wantLimited {
				t.Errorf("IsRateLimited = %v, want %v", got, tc.wantLimited)
			}
		})
	}
}

func TestNewClient_NoAPIKey(t *testing.T) {
	t.Setenv("NOVADA_API_KEY", "")
	if _, err := NewClient(""); err != ErrNoAPIKey {
		t.Fatalf("err = %v, want ErrNoAPIKey", err)
	}
}

func TestNewClient_EnvFallback(t *testing.T) {
	t.Setenv("NOVADA_API_KEY", "env-key")
	c, err := NewClient("")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.apiKey != "env-key" {
		t.Errorf("apiKey = %q, want env-key", c.apiKey)
	}
	// Defaults applied.
	if c.BaseURL() != DefaultBaseURL ||
		c.WebUnblockerURL() != DefaultWebUnblockerURL ||
		c.ScraperURL() != DefaultScraperURL {
		t.Errorf("default base URLs not applied: %q %q %q",
			c.BaseURL(), c.WebUnblockerURL(), c.ScraperURL())
	}
}

func TestDoFormURLEncoded_Encoding(t *testing.T) {
	var gotBody, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{},"msg":"success","timestamp":1}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)

	params := []map[string]any{{"url": "https://example.com/watch?v=1"}}
	encoded, _ := json.Marshal(params)
	form := map[string][]string{
		"scraper_id":     {"youtube_video-post_explore"},
		"scraper_params": {string(encoded)},
	}
	if err := c.DoFormURLEncoded(context.Background(), c.ScraperURL(), "/request", form, nil); err != nil {
		t.Fatalf("DoFormURLEncoded: %v", err)
	}
	if gotCT != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q", gotCT)
	}
	if !strings.Contains(gotBody, "scraper_id=youtube_video-post_explore") {
		t.Errorf("body missing scraper_id: %q", gotBody)
	}
	// The JSON params must be URL-encoded inside the form value.
	if !strings.Contains(gotBody, "scraper_params=%5B%7B%22url%22") {
		t.Errorf("scraper_params not URL-encoded as expected: %q", gotBody)
	}
}

func TestDo_RetriesThenSucceeds(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusBadGateway) // 502: retryable
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{},"msg":"success","timestamp":1}`))
	}))
	defer srv.Close()

	c, err := NewClient("k",
		WithBaseURL(srv.URL),
		WithMaxRetries(3),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := c.DoMultipart(context.Background(), c.BaseURL(), "/v1/white_list/list",
		map[string]string{"product": "1"}, nil); err != nil {
		t.Fatalf("DoMultipart after retries: %v", err)
	}
	if calls != 3 {
		t.Errorf("server received %d calls, want 3", calls)
	}
}

func TestDo_RetriesExhausted(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":-1,"msg":"boom","timestamp":1}`))
	}))
	defer srv.Close()

	c, _ := NewClient("k", WithBaseURL(srv.URL), WithMaxRetries(1))
	err := c.DoMultipart(context.Background(), c.BaseURL(), "/v1/white_list/list",
		map[string]string{"product": "1"}, nil)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("err = %v, want *APIError with HTTPStatus 500", err)
	}
	if apiErr.Message != "boom" {
		t.Errorf("message = %q, want %q (from envelope msg)", apiErr.Message, "boom")
	}
	if calls != 2 { // initial + 1 retry
		t.Errorf("server received %d calls, want 2", calls)
	}
}

func TestDoMultipartRaw_ReturnsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("content-type = %q", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("id,ip\n1,1.2.3.4\n"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	body, err := c.DoMultipartRaw(context.Background(), c.BaseURL(), "/v1/static/export",
		map[string]string{"status": "1"})
	if err != nil {
		t.Fatalf("DoMultipartRaw: %v", err)
	}
	if string(body) != "id,ip\n1,1.2.3.4\n" {
		t.Errorf("body = %q", body)
	}
}

func TestDoFormURLEncodedRaw_ReturnsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("content-type = %q", r.Header.Get("Content-Type"))
		}
		_, _ = w.Write([]byte(`{"scraped":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	body, err := c.DoFormURLEncodedRaw(context.Background(), c.ScraperURL(), "/request",
		url.Values{"scraper_id": {"x"}})
	if err != nil {
		t.Fatalf("DoFormURLEncodedRaw: %v", err)
	}
	if string(body) != `{"scraped":true}` {
		t.Errorf("body = %q", body)
	}
}

func TestDoRaw_HTTPErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.DoMultipartRaw(context.Background(), c.BaseURL(), "/v1/x", nil); !IsAuthError(err) {
		t.Errorf("err = %v, want auth error", err)
	}
	if _, err := c.DoFormURLEncodedRaw(context.Background(), c.BaseURL(), "/x", nil); !IsAuthError(err) {
		t.Errorf("err = %v, want auth error", err)
	}
}

func TestJoinURL(t *testing.T) {
	cases := []struct{ base, path, want string }{
		{"https://h.com", "/v1/x", "https://h.com/v1/x"},
		{"https://h.com/", "/v1/x", "https://h.com/v1/x"},
		{"https://h.com/", "v1/x", "https://h.com/v1/x"},
	}
	for _, tc := range cases {
		if got := joinURL(tc.base, tc.path); got != tc.want {
			t.Errorf("joinURL(%q,%q) = %q, want %q", tc.base, tc.path, got, tc.want)
		}
	}
}

func TestOptions_Apply(t *testing.T) {
	hc := &http.Client{}
	c, err := NewClient("k",
		WithTimeout(5*1e9),
		WithHTTPClient(hc),
		WithUserAgent("custom-agent/1.0"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.httpClient != hc {
		t.Error("WithHTTPClient not applied")
	}
	if c.userAgent != "custom-agent/1.0" {
		t.Errorf("userAgent = %q", c.userAgent)
	}
}

func TestAPIError_Error(t *testing.T) {
	business := &APIError{HTTPStatus: 200, Code: 1001, Message: "bad product"}
	if !strings.Contains(business.Error(), "code=1001") {
		t.Errorf("business error string = %q", business.Error())
	}
	httpErr := &APIError{HTTPStatus: 500, Message: "boom"}
	if !strings.Contains(httpErr.Error(), "status=500") {
		t.Errorf("http error string = %q", httpErr.Error())
	}
}
