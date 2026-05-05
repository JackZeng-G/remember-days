package model

import "time"

// Anniversary 纪念日数据结构
type Anniversary struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Date        string `json:"date"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// AnniversaryView 纪念日视图（包含计算字段）
type AnniversaryView struct {
	Anniversary
	DaysUntil        int       `json:"days_until"`
	IsUpcoming       bool      `json:"is_upcoming"`
	NextOccurrence   time.Time `json:"next_occurrence"`
	DaysPassed       int       `json:"days_passed"`
	AnniversaryCount int       `json:"anniversary_count"`
}
