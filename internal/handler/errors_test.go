package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerFuncWritesHTTPErrorStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return HTTPError{Code: http.StatusNotFound, Msg: "missing"}
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestHTTPErrorErrorReturnsMessage(t *testing.T) {
	err := HTTPError{Msg: "invalid value"}

	if err.Error() != "invalid value" {
		t.Fatalf("expected error message %q, got %q", "invalid value", err.Error())
	}
}
