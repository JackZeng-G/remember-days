package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"anniversary/internal/model"
)

// JSONStore JSON文件存储实现
type JSONStore struct {
	filePath string
	mu       sync.RWMutex
}

// NewJSONStore 创建新的JSON存储
func NewJSONStore(dataDir string) *JSONStore {
	return &JSONStore{
		filePath: filepath.Join(dataDir, "anniversaries.json"),
	}
}

// Load 加载所有纪念日
func (s *JSONStore) Load() ([]model.Anniversary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 如果文件不存在，返回空列表
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []model.Anniversary{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("读取数据文件失败: %w", err)
	}

	var storage struct {
		Anniversaries []model.Anniversary `json:"anniversaries"`
	}

	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	return storage.Anniversaries, nil
}

// Save 保存所有纪念日
func (s *JSONStore) Save(anniversaries []model.Anniversary) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	storage := struct {
		Anniversaries []model.Anniversary `json:"anniversaries"`
	}{
		Anniversaries: anniversaries,
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 使用0600权限，仅所有者读写
	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return fmt.Errorf("写入数据文件失败: %w", err)
	}

	return nil
}

// Add 添加纪念日
func (s *JSONStore) Add(ann model.Anniversary) error {
	anns, err := s.Load()
	if err != nil {
		return err
	}

	anns = append(anns, ann)
	return s.Save(anns)
}

// Update 更新纪念日
func (s *JSONStore) Update(ann model.Anniversary) error {
	anns, err := s.Load()
	if err != nil {
		return err
	}

	for i, a := range anns {
		if a.ID == ann.ID {
			anns[i] = ann
			return s.Save(anns)
		}
	}

	return fmt.Errorf("纪念日不存在: %s", ann.ID)
}

// Delete 删除纪念日
func (s *JSONStore) Delete(id string) error {
	anns, err := s.Load()
	if err != nil {
		return err
	}

	for i, a := range anns {
		if a.ID == id {
			anns = append(anns[:i], anns[i+1:]...)
			return s.Save(anns)
		}
	}

	return fmt.Errorf("纪念日不存在: %s", id)
}
