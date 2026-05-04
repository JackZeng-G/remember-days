package store

import (
	"os"
	"path/filepath"
	"testing"

	"remember/internal/model"
)

func TestJSONStore_Load_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	anns, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(anns) != 0 {
		t.Errorf("Load() returned %d items, want 0", len(anns))
	}
}

func TestJSONStore_AddAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	ann := model.Anniversary{
		ID:          "test1234",
		Name:        "Test Anniversary",
		Date:        "2024-01-01",
		Description: "Test description",
		CreatedAt:   "2024-01-01 00:00:00",
	}

	if err := store.Add(ann); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	anns, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(anns) != 1 {
		t.Errorf("Load() returned %d items, want 1", len(anns))
	}
	if anns[0].ID != ann.ID {
		t.Errorf("Load() returned ID = %s, want %s", anns[0].ID, ann.ID)
	}
}


func TestJSONStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	ann := model.Anniversary{
		ID:   "test1234",
		Name: "Test",
		Date: "2024-01-01",
	}

	store.Add(ann)

	if err := store.Delete("test1234"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	anns, _ := store.Load()
	if len(anns) != 0 {
		t.Errorf("After Delete(), got %d items, want 0", len(anns))
	}
}

func TestJSONStore_FilePermission(t *testing.T) {
	// Skip on Windows - file permissions work differently on Windows
	if os.PathSeparator == '\\' {
		t.Skip("Skipping file permission test on Windows")
	}

	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	ann := model.Anniversary{
		ID:   "test1234",
		Name: "Test",
		Date: "2024-01-01",
	}

	store.Add(ann)

	filePath := filepath.Join(tmpDir, "remembers.json")
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// 检查文件权限是否为0600
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("File permission = %o, want 0600", mode)
	}
}

func TestMemoryStore_Basic(t *testing.T) {
	store := NewMemoryStore()

	ann := model.Anniversary{
		ID:   "test1234",
		Name: "Test",
		Date: "2024-01-01",
	}

	store.Add(ann)

	anns, _ := store.Load()
	if len(anns) != 1 {
		t.Errorf("Load() returned %d items, want 1", len(anns))
	}

	store.Delete("test1234")
	anns, _ = store.Load()
	if len(anns) != 0 {
		t.Errorf("After Delete(), got %d items, want 0", len(anns))
	}
}