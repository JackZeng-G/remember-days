package store

import "remember/internal/model"

// AnniversaryStore 存储接口
type AnniversaryStore interface {
	Load() ([]model.Anniversary, error)
	Save(anniversaries []model.Anniversary) error
	Add(ann model.Anniversary) error
	Update(ann model.Anniversary) error
	Delete(id string) error
}