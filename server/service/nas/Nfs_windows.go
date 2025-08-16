//go:build windows
// +build windows

// Package nas 提供网络存储相关的服务和模型 (Windows版本存根)
package nas

import (
	"server/core/db"

	"gorm.io/gorm"
)

// NfsShare Windows版本的NFS共享配置模型存根
type NfsShare struct {
	db.BaseModel[NfsShare] // 继承基础模型，提供通用字段和方法

	// 基本信息 (仅用于数据库兼容性，功能不可用)
	Name string `gorm:"comment:'共享名称'" json:"name"`
	Path string `gorm:"comment:'共享路径'" json:"path"`

	// 访问控制
	ClientIP   string `gorm:"comment:'客户端IP'" json:"clientIp"`
	Permission string `gorm:"comment:'读写权限'" json:"permission"`

	// 高级选项
	Sync         string `gorm:"comment:'同步模式'" json:"sync"`
	RootSquash   string `gorm:"comment:'Root权限映射'" json:"rootSquash"`
	SubtreeCheck string `gorm:"comment:'子树检查'" json:"subtreeCheck"`

	// 其他选项
	Options string `gorm:"comment:'其他选项'" json:"options"`
	Remark  string `gorm:"comment:'备注'" json:"remark"`
}

// TableName 返回NFS共享表名
func (NfsShare) TableName() string {
	return "nfs_shares"
}

// BeforeCreate Windows版本存根
func (n *NfsShare) BeforeCreate(tx *gorm.DB) (err error) {
	n.SupperBeforeCreate()
	if n.IsDisable == 0 {
		n.IsDisable = 0
	}
	return
}

// AfterFind Windows版本存根
func (n *NfsShare) AfterFind(tx *gorm.DB) (err error) {
	n.SupperAfterFind()
	return
}

// GetExportOptions Windows版本存根
func (n *NfsShare) GetExportOptions() string {
	return "unsupported_on_windows"
}

// GetExportLine Windows版本存根
func (n *NfsShare) GetExportLine() string {
	return "# NFS not supported on Windows"
}

// NfsService Windows版本的NFS服务管理存根
type NfsService struct {
	serviceName string
}
