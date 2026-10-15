package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"resty.dev/v3"
)

func TestSendMetricsPostsKnownMetricURLs(t *testing.T) {
	var gotPaths []string
	var gotMethods []string
	var gotContentTypes []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		gotMethods = append(gotMethods, r.Method)
		gotContentTypes = append(gotContentTypes, r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldBaseURL := baseUrl
	baseUrl = server.URL + "/update/"
	defer func() { baseUrl = oldBaseURL }()

	counter := 7
	metrics := map[string]func() any{
		"gauge/Alloc":       func() any { return uint64(42) },
		"gauge/RandomValue": func() any { return 0.25 },
		"counter/PollCount": func() any { return counter },
	}

	sendMetrics(resty.NewWithClient(server.Client()), metrics)

	wantPaths := map[string]bool{
		"/update/gauge/Alloc/42":         false,
		"/update/gauge/RandomValue/0.25": false,
		"/update/counter/PollCount/7":    false,
	}

	if len(gotPaths) != len(wantPaths) {
		t.Fatalf("expected %d requests, got %d: %v", len(wantPaths), len(gotPaths), gotPaths)
	}

	for i, path := range gotPaths {
		if _, ok := wantPaths[path]; !ok {
			t.Fatalf("unexpected request path %q; all paths: %v", path, gotPaths)
		}
		wantPaths[path] = true

		if gotMethods[i] != http.MethodPost {
			t.Fatalf("expected request %q to use POST, got %s", path, gotMethods[i])
		}
		if gotContentTypes[i] != "text/plain" {
			t.Fatalf("expected request %q content type text/plain, got %q", path, gotContentTypes[i])
		}
	}

	for path, seen := range wantPaths {
		if !seen {
			t.Fatalf("expected request path %q, got %v", path, gotPaths)
		}
	}
}
