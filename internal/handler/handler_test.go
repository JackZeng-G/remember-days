package handler

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"remember/internal/service"
	"remember/internal/store"
)

func setupTestHandler() (*chi.Mux, error) {
	memStore := store.NewMemoryStore()
	svc := service.New(memStore)

	tmpl, err := NewTemplateRenderer("../../web/templates/*.html")
	if err != nil {
		return nil, err
	}

	logger := log.New(io.Discard, "", 0)
	h := New(svc, tmpl, logger)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	return r, nil
}

func TestHandler_APIReminders(t *testing.T) {
	r, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Create a test anniversary first
	req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader("name=Test&year=2024&month=01&day=01&description=Test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Now test API
	req = httptest.NewRequest(http.MethodGet, "/api/reminders", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}
}

func TestHandler_APIStatus(t *testing.T) {
	r, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}
