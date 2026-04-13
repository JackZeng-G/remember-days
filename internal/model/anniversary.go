package model

import "time"

// Anniversary 纪念日数据结构
type Anniversary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Date        string `json:"date"` // YYYY-MM-DD 格式
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// AnniversaryView 纪念日视图（包含计算字段）
type AnniversaryView struct {
	Anniversary
	DaysUntil        int       `json:"days_until"`
	IsUpcoming       bool      `json:"is_upcoming"`
	NextOccurrence   time.Time `json:"next_occurrence"`
	DaysPassed       int       `json:"days_passed"`       // 至今多少天
	AnniversaryCount int       `json:"anniversary_count"` // 第几个纪念日
	HasYear          bool      `json:"has_year"`          // 是否有年份信息
}
