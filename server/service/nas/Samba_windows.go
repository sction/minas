//go:build windows
// +build windows

// Package nas 提供网络存储相关的服务和模型 (Windows版本存根)
package nas

import (
	"server/core/db"
)

// SambaShare Windows版本的Samba共享配置模型存根
type SambaShare struct {
	db.BaseModel[SambaShare] // 继承基础模型，提供通用字段和方法

	// 基本信息 (仅用于数据库兼容性，功能不可用)
	Name string `gorm:"comment:'共享名称'" json:"name"`
	Path string `gorm:"comment:'共享路径'" json:"path"`

	// 共享设置
	Comment    string `gorm:"comment:'共享描述'" json:"comment"`
	Browseable bool   `gorm:"comment:'是否可浏览'" json:"browseable"`
	ReadOnly   bool   `gorm:"comment:'是否只读'" json:"readOnly"`
	GuestOk    bool   `gorm:"comment:'是否允许来宾访问'" json:"guestOk"`

	// 访问控制
	ValidUsers string `gorm:"comment:'有效用户列表'" json:"validUsers"`
	WriteList  string `gorm:"comment:'写权限用户列表'" json:"writeList"`
	ReadList   string `gorm:"comment:'读权限用户列表'" json:"readList"`

	// 权限设置
	CreateMask    string `gorm:"comment:'创建文件权限掩码'" json:"createMask"`
	DirectoryMask string `gorm:"comment:'创建目录权限掩码'" json:"directoryMask"`
	ForceUser     string `gorm:"comment:'强制用户'" json:"forceUser"`
	ForceGroup    string `gorm:"comment:'强制组'" json:"forceGroup"`

	// 其他配置
	AvailableSpace string `gorm:"comment:'可用空间限制'" json:"availableSpace"`
	Options        string `gorm:"comment:'其他选项'" json:"options"`

	// 备注信息
	Remark string `gorm:"comment:'备注'" json:"remark"`
}

// TableName 返回Samba共享表名
func (SambaShare) TableName() string {
	return "samba_shares"
}

// GetSmbConfigSection Windows版本存根
func (s *SambaShare) GetSmbConfigSection() string {
	return "# Samba not supported on Windows"
}

// SambaService Windows版本的Samba服务管理存根
type SambaService struct {
	serviceName string
}

// NewSambaService Windows版本存根
func NewSambaService() *SambaService {
	return &SambaService{
		serviceName: "unsupported_on_windows",
	}
}
