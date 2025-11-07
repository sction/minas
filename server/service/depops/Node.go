package depops

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"server/core/app/request"
	"server/core/db"
	"server/utils/global"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// Node 节点模型
type Node struct {
	db.BaseModel[Node]
	Name           string     `gorm:"size:100;not null;comment:'节点名称'" json:"name" mapstructure:"name"`
	IPAddress      string     `gorm:"size:15;not null;index;comment:'IP地址'" json:"ip_address" mapstructure:"ip_address"`
	SSHPort        int        `gorm:"default:22;comment:'SSH端口'" json:"ssh_port" mapstructure:"ssh_port"`
	Username       string     `gorm:"size:50;not null;comment:'SSH用户名'" json:"username" mapstructure:"username"`
	Password       string     `gorm:"size:255;comment:'SSH密码'" json:"password,omitempty" mapstructure:"password"`
	PrivateKey     string     `gorm:"type:text;comment:'SSH私钥'" json:"private_key,omitempty" mapstructure:"private_key"`
	PrivateKeyPass string     `gorm:"size:255;comment:'私钥密码'" json:"private_key_pass,omitempty" mapstructure:"private_key_pass"`
	Tags           string     `gorm:"size:500;comment:'标签(JSON数组)'" json:"tags" mapstructure:"tags"`
	OSType         string     `gorm:"size:50;comment:'操作系统类型'" json:"os_type" mapstructure:"os_type"`
	OSVersion      string     `gorm:"size:100;comment:'操作系统版本'" json:"os_version" mapstructure:"os_version"`
	Architecture   string     `gorm:"size:50;comment:'系统架构(amd64/arm64/386等)'" json:"architecture" mapstructure:"architecture"`
	PackageManager string     `gorm:"size:50;comment:'包管理器类型(deb/rpm/apk等)'" json:"package_manager" mapstructure:"package_manager"`
	Hostname       string     `gorm:"size:100;comment:'主机名'" json:"hostname" mapstructure:"hostname"`
	Timezone       string     `gorm:"size:50;comment:'时区'" json:"timezone" mapstructure:"timezone"`
	LastConnected  *time.Time `gorm:"comment:'最后连接时间'" json:"last_connected" mapstructure:"last_connected"`
	Status         int        `gorm:"default:0;comment:'状态:0-未知,1-在线,2-离线'" json:"status" mapstructure:"status"`
	Remark         string     `gorm:"size:500;comment:'备注'" json:"remark" mapstructure:"remark"`
}

// TableName 指定表名
func (Node) TableName() string {
	return "depops_nodes"
}

// BeforeSave 保存前的钩子函数
func (n *Node) BeforeSave(tx *gorm.DB) error {
	// 验证IP地址格式
	if net.ParseIP(n.IPAddress) == nil {
		return errors.New("无效的IP地址")
	}

	// 验证SSH端口范围
	if n.SSHPort < 1 || n.SSHPort > 65535 {
		return errors.New("SSH端口必须在1-65535范围内")
	}

	// 验证认证方式
	if n.Password == "" && n.PrivateKey == "" {
		return errors.New("必须提供密码或私钥")
	}

	return nil
}

// GetTags 获取标签列表
func (n *Node) GetTags() []string {
	var tags []string
	if n.Tags != "" {
		json.Unmarshal([]byte(n.Tags), &tags)
	}
	return tags
}

// SetTags 设置标签列表
func (n *Node) SetTags(tags []string) {
	if tagBytes, err := json.Marshal(tags); err == nil {
		n.Tags = string(tagBytes)
	}
}

// GetSSHClient 获取SSH客户端连接
func (n *Node) GetSSHClient() (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            n.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// 根据认证方式配置
	if n.PrivateKey != "" {
		// 使用私钥认证
		var signer ssh.Signer
		var err error

		if n.PrivateKeyPass != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(n.PrivateKey), []byte(n.PrivateKeyPass))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(n.PrivateKey))
		}

		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %v", err)
		}

		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		// 使用密码认证
		config.Auth = []ssh.AuthMethod{ssh.Password(n.Password)}
	}

	// 建立SSH连接
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", n.IPAddress, n.SSHPort), config)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %v", err)
	}

	return client, nil
}

// TestConnectionOnly 仅测试连接，不更新数据库
func (n *Node) TestConnectionOnly() error {
	client, err := n.GetSSHClient()
	if err != nil {
		return err
	}
	defer client.Close()

	return nil
}

// TestConnection 测试节点连接
func (n *Node) TestConnection() error {
	client, err := n.GetSSHClient()
	if err != nil {
		// 连接失败，更新状态为离线
		n.Status = 2 // 离线
		global.DB.Model(n).Updates(map[string]interface{}{
			"status": 2,
		})
		return err
	}
	defer client.Close()

	// 更新最后连接时间和状态
	now := time.Now()
	n.LastConnected = &now
	n.Status = 1 // 在线
	global.DB.Model(n).Updates(map[string]interface{}{
		"last_connected": now,
		"status":         1,
	})

	return nil
}

// ExecuteCommand 执行SSH命令
func (n *Node) ExecuteCommand(command string) (string, error) {
	client, err := n.GetSSHClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}

	return string(output), nil
}

// GetSystemInfo 获取系统信息
func (n *Node) GetSystemInfo() error {
	client, err := n.GetSSHClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// 获取操作系统信息
	osInfo, err := n.executeSSHCommand(client, "cat /etc/os-release 2>/dev/null | grep -E '^ID=|^VERSION_ID=' | cut -d= -f2 | tr -d '\"' | head -2")
	if err == nil && osInfo != "" {
		lines := strings.Split(strings.TrimSpace(osInfo), "\n")
		if len(lines) >= 1 {
			n.OSType = strings.TrimSpace(lines[0])
		}
		if len(lines) >= 2 {
			n.OSVersion = strings.TrimSpace(lines[1])
		}
	} else {
		// 如果 /etc/os-release 不存在，尝试其他方法
		// 尝试获取系统类型
		unameInfo, err := n.executeSSHCommand(client, "uname -s")
		if err == nil {
			n.OSType = strings.TrimSpace(unameInfo)
		}

		// 尝试获取系统版本
		versionInfo, err := n.executeSSHCommand(client, "uname -r")
		if err == nil {
			n.OSVersion = strings.TrimSpace(versionInfo)
		}
	}

	// 获取系统架构
	archInfo, err := n.executeSSHCommand(client, "uname -m")
	if err == nil {
		arch := strings.TrimSpace(archInfo)
		// 将常见的架构名称标准化
		switch arch {
		case "x86_64":
			n.Architecture = "amd64"
		case "aarch64":
			n.Architecture = "arm64"
		case "i386", "i686":
			n.Architecture = "386"
		case "armv7l":
			n.Architecture = "arm"
		default:
			n.Architecture = arch
		}
	}

	// 检测包管理器类型
	n.detectPackageManager(client)

	// 获取主机名
	hostname, err := n.executeSSHCommand(client, "hostname")
	if err == nil {
		n.Hostname = strings.TrimSpace(hostname)
	}

	// 获取时区
	timezone, err := n.executeSSHCommand(client, "timedatectl show --value --property=Timezone 2>/dev/null || date +%Z")
	if err == nil {
		n.Timezone = strings.TrimSpace(timezone)
	}

	// 更新连接状态和时间
	now := time.Now()
	n.LastConnected = &now
	n.Status = 1 // 在线

	// 更新到数据库
	return global.DB.Model(n).Updates(map[string]interface{}{
		"os_type":         n.OSType,
		"os_version":      n.OSVersion,
		"architecture":    n.Architecture,
		"package_manager": n.PackageManager,
		"hostname":        n.Hostname,
		"timezone":        n.Timezone,
		"last_connected":  now,
		"status":          1,
	}).Error
}

// detectPackageManager 检测包管理器类型
func (n *Node) detectPackageManager(client *ssh.Client) {
	// 检测各种包管理器
	packageManagers := []struct {
		command string
		name    string
	}{
		{"which apt 2>/dev/null", "deb"},        // Debian/Ubuntu
		{"which yum 2>/dev/null", "rpm"},        // CentOS/RHEL (old)
		{"which dnf 2>/dev/null", "rpm"},        // Fedora/CentOS/RHEL (new)
		{"which zypper 2>/dev/null", "rpm"},     // openSUSE
		{"which apk 2>/dev/null", "apk"},        // Alpine Linux
		{"which pacman 2>/dev/null", "pacman"},  // Arch Linux
		{"which emerge 2>/dev/null", "portage"}, // Gentoo
		{"which pkg 2>/dev/null", "pkg"},        // FreeBSD
		{"which brew 2>/dev/null", "brew"},      // macOS Homebrew
		{"which port 2>/dev/null", "macports"},  // macOS MacPorts
	}

	for _, pm := range packageManagers {
		output, err := n.executeSSHCommand(client, pm.command)
		if err == nil && strings.TrimSpace(output) != "" {
			n.PackageManager = pm.name
			break
		}
	}

	// 如果都没检测到，尝试根据操作系统类型推断
	if n.PackageManager == "" {
		switch strings.ToLower(n.OSType) {
		case "ubuntu", "debian":
			n.PackageManager = "deb"
		case "centos", "rhel", "fedora", "rocky", "almalinux":
			n.PackageManager = "rpm"
		case "alpine":
			n.PackageManager = "apk"
		case "arch", "manjaro":
			n.PackageManager = "pacman"
		case "gentoo":
			n.PackageManager = "portage"
		case "freebsd":
			n.PackageManager = "pkg"
		case "darwin":
			n.PackageManager = "brew"
		default:
			n.PackageManager = "unknown"
		}
	}
}

// executeSSHCommand 执行SSH命令的辅助方法
func (n *Node) executeSSHCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}

// GetNodesByTags 根据标签获取节点列表
func GetNodesByTags(tags []string) ([]Node, error) {
	var nodes []Node
	db := global.DB.Model(&Node{})

	for _, tag := range tags {
		db = db.Where("tags LIKE ?", "%"+tag+"%")
	}

	err := db.Find(&nodes).Error
	return nodes, err
}

// GetAllTags 获取所有节点的标签列表
func GetAllTags() ([]string, error) {
	var nodes []Node
	var allTags []string
	tagMap := make(map[string]string)

	if err := global.DB.Select("tags").Find(&nodes).Error; err != nil {
		return nil, err
	}

	for _, node := range nodes {
		tags := node.GetTags()
		for _, tag := range tags {
			if tag != "" {
				tagMap[tag] = tag
			}
		}
	}

	for _, tag := range tagMap {
		allTags = append(allTags, tag)
	}

	return allTags, nil
}

// List 实现节点列表查询，支持搜索和标签过滤
func (n Node) List(query request.PageQuery) ([]Node, int64, error) {
	var nodes []Node
	var total int64

	db := global.DB.Model(&Node{})

	// 处理查询过滤条件
	for _, filter := range query.Filters {
		db = n.applyFilter(db, filter)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页处理
	if query.Size > 0 {
		offset := (query.Page - 1) * query.Size
		db = db.Offset(offset).Limit(query.Size)
	}

	// 排序处理
	if len(query.Sorts) > 0 {
		for _, sort := range query.Sorts {
			db = db.Order(sort)
		}
	} else {
		db = db.Order("id DESC")
	}

	// 查询数据
	if err := db.Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

// applyFilter 应用查询过滤条件
func (n Node) applyFilter(db *gorm.DB, filter request.QueryFilter) *gorm.DB {
	if len(filter.Filters) > 0 {
		// 处理嵌套过滤条件（OR组合）
		var orConditions []string
		var orValues []interface{}

		for _, subFilter := range filter.Filters {
			if subFilter.Column != "" && subFilter.Operator != "" {
				orConditions = append(orConditions, subFilter.Column+subFilter.Operator)
				orValues = append(orValues, subFilter.Value)
			}
		}

		if len(orConditions) > 0 {
			// 构建OR查询
			orClause := strings.Join(orConditions, " OR ")
			return db.Where(orClause, orValues...)
		}
		return db
	} else {
		// 处理简单过滤条件
		return n.applySimpleFilter(db, filter)
	}
}

// applySimpleFilter 应用简单过滤条件
func (n Node) applySimpleFilter(db *gorm.DB, filter request.QueryFilter) *gorm.DB {
	if filter.Or {
		return db.Or(filter.Column+filter.Operator, filter.Value)
	}
	return db.Where(filter.Column+filter.Operator, filter.Value)
}
