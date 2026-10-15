package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestLogRequestsLogsMethodAndPath(t *testing.T) {
	var output bytes.Buffer
	logger := zerolog.New(&output)
	called := false

	router := gin.New()
	router.Use(LogRequests(&logger))
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

	var entry struct {
		Level      string   `json:"level"`
		URI        string   `json:"uri"`
		Method     string   `json:"method"`
		DurationMS *float64 `json:"duration_ms"`
		Status     int      `json:"status"`
		Size       *int     `json:"size"`
	}
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("expected a JSON log entry, got %q: %v", output.String(), err)
	}
	if entry.Level != "info" {
		t.Fatalf("expected Info log level, got %q", entry.Level)
	}
	if entry.URI != "/update/gauge/temperature/12.5?from=test" {
		t.Fatalf("expected full request URI, got %q", entry.URI)
	}
	if entry.Method != http.MethodPost {
		t.Fatalf("expected request method %q, got %q", http.MethodPost, entry.Method)
	}
	if entry.DurationMS == nil || *entry.DurationMS < 0 {
		t.Fatalf("expected a nonnegative request duration, got %v", entry.DurationMS)
	}
	if entry.Status != http.StatusAccepted {
		t.Fatalf("expected logged status %d, got %d", http.StatusAccepted, entry.Status)
	}
	if entry.Size == nil || *entry.Size != 0 {
		t.Fatalf("expected logged response size 0, got %v", entry.Size)
	}
}
