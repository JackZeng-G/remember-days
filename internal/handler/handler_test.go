package handler

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"remember/internal/service"
	"remember/internal/store"
)

const testUserID = "test0001"

func setupTestHandler() (*chi.Mux, *service.SessionManager, error) {
	memStore := store.NewMemoryStore()
	annSvc := service.New(memStore)

	webFS := os.DirFS("../..")
	tmpl, err := NewTemplateRendererFromFS(webFS)
	if err != nil {
		return nil, nil, err
	}

	sessionMgr := service.NewSessionManager("test-secret-key-32chars-long-enough")

	logger := log.New(io.Discard, "", 0)
	h := New(annSvc, nil, sessionMgr, tmpl, logger)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	return r, sessionMgr, nil
}

func TestHandler_APIStatus(t *testing.T) {
	r, _, err := setupTestHandler()
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

func TestHandler_APIReminders_RequiresAuth(t *testing.T) {
	r, _, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/reminders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want 401", w.Code)
	}
}

func TestHandler_APIReminders_WithAuth(t *testing.T) {
	r, sessionMgr, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 创建一个带 session cookie 的请求
	req := httptest.NewRequest(http.MethodGet, "/api/reminders", nil)
	w := httptest.NewRecorder()
	sessionMgr.CreateSession(w, testUserID)

	// 从响应中提取 cookie 并设置到下一个请求
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		req.AddCookie(c)
	}

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
