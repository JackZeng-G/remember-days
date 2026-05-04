package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"remember/internal/model"
	"remember/internal/store"
)

// AnniversaryService 服务接口
type AnniversaryService interface {
	List() ([]model.AnniversaryView, error)
	Get(id string) (*model.AnniversaryView, error)
	Create(name, date, desc string) (*model.Anniversary, error)
	Update(id, name, date, desc string) error
	Delete(id string) error
}

// anniversaryService 服务实现
type anniversaryService struct {
	store store.AnniversaryStore
}

// New 创建新的服务实例
func New(s store.AnniversaryStore) AnniversaryService {
	return &anniversaryService{store: s}
}

// List 获取所有纪念日视图
func (s *anniversaryService) List() ([]model.AnniversaryView, error) {
	anns, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	return calculateViews(anns), nil
}

// Get 获取单个纪念日视图
func (s *anniversaryService) Get(id string) (*model.AnniversaryView, error) {
	if !isValidID(id) {
		return nil, ErrInvalidID
	}

	anns, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for _, ann := range anns {
		if ann.ID == id {
			view := calculateViews([]model.Anniversary{ann})[0]
			return &view, nil
		}
	}

	return nil, ErrNotFound
}

// Create 创建新纪念日
func (s *anniversaryService) Create(name, date, desc string) (*model.Anniversary, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	if err := ValidateDate(date); err != nil {
		return nil, err
	}

	ann := model.Anniversary{
		ID:          uuid.New().String()[:8],
		Name:        name,
		Date:        date,
		Description: desc,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.store.Add(ann); err != nil {
		return nil, err
	}

	return &ann, nil
}

// Update 更新纪念日
func (s *anniversaryService) Update(id, name, date, desc string) error {
	if !isValidID(id) {
		return ErrInvalidID
	}
	if err := ValidateName(name); err != nil {
		return err
	}
	if err := ValidateDate(date); err != nil {
		return err
	}

	anns, err := s.store.Load()
	if err != nil {
		return err
	}

	var found bool
	for i, ann := range anns {
		if ann.ID == id {
			anns[i].Name = name
			anns[i].Date = date
			anns[i].Description = desc
			found = true
			break
		}
	}

	if !found {
		return ErrNotFound
	}

	return s.store.Save(anns)
}

// Delete 删除纪念日
func (s *anniversaryService) Delete(id string) error {
	if !isValidID(id) {
		return ErrInvalidID
	}

	return s.store.Delete(id)
}

// calculateViews 计算纪念日视图
func calculateViews(anniversaries []model.Anniversary) []model.AnniversaryView {
	var views []model.AnniversaryView
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

		views = append(views, model.AnniversaryView{
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

