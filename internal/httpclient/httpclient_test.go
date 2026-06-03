package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pinata/internal/version"
)

func TestClientSetsUserAgent(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	resp, err := Client.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	want := version.UserAgent()
	if got != want {
		t.Errorf("User-Agent = %q, want %q", got, want)
	}
}

func TestWithTimeoutSetsUserAgent(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	client := WithTimeout(3 * time.Second)
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if want := version.UserAgent(); got != want {
		t.Errorf("User-Agent = %q, want %q", got, want)
	}
}

// The transport must not mutate the caller's request (http.RoundTripper contract).
func TestTransportDoesNotMutateRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if ua := req.Header.Get("User-Agent"); ua != "" {
		t.Errorf("original request was mutated: User-Agent = %q", ua)
	}
}
