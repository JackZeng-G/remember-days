package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

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

// Storage 数据存储结构
type Storage struct {
	Anniversaries []Anniversary `json:"anniversaries"`
}

// GetAnniversaryView 计算纪念日的视图数据
func GetAnniversaryView(anniversaries []Anniversary) []AnniversaryView {
	var views []AnniversaryView
	now := time.Now()
	currentYear := now.Year()

	for _, a := range anniversaries {
		// 解析日期 (YYYY-MM-DD 或 MM-DD 格式)
		var year, month, day int
		hasYear := true
		_, err := fmt.Sscanf(a.Date, "%d-%d-%d", &year, &month, &day)
		if err != nil {
			// 尝试旧格式 MM-DD
			hasYear = false
			_, err = fmt.Sscanf(a.Date, "%d-%d", &month, &day)
			if err != nil {
				continue
			}
		}

		// 计算下次纪念日
		nextOccurrence := time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local)
		if nextOccurrence.Before(now) {
			nextOccurrence = time.Date(currentYear+1, time.Month(month), day, 0, 0, 0, 0, time.Local)
		}

		// 计算距离天数
		daysUntil := int(time.Until(nextOccurrence).Hours() / 24)

		// 计算至今多少天和第几个纪念日
		daysPassed := 0
		anniversaryCount := 0
		if hasYear {
			// 计算从原始日期到今天的天数
			originalDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			daysPassed = int(now.Sub(originalDate).Hours() / 24)

			// 计算第几个纪念日（过了多少个整年 + 1）
			anniversaryCount = currentYear - year
			if time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local).After(now) {
				anniversaryCount-- // 今年的还没到
			}
			anniversaryCount++ // 第N个纪念日
		}

		views = append(views, AnniversaryView{
			Anniversary:      a,
			DaysUntil:        daysUntil,
			IsUpcoming:       daysUntil <= 7 && daysUntil > 0,
			NextOccurrence:   nextOccurrence,
			DaysPassed:       daysPassed,
			AnniversaryCount: anniversaryCount,
			HasYear:          hasYear,
		})
	}

	return views
}

// AddAnniversary 添加纪念日
func AddAnniversary(storage *Storage, name, date, description string) {
	a := Anniversary{
		ID:          uuid.New().String()[:8],
		Name:        name,
		Date:        date,
		Description: description,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
	storage.Anniversaries = append(storage.Anniversaries, a)
}

// DeleteAnniversary 删除纪念日
func DeleteAnniversary(storage *Storage, id string) {
	for i, a := range storage.Anniversaries {
		if a.ID == id {
			storage.Anniversaries = append(storage.Anniversaries[:i], storage.Anniversaries[i+1:]...)
			break
		}
	}
}
