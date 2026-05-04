package store

import (
	"fmt"
	"sync"

	"remember/internal/model"
)

// MemoryStore 内存存储实现，用于测试
type MemoryStore struct {
	data map[string]model.Anniversary
	mu   sync.RWMutex
}

// NewMemoryStore 创建新的内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]model.Anniversary),
	}
}

// Load 加载所有纪念日
func (s *MemoryStore) Load() ([]model.Anniversary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]model.Anniversary, 0, len(s.data))
	for _, ann := range s.data {
		result = append(result, ann)
	}
	return result, nil
}

// Save 保存所有纪念日
func (s *MemoryStore) Save(anniversaries []model.Anniversary) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]model.Anniversary, len(anniversaries))
	for _, ann := range anniversaries {
		s.data[ann.ID] = ann
	}
	return nil
}

// Add 添加纪念日
func (s *MemoryStore) Add(ann model.Anniversary) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[ann.ID] = ann
	return nil
}


// Delete 删除纪念日
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[id]; !exists {
		return fmt.Errorf("纪念日不存在: %s", id)
	}

	delete(s.data, id)
	return nil
}