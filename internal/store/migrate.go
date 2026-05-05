package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"remember/internal/model"
)

type jsonUsers struct {
	Users []model.User `json:"users"`
}

type jsonRemembers struct {
	Remembers []model.Anniversary `json:"remembers"`
}

// MigrateFromJSON 将 JSON 文件数据一次性导入 SQLite
func MigrateFromJSON(s *SQLiteStore, dataDir string) error {
	marker := filepath.Join(dataDir, ".migrated")
	if _, err := os.Stat(marker); err == nil {
		return nil
	}

	usersFile := filepath.Join(dataDir, "users.json")
	remembersFile := filepath.Join(dataDir, "remembers.json")

	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		if _, err := os.Stat(remembersFile); os.IsNotExist(err) {
			return nil
		}
	}

	var users []model.User
	if data, err := os.ReadFile(usersFile); err == nil {
		var ju jsonUsers
		if err := json.Unmarshal(data, &ju); err != nil {
			return fmt.Errorf("解析 users.json 失败: %w", err)
		}
		users = ju.Users
	}

	var anns []model.Anniversary
	if data, err := os.ReadFile(remembersFile); err == nil {
		var jr jsonRemembers
		if err := json.Unmarshal(data, &jr); err != nil {
			return fmt.Errorf("解析 remembers.json 失败: %w", err)
		}
		anns = jr.Remembers
	}

	if len(users) == 0 && len(anns) == 0 {
		return writeMarker(marker)
	}

	// 收集所有存在的用户 ID
	userIDs := make(map[string]bool)
	for _, u := range users {
		userIDs[u.ID] = true
	}

	// 为孤立纪念日创建占位用户
	for _, a := range anns {
		if a.UserID != "" && !userIDs[a.UserID] {
			users = append(users, model.User{
				ID:        a.UserID,
				Username:  "legacy_" + a.UserID[:4],
				Password:  "-",
				IsAdmin:   false,
				CreatedAt: "2026-01-01",
			})
			userIDs[a.UserID] = true
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始迁移事务失败: %w", err)
	}

	for _, u := range users {
		isAdmin := 0
		if u.IsAdmin {
			isAdmin = 1
		}
		_, err := tx.Exec(
			"INSERT OR IGNORE INTO users (id, username, password, is_admin, created_at) VALUES (?, ?, ?, ?, ?)",
			u.ID, u.Username, u.Password, isAdmin, u.CreatedAt,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("导入用户 %s 失败: %w", u.Username, err)
		}
	}

	for _, a := range anns {
		_, err := tx.Exec(
			"INSERT OR IGNORE INTO anniversaries (id, user_id, name, date, description, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			a.ID, a.UserID, a.Name, a.Date, a.Description, a.CreatedAt,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("导入纪念日 %s 失败: %w", a.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交迁移事务失败: %w", err)
	}

	fmt.Printf("数据迁移完成: %d 个用户, %d 条纪念日\n", len(users), len(anns))
	return writeMarker(marker)
}

func writeMarker(path string) error {
	return os.WriteFile(path, []byte("migrated"), 0644)
}
