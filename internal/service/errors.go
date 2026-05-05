package service

import "errors"

var (
	ErrInvalidID    = errors.New("无效的纪念日ID")
	ErrNotFound     = errors.New("纪念日不存在")
	ErrInvalidDate  = errors.New("日期格式错误")
	ErrEmptyName    = errors.New("名称不能为空")
	ErrInvalidInput = errors.New("输入参数无效")

	ErrUserNotFound       = errors.New("用户不存在")
	ErrUsernameTaken      = errors.New("用户名已被使用")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrInvalidUsername    = errors.New("用户名格式无效")
	ErrPasswordTooShort   = errors.New("密码长度不能少于6位")
	ErrCannotDeleteSelf   = errors.New("不能删除自己的账号")
	ErrCannotDeleteAdmin  = errors.New("不能删除管理员账号")
)