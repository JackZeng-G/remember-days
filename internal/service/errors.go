package service

import "errors"

var (
	ErrInvalidID    = errors.New("无效的纪念日ID")
	ErrNotFound     = errors.New("纪念日不存在")
	ErrInvalidDate  = errors.New("日期格式错误")
	ErrEmptyName    = errors.New("名称不能为空")
	ErrInvalidInput = errors.New("输入参数无效")
)