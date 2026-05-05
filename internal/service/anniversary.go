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
	List(userID string) ([]model.AnniversaryView, error)
	Get(userID, id string) (*model.AnniversaryView, error)
	Create(userID, name, date, desc string) (*model.Anniversary, error)
	Update(userID, id, name, date, desc string) error
	Delete(userID, id string) error
}

type anniversaryService struct {
	store store.AnniversaryStore
}

func New(s store.AnniversaryStore) AnniversaryService {
	return &anniversaryService{store: s}
}

func (s *anniversaryService) List(userID string) ([]model.AnniversaryView, error) {
	anns, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	var filtered []model.Anniversary
	for _, a := range anns {
		if a.UserID == userID {
			filtered = append(filtered, a)
		}
	}

	return calculateViews(filtered), nil
}

func (s *anniversaryService) Get(userID, id string) (*model.AnniversaryView, error) {
	if !isValidID(id) {
		return nil, ErrInvalidID
	}

	anns, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for _, ann := range anns {
		if ann.ID == id && ann.UserID == userID {
			view := calculateViews([]model.Anniversary{ann})[0]
			return &view, nil
		}
	}

	return nil, ErrNotFound
}

func (s *anniversaryService) Create(userID, name, date, desc string) (*model.Anniversary, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	if err := ValidateDate(date); err != nil {
		return nil, err
	}

	ann := model.Anniversary{
		ID:          uuid.New().String()[:8],
		UserID:      userID,
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

func (s *anniversaryService) Update(userID, id, name, date, desc string) error {
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
		if ann.ID == id && ann.UserID == userID {
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

func (s *anniversaryService) Delete(userID, id string) error {
	if !isValidID(id) {
		return ErrInvalidID
	}

	anns, err := s.store.Load()
	if err != nil {
		return err
	}

	for i, ann := range anns {
		if ann.ID == id && ann.UserID == userID {
			return s.store.Save(append(anns[:i], anns[i+1:]...))
		}
	}

	return ErrNotFound
}

func calculateViews(anniversaries []model.Anniversary) []model.AnniversaryView {
	var views []model.AnniversaryView
	now := time.Now()
	currentYear := now.Year()

	for _, a := range anniversaries {
		var year, month, day int
		_, err := fmt.Sscanf(a.Date, "%d-%d-%d", &year, &month, &day)
		if err != nil {
			continue
		}

		nextOccurrence := time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local)
		if nextOccurrence.Before(now) {
			nextOccurrence = time.Date(currentYear+1, time.Month(month), day, 0, 0, 0, 0, time.Local)
		}

		daysUntil := int(time.Until(nextOccurrence).Hours() / 24)

		originalDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		daysPassed := int(now.Sub(originalDate).Hours() / 24)

		anniversaryCount := currentYear - year
		if time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local).After(now) {
			anniversaryCount--
		}
		anniversaryCount++

		views = append(views, model.AnniversaryView{
			Anniversary:      a,
			DaysUntil:        daysUntil,
			IsUpcoming:       daysUntil <= 7 && daysUntil > 0,
			NextOccurrence:   nextOccurrence,
			DaysPassed:       daysPassed,
			AnniversaryCount: anniversaryCount,
		})
	}

	return views
}
