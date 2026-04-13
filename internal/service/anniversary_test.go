package service

import (
	"testing"

	"anniversary/internal/store"
)

func TestAnniversaryService_Create(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	ann, err := svc.Create("Test", "2024-01-01", "Description")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if ann.Name != "Test" {
		t.Errorf("Name = %s, want Test", ann.Name)
	}
	if ann.Date != "2024-01-01" {
		t.Errorf("Date = %s, want 2024-01-01", ann.Date)
	}
	if len(ann.ID) != 8 {
		t.Errorf("ID length = %d, want 8", len(ann.ID))
	}
}

func TestAnniversaryService_Create_InvalidName(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Create("", "2024-01-01", "")
	if err != ErrEmptyName {
		t.Errorf("Create() error = %v, want ErrEmptyName", err)
	}
}

func TestAnniversaryService_Create_InvalidDate(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Create("Test", "invalid-date", "")
	if err != ErrInvalidDate {
		t.Errorf("Create() error = %v, want ErrInvalidDate", err)
	}
}

func TestAnniversaryService_List(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	svc.Create("First", "2024-01-01", "")
	svc.Create("Second", "2024-06-15", "")

	views, err := svc.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(views) != 2 {
		t.Errorf("List() returned %d items, want 2", len(views))
	}
}

func TestAnniversaryService_Get_NotFound(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Get("nonexist")
	if err != ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestAnniversaryService_Get_InvalidID(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Get("bad-id")
	if err != ErrInvalidID {
		t.Errorf("Get() error = %v, want ErrInvalidID", err)
	}
}

func TestAnniversaryService_Update(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	ann, _ := svc.Create("Original", "2024-01-01", "Original desc")

	err := svc.Update(ann.ID, "Updated", "2024-12-31", "Updated desc")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, _ := svc.Get(ann.ID)
	if updated.Name != "Updated" {
		t.Errorf("Name = %s, want Updated", updated.Name)
	}
}

func TestAnniversaryService_Delete(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	ann, _ := svc.Create("ToDelete", "2024-01-01", "")

	err := svc.Delete(ann.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = svc.Get(ann.ID)
	if err != ErrNotFound {
		t.Errorf("After Delete(), Get() error = %v, want ErrNotFound", err)
	}
}