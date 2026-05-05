package service

import (
	"testing"

	"remember/internal/store"
)

const testUserID = "test0001"

func TestAnniversaryService_Create(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	ann, err := svc.Create(testUserID, "Test", "2024-01-01", "Description")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if ann.Name != "Test" {
		t.Errorf("Name = %s, want Test", ann.Name)
	}
	if ann.UserID != testUserID {
		t.Errorf("UserID = %s, want %s", ann.UserID, testUserID)
	}
}

func TestAnniversaryService_Create_InvalidName(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Create(testUserID, "", "2024-01-01", "")
	if err != ErrEmptyName {
		t.Errorf("Create() error = %v, want ErrEmptyName", err)
	}
}

func TestAnniversaryService_Create_InvalidDate(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Create(testUserID, "Test", "invalid-date", "")
	if err != ErrInvalidDate {
		t.Errorf("Create() error = %v, want ErrInvalidDate", err)
	}
}

func TestAnniversaryService_List(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	svc.Create(testUserID, "First", "2024-01-01", "")
	svc.Create("other000", "Second", "2024-06-15", "")

	views, err := svc.List(testUserID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(views) != 1 {
		t.Errorf("List() returned %d items, want 1", len(views))
	}
}

func TestAnniversaryService_Get_NotFound(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Get(testUserID, "nonexist")
	if err != ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestAnniversaryService_Get_InvalidID(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	_, err := svc.Get(testUserID, "bad-id")
	if err != ErrInvalidID {
		t.Errorf("Get() error = %v, want ErrInvalidID", err)
	}
}

func TestAnniversaryService_Update(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	ann, _ := svc.Create(testUserID, "Original", "2024-01-01", "Original desc")

	err := svc.Update(testUserID, ann.ID, "Updated", "2024-12-31", "Updated desc")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, _ := svc.Get(testUserID, ann.ID)
	if updated.Name != "Updated" {
		t.Errorf("Name = %s, want Updated", updated.Name)
	}
}

func TestAnniversaryService_Delete(t *testing.T) {
	memStore := store.NewMemoryStore()
	svc := New(memStore)

	ann, _ := svc.Create(testUserID, "ToDelete", "2024-01-01", "")

	err := svc.Delete(testUserID, ann.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = svc.Get(testUserID, ann.ID)
	if err != ErrNotFound {
		t.Errorf("After Delete(), Get() error = %v, want ErrNotFound", err)
	}
}
