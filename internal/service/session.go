package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sessionCookieName = "session_token"
	sessionMaxAge     = 86400 * 7 // 7天
)

// SessionManager 会话管理器
type SessionManager struct {
	secret []byte
}

// NewSessionManager 创建会话管理器
func NewSessionManager(secret string) *SessionManager {
	return &SessionManager{secret: []byte(secret)}
}

// CreateSession 创建会话
func (sm *SessionManager) CreateSession(w http.ResponseWriter, userID string) error {
	expiry := time.Now().Add(time.Duration(sessionMaxAge) * time.Second).Unix()

	expiryBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expiryBytes, uint64(expiry))

	mac := hmac.New(sha256.New, sm.secret)
	mac.Write([]byte(userID))
	mac.Write(expiryBytes)
	sig := mac.Sum(nil)

	token := fmt.Sprintf("%s:%d:%s",
		base64.RawURLEncoding.EncodeToString([]byte(userID)),
		expiry,
		base64.RawURLEncoding.EncodeToString(sig),
	)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   sessionMaxAge,
	})

	return nil
}

// GetSession 获取会话中的用户ID
func (sm *SessionManager) GetSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", fmt.Errorf("未找到会话")
	}

	parts := strings.SplitN(cookie.Value, ":", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("会话格式无效")
	}

	userIDBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("会话解码失败")
	}
	userID := string(userIDBytes)

	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("会话过期时间无效")
	}

	if time.Now().Unix() > expiry {
		return "", fmt.Errorf("会话已过期")
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("会话签名无效")
	}

	expiryBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expiryBytes, uint64(expiry))

	mac := hmac.New(sha256.New, sm.secret)
	mac.Write([]byte(userID))
	mac.Write(expiryBytes)

	if !hmac.Equal(mac.Sum(nil), sig) {
		return "", fmt.Errorf("会话签名不匹配")
	}

	return userID, nil
}

// DestroySession 销毁会话
func (sm *SessionManager) DestroySession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
