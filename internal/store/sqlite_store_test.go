package store

import (
	"os"
	"path/filepath"
	"testing"

	"remember/internal/model"
)

func newTestSQLiteStore(t *testing.T) *SQLiteStore {
	t.Helper()
	tmpDir := t.TempDir()
	s, err := NewSQLiteStore(tmpDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func newTestSQLiteStoreWithUser(t *testing.T) *SQLiteStore {
	t.Helper()
	s := newTestSQLiteStore(t)
	s.AddUser(model.User{ID: "u1", Username: "testuser", Password: "h", CreatedAt: "2026-01-01"})
	return s
}

// ========== AnniversaryStore 测试 ==========

func TestSQLiteStore_Load_Empty(t *testing.T) {
	s := newTestSQLiteStoreWithUser(t)
	anns, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(anns) != 0 {
		t.Fatalf("Load() empty DB = %d, want 0", len(anns))
	}
}

func TestSQLiteStore_AddAndLoad(t *testing.T) {
	s := newTestSQLiteStoreWithUser(t)
	ann := model.Anniversary{
		ID: "test1", UserID: "u1", Name: "测试", Date: "2020-01-01", Description: "desc", CreatedAt: "2026-01-01",
	}
	if err := s.Add(ann); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	anns, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(anns) != 1 {
		t.Fatalf("Load() = %d, want 1", len(anns))
	}
	if anns[0].ID != "test1" || anns[0].Name != "测试" {
		t.Fatalf("Load() = %+v, want id=test1 name=测试", anns[0])
	}
}

func TestSQLiteStore_Delete(t *testing.T) {
	s := newTestSQLiteStoreWithUser(t)
	s.Add(model.Anniversary{ID: "d1", UserID: "u1", Name: "n", Date: "2020-01-01", CreatedAt: "2026-01-01"})

	if err := s.Delete("d1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	anns, _ := s.Load()
	if len(anns) != 0 {
		t.Fatalf("after Delete, Load() = %d, want 0", len(anns))
	}
}

func TestSQLiteStore_Delete_NotFound(t *testing.T) {
	s := newTestSQLiteStoreWithUser(t)
	err := s.Delete("nonexist")
	if err == nil {
		t.Fatal("Delete(nonexist) should error")
	}
}

func TestSQLiteStore_Save(t *testing.T) {
	s := newTestSQLiteStoreWithUser(t)
	s.Add(model.Anniversary{ID: "a", UserID: "u1", Name: "old", Date: "2020-01-01", CreatedAt: "2026-01-01"})
	s.Add(model.Anniversary{ID: "b", UserID: "u1", Name: "old2", Date: "2020-01-02", CreatedAt: "2026-01-01"})

	newList := []model.Anniversary{
		{ID: "c", UserID: "u1", Name: "new", Date: "2021-01-01", CreatedAt: "2026-01-02"},
	}
	if err := s.Save(newList); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	anns, _ := s.Load()
	if len(anns) != 1 || anns[0].ID != "c" {
		t.Fatalf("after Save, Load() = %+v, want [{ID:c}]", anns)
	}
}

// ========== UserStore 测试 ==========

func TestSQLiteStore_AddUserAndGetByUsername(t *testing.T) {
	s := newTestSQLiteStore(t)
	u := model.User{ID: "u1", Username: "alice", Password: "hash", IsAdmin: true, CreatedAt: "2026-01-01"}
	if err := s.AddUser(u); err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	got, err := s.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}
	if got == nil || got.ID != "u1" || !got.IsAdmin {
		t.Fatalf("GetUserByUsername() = %+v, want id=u1 admin=true", got)
	}
}

func TestSQLiteStore_GetUserByID(t *testing.T) {
	s := newTestSQLiteStore(t)
	s.AddUser(model.User{ID: "u2", Username: "bob", Password: "hash", CreatedAt: "2026-01-01"})

	got, err := s.GetUserByID("u2")
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if got == nil || got.Username != "bob" {
		t.Fatalf("GetUserByID() = %+v", got)
	}
}

func TestSQLiteStore_GetUserByUsername_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)
	got, err := s.GetUserByUsername("nobody")
	if err != nil {
		t.Fatalf("GetUserByUsername(notfound) error = %v", err)
	}
	if got != nil {
		t.Fatalf("GetUserByUsername(notfound) = %+v, want nil", got)
	}
}

func TestSQLiteStore_DeleteUser(t *testing.T) {
	s := newTestSQLiteStore(t)
	s.AddUser(model.User{ID: "u3", Username: "charlie", Password: "h", CreatedAt: "2026-01-01"})

	if err := s.DeleteUser("u3"); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	got, _ := s.GetUserByID("u3")
	if got != nil {
		t.Fatal("after DeleteUser, user still exists")
	}
}

func TestSQLiteStore_DeleteUser_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)
	err := s.DeleteUser("nonexist")
	if err == nil {
		t.Fatal("DeleteUser(nonexist) should error")
	}
}

func TestSQLiteStore_UserCount(t *testing.T) {
	s := newTestSQLiteStore(t)
	if count, _ := s.UserCount(); count != 0 {
		t.Fatalf("UserCount() = %d, want 0", count)
	}
	s.AddUser(model.User{ID: "u1", Username: "a", Password: "h", CreatedAt: "2026-01-01"})
	s.AddUser(model.User{ID: "u2", Username: "b", Password: "h", CreatedAt: "2026-01-01"})
	if count, _ := s.UserCount(); count != 2 {
		t.Fatalf("UserCount() = %d, want 2", count)
	}
}

func TestSQLiteStore_UniqueUsername(t *testing.T) {
	s := newTestSQLiteStore(t)
	s.AddUser(model.User{ID: "u1", Username: "dup", Password: "h", CreatedAt: "2026-01-01"})
	err := s.AddUser(model.User{ID: "u2", Username: "dup", Password: "h", CreatedAt: "2026-01-01"})
	if err == nil {
		t.Fatal("duplicate username should error")
	}
}

// ========== 迁移测试 ==========

func TestMigrateFromJSON(t *testing.T) {
	tmpDir := t.TempDir()

	usersJSON := `{"users":[{"id":"abc","username":"test","password":"hash","is_admin":true,"created_at":"2026-01-01"}]}`
	remembersJSON := `{"remembers":[{"id":"r1","user_id":"abc","name":"生日","date":"2020-06-15","description":"","created_at":"2026-01-01"}]}`

	os.WriteFile(filepath.Join(tmpDir, "users.json"), []byte(usersJSON), 0644)
	os.WriteFile(filepath.Join(tmpDir, "remembers.json"), []byte(remembersJSON), 0644)

	s, err := NewSQLiteStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := MigrateFromJSON(s, tmpDir); err != nil {
		t.Fatalf("MigrateFromJSON() error = %v", err)
	}

	users, _ := s.LoadUsers()
	if len(users) != 1 || users[0].Username != "test" {
		t.Fatalf("after migration, users = %+v", users)
	}

	anns, _ := s.Load()
	if len(anns) != 1 || anns[0].Name != "生日" {
		t.Fatalf("after migration, anns = %+v", anns)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ".migrated")); os.IsNotExist(err) {
		t.Fatal(".migrated marker not created")
	}
}

func TestMigrateFromJSON_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "users.json"), []byte(`{"users":[{"id":"u1","username":"a","password":"h","created_at":"2026-01-01"}]}`), 0644)

	s, _ := NewSQLiteStore(tmpDir)
	defer s.Close()

	MigrateFromJSON(s, tmpDir)
	MigrateFromJSON(s, tmpDir)

	users, _ := s.LoadUsers()
	if len(users) != 1 {
		t.Fatalf("idempotent migration: users = %d, want 1", len(users))
	}
}
