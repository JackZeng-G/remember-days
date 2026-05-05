package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

	"remember/internal/service"
)

type contextKey string

const ctxUserIDKey contextKey = "user_id"
const ctxCSRFKey contextKey = "csrf_token"

const csrfCookieName = "csrf_token"

func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)
}

func setCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}

func GetCSRFCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetCSRFToken 获取 CSRF token（优先从 context，fallback 从 cookie）
func GetCSRFToken(r *http.Request) string {
	if v, ok := r.Context().Value(ctxCSRFKey).(string); ok && v != "" {
		return v
	}
	if v, err := GetCSRFCookie(r); err == nil && v != "" {
		return v
	}
	return ""
}

func GetUserID(r *http.Request) string {
	if v, ok := r.Context().Value(ctxUserIDKey).(string); ok {
		return v
	}
	return ""
}

// LoggingMiddleware 日志中间件（跳过静态资源和频繁请求）
func LoggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			if !strings.HasPrefix(r.URL.Path, "/static/") && !strings.HasPrefix(r.URL.Path, "/api/") {
				logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
			}
		})
	}
}

// CSRFMiddleware CSRF 保护中间件
func CSRFMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" || r.Method == "HEAD" {
				token, err := GetCSRFCookie(r)
				if err != nil || token == "" {
					token = generateCSRFToken()
					setCSRFCookie(w, token)
				}
				ctx := context.WithValue(r.Context(), ctxCSRFKey, token)
					next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			cookieToken, err := GetCSRFCookie(r)
			if err != nil {
				http.Error(w, "CSRF Token Missing", http.StatusForbidden)
				return
			}

			formToken := r.FormValue("_csrf")
			if formToken == "" {
				formToken = r.Header.Get("X-CSRF-Token")
			}

			if formToken != cookieToken {
				http.Error(w, "CSRF Token Invalid", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware 认证中间件
func AuthMiddleware(sessionMgr *service.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := sessionMgr.GetSession(r)
			if err != nil || userID == "" {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware 管理员中间件
func AdminMiddleware(userSvc service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)
			user, err := userSvc.GetByID(userID)
			if err != nil || !user.IsAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
