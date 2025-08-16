//go:build windows
// +build windows

// Package nfs 提供NFS服务管理相关的API接口 (Windows版本存根)
package nfs

import (
	"github.com/gin-gonic/gin"
)

// AddRoutes Windows版本的空实现
// 在Windows系统上NFS功能不可用，因此不注册任何路由
func AddRoutes(parentGroup *gin.RouterGroup) {
	// Windows系统上不支持NFS，不注册任何路由
}
