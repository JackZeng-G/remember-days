package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"remember/internal/model"
	"remember/internal/store"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\p{Han}]{2,20}$`)

// UserService 用户服务接口
type UserService interface {
	Register(username, password string) (*model.User, error)
	Login(username, password string) (*model.User, error)
	GetByID(id string) (*model.User, error)
	ListAll() ([]model.User, error)
	DeleteUser(targetID, currentID string) error
}

type userService struct {
	userStore store.UserStore
	annStore  store.AnniversaryStore
}

// NewUserService 创建用户服务
func NewUserService(us store.UserStore, as store.AnniversaryStore) UserService {
	return &userService{userStore: us, annStore: as}
}

func (s *userService) Register(username, password string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if !usernameRegex.MatchString(username) {
		return nil, ErrInvalidUsername
	}
	if len(password) < 6 {
		return nil, ErrPasswordTooShort
	}

	existing, err := s.userStore.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	count, err := s.userStore.UserCount()
	if err != nil {
		return nil, fmt.Errorf("查询用户数失败: %w", err)
	}

	user := model.User{
		ID:        uuid.New().String()[:8],
		Username:  username,
		Password:  string(hash),
		IsAdmin:   count == 0,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.userStore.AddUser(user); err != nil {
		return nil, fmt.Errorf("保存用户失败: %w", err)
	}

	// 首个用户自动认领旧数据
	if count == 0 {
		s.migrateLegacyData(user.ID)
	}

	return &user, nil
}

func (s *userService) Login(username, password string) (*model.User, error) {
	user, err := s.userStore.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *userService) GetByID(id string) (*model.User, error) {
	user, err := s.userStore.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) ListAll() ([]model.User, error) {
	return s.userStore.LoadUsers()
}

func (s *userService) DeleteUser(targetID, currentID string) error {
	if targetID == currentID {
		return ErrCannotDeleteSelf
	}

	target, err := s.userStore.GetUserByID(targetID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	if target.IsAdmin {
		return ErrCannotDeleteAdmin
	}

	return s.userStore.DeleteUser(targetID)
}

func (s *userService) migrateLegacyData(userID string) {
	anns, err := s.annStore.Load()
	if err != nil {
		return
	}
	changed := false
	for i := range anns {
		if anns[i].UserID == "" {
			anns[i].UserID = userID
			changed = true
		}
	}
	if changed {
		s.annStore.Save(anns)
	}
}
