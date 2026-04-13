package handler

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"anniversary/internal/config"
)

const csrfCookieName = "csrf_token"

// GenerateCSRFToken 生成 CSRF Token
func GenerateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// SetCSRFCookie 设置 CSRF Cookie
func SetCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}

// GetCSRFCookie 获取 CSRF Cookie
func GetCSRFCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
		})
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Printf("PANIC: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CSRFMiddleware CSRF 保护中间件
func CSRFMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// GET/HEAD 请求：设置 CSRF Token
			if r.Method == "GET" || r.Method == "HEAD" {
				token, err := GetCSRFCookie(r)
				if err != nil || token == "" {
					token = GenerateCSRFToken()
					SetCSRFCookie(w, token)
				}
				next.ServeHTTP(w, r)
				return
			}

			// POST/PUT/DELETE 请求：验证 CSRF Token
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