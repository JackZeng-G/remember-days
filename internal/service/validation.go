package service

import (
	"regexp"
	"strings"
	"time"
)

var validID = regexp.MustCompile(`^[a-zA-Z0-9]{8}$`)

// isValidID 验证ID格式
func isValidID(id string) bool {
	return validID.MatchString(id)
}

// ValidateName 验证名称
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 100 {
		return ErrInvalidInput
	}
	return nil
}

// ValidateDate 验证日期格式
func ValidateDate(date string) error {
	// 支持 YYYY-MM-DD 格式
	if len(date) == 10 {
		_, err := time.Parse("2006-01-02", date)
		if err == nil {
			return nil
		}
	}
	return ErrInvalidDate
}

// SanitizeForLog 清理日志输入
func SanitizeForLog(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}
