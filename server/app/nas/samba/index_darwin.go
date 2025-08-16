//go:build darwin
// +build darwin

// Package samba 提供Samba服务管理相关的API接口 (macOS版本存根)
package samba

import (
	"github.com/gin-gonic/gin"
)

// AddRoutes macOS版本的空实现
// 在macOS系统上Samba功能不可用，因此不注册任何路由
func AddRoutes(parentGroup *gin.RouterGroup) {
	// macOS系统上不支持Samba，不注册任何路由
}
