package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// 数据文件路径
const dataFile = "data/anniversaries.json"

// LoadData 加载数据
func LoadData() (*Storage, error) {
	// 确保目录存在
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 读取文件
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回空数据
			return &Storage{Anniversaries: []Anniversary{}}, nil
		}
		return nil, err
	}

	// 解析JSON
	var storage Storage
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, err
	}

	return &storage, nil
}

// SaveData 保存数据
func SaveData(storage *Storage) error {
	// 确保目录存在
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(dataFile, data, 0644)
}
