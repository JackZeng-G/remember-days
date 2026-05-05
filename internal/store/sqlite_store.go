package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"remember/internal/model"
)

// SQLiteStore 基于 SQLite 的存储实现
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore 创建 SQLite 存储
func NewSQLiteStore(dataDir string) (*SQLiteStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dbPath := filepath.Join(dataDir, "remember.db")
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Close 关闭数据库连接
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ========== 数据库迁移 ==========

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER NOT NULL PRIMARY KEY
	)`,
	`CREATE TABLE IF NOT EXISTS users (
		id         TEXT NOT NULL PRIMARY KEY,
		username   TEXT NOT NULL UNIQUE,
		password   TEXT NOT NULL,
		is_admin   INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS anniversaries (
		id          TEXT NOT NULL PRIMARY KEY,
		user_id     TEXT NOT NULL REFERENCES users(id),
		name        TEXT NOT NULL,
		date        TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		created_at  TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_anniversaries_user_id ON anniversaries(user_id)`,
}

func runMigrations(db *sql.DB) error {
	for i, m := range migrations {
		version := i + 1
		var exists int
		err := db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err == nil {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("开始事务失败: %w", err)
		}

		if _, err := tx.Exec(m); err != nil {
			tx.Rollback()
			return fmt.Errorf("执行迁移 v%d 失败: %w", version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("记录迁移版本失败: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交迁移失败: %w", err)
		}
	}
	return nil
}

// ========== AnniversaryStore 接口实现 ==========

func (s *SQLiteStore) Load() ([]model.Anniversary, error) {
	rows, err := s.db.Query("SELECT id, user_id, name, date, description, created_at FROM anniversaries")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Anniversary
	for rows.Next() {
		var a model.Anniversary
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Date, &a.Description, &a.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (s *SQLiteStore) Save(anniversaries []model.Anniversary) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM anniversaries"); err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO anniversaries (id, user_id, name, date, description, created_at) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, a := range anniversaries {
		if _, err := stmt.Exec(a.ID, a.UserID, a.Name, a.Date, a.Description, a.CreatedAt); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) Add(ann model.Anniversary) error {
	_, err := s.db.Exec(
		"INSERT INTO anniversaries (id, user_id, name, date, description, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		ann.ID, ann.UserID, ann.Name, ann.Date, ann.Description, ann.CreatedAt,
	)
	return err
}

func (s *SQLiteStore) Delete(id string) error {
	res, err := s.db.Exec("DELETE FROM anniversaries WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("纪念日不存在: %s", id)
	}
	return nil
}

// ========== UserStore 接口实现 ==========

func (s *SQLiteStore) LoadUsers() ([]model.User, error) {
	rows, err := s.db.Query("SELECT id, username, password, is_admin, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.User
	for rows.Next() {
		var u model.User
		var isAdmin int
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &isAdmin, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.IsAdmin = isAdmin == 1
		result = append(result, u)
	}
	return result, rows.Err()
}

func (s *SQLiteStore) SaveUsers(users []model.User) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM users"); err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO users (id, username, password, is_admin, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, u := range users {
		isAdmin := 0
		if u.IsAdmin {
			isAdmin = 1
		}
		if _, err := stmt.Exec(u.ID, u.Username, u.Password, isAdmin, u.CreatedAt); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) AddUser(user model.User) error {
	isAdmin := 0
	if user.IsAdmin {
		isAdmin = 1
	}
	_, err := s.db.Exec(
		"INSERT INTO users (id, username, password, is_admin, created_at) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Username, user.Password, isAdmin, user.CreatedAt,
	)
	return err
}

func (s *SQLiteStore) GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	var isAdmin int
	err := s.db.QueryRow(
		"SELECT id, username, password, is_admin, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &isAdmin, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.IsAdmin = isAdmin == 1
	return &u, nil
}

func (s *SQLiteStore) GetUserByID(id string) (*model.User, error) {
	var u model.User
	var isAdmin int
	err := s.db.QueryRow(
		"SELECT id, username, password, is_admin, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &isAdmin, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.IsAdmin = isAdmin == 1
	return &u, nil
}

func (s *SQLiteStore) DeleteUser(id string) error {
	res, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("用户不存在: %s", id)
	}
	return nil
}

func (s *SQLiteStore) UserCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
