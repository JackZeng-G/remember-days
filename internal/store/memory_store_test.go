package store

import (
	"testing"

	"remember/internal/model"
)

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
