package store

import "remember/internal/model"

// UserStore 用户存储接口
type UserStore interface {
	LoadUsers() ([]model.User, error)
	SaveUsers(users []model.User) error
	AddUser(user model.User) error
	GetUserByUsername(username string) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	DeleteUser(id string) error
	UserCount() (int, error)
}
