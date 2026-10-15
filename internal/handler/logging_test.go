package handler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogRequestsLogsMethodAndPath(t *testing.T) {
	var output bytes.Buffer
	logger := log.New(&output, "", 0)
	called := false

	router := gin.New()
	router.Use(LogRequests(logger))
	router.POST("/update/gauge/:metric/:value", func(c *gin.Context) {
		called = true
		c.Status(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/12.5?from=test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}

	logLine := output.String()
	if !strings.Contains(logLine, "method=POST") {
		t.Fatalf("expected log line to contain request method, got %q", logLine)
	}

	if !strings.Contains(logLine, "path=/update/gauge/temperature/12.5") {
		t.Fatalf("expected log line to contain request path, got %q", logLine)
	}
}
