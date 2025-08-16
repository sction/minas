//go:build linux
// +build linux

// Package nas 提供网络存储相关的服务和模型
package nas

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"server/core/db"
	"server/utils/global"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// SambaShare 定义Samba共享配置模型
// 用于存储Samba共享信息，包括路径、权限、访问控制等
type SambaShare struct {
	db.BaseModel[SambaShare] // 继承基础模型，提供通用字段和方法

	// 基本信息
	Name string `gorm:"comment:'共享名称'" json:"name"` // Samba共享名称
	Path string `gorm:"comment:'共享路径'" json:"path"` // Samba共享目录路径

	// 共享设置
	Comment    string `gorm:"comment:'共享描述'" json:"comment"`     // 共享描述
	Browseable bool   `gorm:"comment:'是否可浏览'" json:"browseable"` // 是否可浏览
	ReadOnly   bool   `gorm:"comment:'是否只读'" json:"readOnly"`    // 是否只读
	GuestOk    bool   `gorm:"comment:'是否允许来宾访问'" json:"guestOk"` // 是否允许来宾访问

	// 访问控制
	ValidUsers string `gorm:"comment:'有效用户列表'" json:"validUsers"` // 有效用户列表，逗号分隔
	WriteList  string `gorm:"comment:'写权限用户列表'" json:"writeList"` // 写权限用户列表，逗号分隔
	ReadList   string `gorm:"comment:'读权限用户列表'" json:"readList"`  // 读权限用户列表，逗号分隔

	// 权限设置
	CreateMask    string `gorm:"comment:'创建文件权限掩码'" json:"createMask"`    // 创建文件权限掩码
	DirectoryMask string `gorm:"comment:'创建目录权限掩码'" json:"directoryMask"` // 创建目录权限掩码
	ForceUser     string `gorm:"comment:'强制用户'" json:"forceUser"`         // 强制用户
	ForceGroup    string `gorm:"comment:'强制组'" json:"forceGroup"`         // 强制组

	// 其他配置
	AvailableSpace string `gorm:"comment:'可用空间限制'" json:"availableSpace"` // 可用空间限制
	Options        string `gorm:"comment:'其他选项'" json:"options"`          // 其他自定义Samba选项

	// 备注信息
	Remark string `gorm:"comment:'备注'" json:"remark"` // 备注信息
}

// TableName 返回Samba共享表名
// 实现gorm的Tabler接口
func (SambaShare) TableName() string {
	return "samba_shares"
}

// GetSmbConfigSection 获取SMB配置段
func (s *SambaShare) GetSmbConfigSection() string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("[%s]\n", s.Name))
	config.WriteString(fmt.Sprintf("path = %s\n", s.Path))

	if s.Comment != "" {
		config.WriteString(fmt.Sprintf("comment = %s\n", s.Comment))
	}

	config.WriteString(fmt.Sprintf("browseable = %s\n", boolToYesNo(s.Browseable)))
	config.WriteString(fmt.Sprintf("read only = %s\n", boolToYesNo(s.ReadOnly)))
	config.WriteString(fmt.Sprintf("guest ok = %s\n", boolToYesNo(s.GuestOk)))

	if s.ValidUsers != "" {
		config.WriteString(fmt.Sprintf("valid users = %s\n", s.ValidUsers))
	}

	if s.WriteList != "" {
		config.WriteString(fmt.Sprintf("write list = %s\n", s.WriteList))
	}

	if s.ReadList != "" {
		config.WriteString(fmt.Sprintf("read list = %s\n", s.ReadList))
	}

	if s.CreateMask != "" {
		config.WriteString(fmt.Sprintf("create mask = %s\n", s.CreateMask))
	}

	if s.DirectoryMask != "" {
		config.WriteString(fmt.Sprintf("directory mask = %s\n", s.DirectoryMask))
	}

	if s.ForceUser != "" {
		config.WriteString(fmt.Sprintf("force user = %s\n", s.ForceUser))
	}

	if s.ForceGroup != "" {
		config.WriteString(fmt.Sprintf("force group = %s\n", s.ForceGroup))
	}

	// 添加其他选项
	if s.Options != "" {
		lines := strings.Split(s.Options, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				config.WriteString(fmt.Sprintf("%s\n", trimmedLine))
			}
		}
	}

	return config.String()
}

// boolToYesNo 将布尔值转换为yes/no字符串
func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// SambaService Samba服务管理
type SambaService struct {
	// 缓存检测到的Samba服务名称
	serviceName string
}

// NewSambaService 创建新的Samba服务实例
func NewSambaService() *SambaService {
	return &SambaService{
		serviceName: "smbd", // 默认服务名称
	}
}

// detectSambaService 检测当前系统的Samba服务名称
func (s *SambaService) detectSambaService() string {
	if s.serviceName != "" && s.serviceName != "smbd" {
		return s.serviceName
	}

	// 常见的Samba服务名称
	serviceNames := []string{"smbd", "samba", "smb"}

	for _, name := range serviceNames {
		// 检查服务是否存在
		cmd := exec.Command("systemctl", "list-unit-files", name+".service")
		if err := cmd.Run(); err == nil {
			s.serviceName = name
			fmt.Printf("检测到Samba服务: %s\n", name)
			return s.serviceName
		}
	}

	// 如果都找不到，使用默认名称
	fmt.Println("警告: 未检测到已知的Samba服务，使用默认服务名称:", s.serviceName)
	return s.serviceName
}

// Start 启动Samba服务
func (s *SambaService) Start() error {
	serviceName := s.detectSambaService()
	cmd := exec.Command("sudo", "systemctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动Samba服务失败: %w", err)
	}
	return nil
}

// Stop 停止Samba服务
func (s *SambaService) Stop() error {
	serviceName := s.detectSambaService()
	cmd := exec.Command("sudo", "systemctl", "stop", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("停止Samba服务失败: %w", err)
	}
	return nil
}

// Restart 重启Samba服务
func (s *SambaService) Restart() error {
	serviceName := s.detectSambaService()
	cmd := exec.Command("sudo", "systemctl", "restart", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("重启Samba服务失败: %w", err)
	}
	return nil
}

// Enable 启用Samba服务开机自启
func (s *SambaService) Enable() error {
	serviceName := s.detectSambaService()
	cmd := exec.Command("sudo", "systemctl", "enable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启用Samba服务开机自启失败: %w", err)
	}
	return nil
}

// Disable 禁用Samba服务开机自启
func (s *SambaService) Disable() error {
	serviceName := s.detectSambaService()
	cmd := exec.Command("sudo", "systemctl", "disable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("禁用Samba服务开机自启失败: %w", err)
	}
	return nil
}

// GetStatus 获取Samba服务状态
func (s *SambaService) GetStatus() map[string]interface{} {
	serviceName := s.detectSambaService()
	status := map[string]interface{}{
		"installed": false,
		"running":   false,
		"enabled":   false,
	}

	// 检查服务是否已安装 - 简化版本检查
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	if err := cmd.Run(); err == nil {
		status["installed"] = true

		// 检查服务是否正在运行
		cmd = exec.Command("systemctl", "is-active", serviceName)
		if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) == "active" {
			status["running"] = true
		}

		// 检查服务是否开机自启
		cmd = exec.Command("systemctl", "is-enabled", serviceName)
		if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) == "enabled" {
			status["enabled"] = true
		}
	}

	return status
}

// hasWritePermission 检查当前用户是否有写入/etc/samba/smb.conf的权限
func (s *SambaService) hasWritePermission() bool {
	file, err := os.OpenFile("/etc/samba/smb.conf", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// GenerateShares 生成Samba共享配置到/etc/samba/smb.conf文件
func (s *SambaService) GenerateShares(shares []SambaShare) error {
	// 检查权限，如果没有写权限则使用sudo
	needSudo := !s.hasWritePermission()

	// 准备配置内容
	configLines, err := s.prepareSmbConfig(shares, needSudo)
	if err != nil {
		return err
	}

	// 根据权限情况写入文件
	if needSudo {
		return s.writeSmbConfigWithSudo(configLines)
	} else {
		return s.writeSmbConfig(configLines)
	}
}

// prepareSmbConfig 准备smb.conf配置内容
func (s *SambaService) prepareSmbConfig(shares []SambaShare, useSudo bool) ([]string, error) {
	var lines []string

	// 读取现有配置
	var existingLines []string
	var err error

	if useSudo {
		existingLines, err = s.readSmbConfigWithSudo()
	} else {
		existingLines, err = s.readSmbConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("读取现有配置失败: %w", err)
	}

	// 创建当前要导出的共享名称映射
	currentShareNames := make(map[string]bool)
	for _, share := range shares {
		currentShareNames[share.Name] = true
	}

	// 过滤掉由本系统管理的配置段和重复的配置
	filteredLines := []string{}
	inMinasSection := false
	inShareSection := false

	for _, line := range existingLines {
		trimmedLine := strings.TrimSpace(line)

		// 检查是否是Minas管理的标记
		if trimmedLine == "# Samba shares managed by Minas" {
			inMinasSection = true
			continue // 跳过这个标记行
		}

		// 检查是否是共享段开始
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			sectionName := strings.Trim(trimmedLine, "[]")

			// 如果在Minas管理的区域内，或者是我们要管理的共享，跳过整个段
			if inMinasSection || currentShareNames[sectionName] {
				inShareSection = true
				continue
			} else {
				inMinasSection = false
				inShareSection = false
			}
		}

		// 如果在被管理的段内，跳过所有行
		if inMinasSection || inShareSection {
			// 遇到空行可能表示段结束
			if trimmedLine == "" {
				inShareSection = false
			}
			continue
		}

		// 保留其他行
		filteredLines = append(filteredLines, line)
	}

	// 添加现有配置
	lines = append(lines, filteredLines...)

	// 如果有共享需要导出，添加Minas管理标记和新配置
	if len(shares) > 0 {
		// 确保有空行分隔
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}

		// 添加Minas管理标记
		lines = append(lines, "# Samba shares managed by Minas")
		lines = append(lines, "")

		// 添加所有启用的共享配置
		for _, share := range shares {
			if share.IsDisable == 0 { // IsDisable=0表示启用
				shareConfig := strings.Split(share.GetSmbConfigSection(), "\n")
				for _, configLine := range shareConfig {
					if strings.TrimSpace(configLine) != "" {
						lines = append(lines, configLine)
					}
				}
				lines = append(lines, "") // 共享段之间添加空行
			}
		}
	}

	return lines, nil
}

// readSmbConfig 读取/etc/samba/smb.conf文件内容
func (s *SambaService) readSmbConfig() ([]string, error) {
	file, err := os.Open("/etc/samba/smb.conf")
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // 文件不存在返回空切片
		}
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// readSmbConfigWithSudo 使用sudo读取/etc/samba/smb.conf文件内容
func (s *SambaService) readSmbConfigWithSudo() ([]string, error) {
	cmd := exec.Command("sudo", "cat", "/etc/samba/smb.conf")
	output, err := cmd.Output()
	if err != nil {
		// 如果文件不存在，返回空切片
		if strings.Contains(err.Error(), "No such file") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("读取/etc/samba/smb.conf文件失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	// 移除最后一个空行（如果存在）
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

// writeSmbConfig 写入/etc/samba/smb.conf文件
func (s *SambaService) writeSmbConfig(lines []string) error {
	file, err := os.OpenFile("/etc/samba/smb.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// writeSmbConfigWithSudo 使用sudo写入/etc/samba/smb.conf文件
func (s *SambaService) writeSmbConfigWithSudo(lines []string) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "smb_conf_*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // 清理临时文件

	// 写入临时文件
	writer := bufio.NewWriter(tmpFile)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			tmpFile.Close()
			return fmt.Errorf("写入临时文件失败: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("刷新临时文件失败: %w", err)
	}
	tmpFile.Close()

	// 确保/etc/samba目录存在
	cmd := exec.Command("sudo", "mkdir", "-p", "/etc/samba")
	cmd.Run() // 忽略错误，目录可能已存在

	// 使用sudo复制临时文件到/etc/samba/smb.conf
	cmd = exec.Command("sudo", "cp", tmpFile.Name(), "/etc/samba/smb.conf")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("复制文件到/etc/samba/smb.conf失败: %w", err)
	}

	// 设置正确的权限
	cmd = exec.Command("sudo", "chmod", "644", "/etc/samba/smb.conf")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置/etc/samba/smb.conf权限失败: %w", err)
	}

	return nil
}

// ReloadConfig 重新加载Samba配置
func (s *SambaService) ReloadConfig() error {
	// 首先验证配置
	if err := s.TestConfig(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 重新加载配置
	cmd := exec.Command("sudo", "smbcontrol", "reload-config")
	if err := cmd.Run(); err != nil {
		// 如果reload-config失败，尝试重启服务
		return s.Restart()
	}

	return nil
}

// TestConfig 测试Samba配置
func (s *SambaService) TestConfig() error {
	cmd := exec.Command("testparm", "-s")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("配置测试失败: %w, 输出: %s", err, string(output))
	}
	return nil
}

// ParseSmbConfig 解析/etc/samba/smb.conf文件，提取Samba共享配置
func (s *SambaService) ParseSmbConfig() ([]SambaShare, error) {
	// 读取smb.conf文件
	var lines []string
	var err error

	if s.hasWritePermission() {
		lines, err = s.readSmbConfig()
	} else {
		lines, err = s.readSmbConfigWithSudo()
	}

	if err != nil {
		return nil, fmt.Errorf("读取smb.conf文件失败: %w", err)
	}

	var shares []SambaShare
	var currentShare *SambaShare
	var currentSection string

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 检查是否是段开始
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.Trim(line, "[]")
			currentSection = sectionName

			// 如果是系统段（global、homes、printers等），不创建共享对象
			if sectionName == "global" || sectionName == "homes" || sectionName == "printers" || sectionName == "print$" {
				// 保存之前的共享对象
				if currentShare != nil {
					shares = append(shares, *currentShare)
				}
				currentShare = nil
				continue
			}

			// 开始新的共享段
			if currentShare != nil {
				shares = append(shares, *currentShare)
			}

			currentShare = &SambaShare{
				Name:          sectionName,
				Browseable:    true,  // Samba默认值
				ReadOnly:      true,  // Samba默认值
				GuestOk:       false, // Samba默认值
				CreateMask:    "0744",
				DirectoryMask: "0755",
			}
			// 设置默认启用状态
			currentShare.IsDisable = 0
			continue
		}

		// 解析配置行
		if currentShare != nil {
			// 只有在共享段中才解析配置项
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch strings.ToLower(key) {
			case "path":
				currentShare.Path = value
			case "comment":
				currentShare.Comment = value
			case "browseable", "browsable":
				currentShare.Browseable = strings.ToLower(value) == "yes" || strings.ToLower(value) == "true"
			case "read only":
				currentShare.ReadOnly = strings.ToLower(value) == "yes" || strings.ToLower(value) == "true"
			case "guest ok":
				currentShare.GuestOk = strings.ToLower(value) == "yes" || strings.ToLower(value) == "true"
			case "valid users":
				currentShare.ValidUsers = value
			case "write list":
				currentShare.WriteList = value
			case "read list":
				currentShare.ReadList = value
			case "create mask", "create mode":
				currentShare.CreateMask = value
			case "directory mask", "directory mode":
				currentShare.DirectoryMask = value
			case "force user":
				currentShare.ForceUser = value
			case "force group":
				currentShare.ForceGroup = value
			default:
				// 其他选项添加到Options字段
				if currentShare.Options != "" {
					currentShare.Options += "\n"
				}
				currentShare.Options += fmt.Sprintf("%s = %s", key, value)
			}
		} else if currentSection != "" {
			// 在系统段中，不报错，静默跳过配置项
			// 这样可以避免"配置项不在任何段中"的警告
			continue
		} else {
			// 这种情况下确实是配置项不在任何段中
			fmt.Printf("警告: 第%d行配置项不在任何段中: %s\n", lineNum+1, line)
		}
	}

	// 添加最后一个共享
	if currentShare != nil {
		shares = append(shares, *currentShare)
	}

	return shares, nil
}

// SyncFromServer 从服务器同步Samba配置到数据库
func (s *SambaService) SyncFromServer() (int, error) {
	// 解析服务器上的配置
	serverShares, err := s.ParseSmbConfig()
	if err != nil {
		return 0, fmt.Errorf("解析服务器配置失败: %w", err)
	}

	if len(serverShares) == 0 {
		return 0, nil
	}

	// 获取数据库中现有的配置
	var dbShares []SambaShare
	if err := global.DB.Find(&dbShares).Error; err != nil {
		return 0, fmt.Errorf("查询数据库配置失败: %w", err)
	}

	// 创建数据库配置的共享名称映射，用于重复检测
	dbShareMap := make(map[string]SambaShare)
	for _, share := range dbShares {
		dbShareMap[share.Name] = share
	}

	syncCount := 0
	for _, serverShare := range serverShares {
		// 检查数据库中是否已存在相同名称的配置
		if _, exists := dbShareMap[serverShare.Name]; !exists {
			// 不存在重复，添加到数据库
			serverShare.Remark = "从服务器同步"
			if err := global.DB.Create(&serverShare).Error; err != nil {
				return syncCount, fmt.Errorf("保存配置失败 [%s]: %w", serverShare.Name, err)
			}
			syncCount++
		} else {
			// 存在重复，跳过同步
			fmt.Printf("跳过重复配置: %s\n", serverShare.Name)
		}
	}

	// 同步完成后，立即重新生成配置以确保一致性
	if syncCount > 0 {
		// 查询所有启用的Samba共享
		var allShares []SambaShare
		if err := global.DB.Where("is_disable = ?", 0).Find(&allShares).Error; err != nil {
			return syncCount, fmt.Errorf("重新查询配置失败: %w", err)
		}

		// 重新生成所有配置，这样可以确保smb.conf文件由系统统一管理
		if err := s.GenerateShares(allShares); err != nil {
			// 生成失败不影响同步结果，只记录警告
			fmt.Printf("警告: 同步后重新生成配置失败: %v\n", err)
		}
	}

	return syncCount, nil
}

// SyncToServer 将数据库配置同步到服务器
func (s *SambaService) SyncToServer() error {
	// 查询所有启用的Samba共享
	var shares []SambaShare
	if err := global.DB.Where("is_disable = ?", 0).Find(&shares).Error; err != nil {
		return fmt.Errorf("查询数据库配置失败: %w", err)
	}

	// 生成配置到服务器
	if err := s.GenerateShares(shares); err != nil {
		return fmt.Errorf("生成配置到服务器失败: %w", err)
	}

	// 重新加载配置
	if err := s.ReloadConfig(); err != nil {
		return fmt.Errorf("重新加载服务器配置失败: %w", err)
	}

	return nil
}

// GetSyncStatus 获取同步状态信息
func (s *SambaService) GetSyncStatus() (map[string]interface{}, error) {
	// 解析服务器配置
	serverShares, err := s.ParseSmbConfig()
	if err != nil {
		return nil, fmt.Errorf("解析服务器配置失败: %w", err)
	}

	// 查询数据库配置
	var dbShares []SambaShare
	if err := global.DB.Find(&dbShares).Error; err != nil {
		return nil, fmt.Errorf("查询数据库配置失败: %w", err)
	}

	// 分析差异
	serverNames := make(map[string]bool)
	for _, share := range serverShares {
		serverNames[share.Name] = true
	}

	dbNames := make(map[string]bool)
	enabledDbNames := make(map[string]bool)
	for _, share := range dbShares {
		dbNames[share.Name] = true
		if share.IsDisable == 0 { // IsDisable=0表示启用
			enabledDbNames[share.Name] = true
		}
	}

	// 计算差异
	serverOnlyCount := 0
	for name := range serverNames {
		if !dbNames[name] {
			serverOnlyCount++
		}
	}

	dbOnlyCount := 0
	for name := range enabledDbNames {
		if !serverNames[name] {
			dbOnlyCount++
		}
	}

	return map[string]interface{}{
		"server_shares_count":   len(serverShares),
		"db_shares_count":       len(dbShares),
		"server_only_count":     serverOnlyCount,
		"db_only_count":         dbOnlyCount,
		"need_sync_from_server": serverOnlyCount > 0,
		"need_sync_to_server":   dbOnlyCount > 0,
	}, nil
}

// GetSmbConfig 获取smb.conf配置文件内容
func (s *SambaService) GetSmbConfig() (string, error) {
	var lines []string
	var err error

	if s.hasWritePermission() {
		lines, err = s.readSmbConfig()
	} else {
		lines, err = s.readSmbConfigWithSudo()
	}

	if err != nil {
		return "", fmt.Errorf("读取smb.conf配置文件失败: %w", err)
	}

	// 将所有行合并为一个字符串
	return strings.Join(lines, "\n"), nil
}

// BackupSmbConfig 备份smb.conf配置文件
func (s *SambaService) BackupSmbConfig() error {
	backupPath := fmt.Sprintf("/etc/samba/smb.conf.backup_%d", time.Now().Unix())
	cmd := exec.Command("sudo", "cp", "/etc/samba/smb.conf", backupPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("备份smb.conf失败: %w", err)
	}
	return nil
}

// ListSambaUsers 获取Samba用户列表
func (s *SambaService) ListSambaUsers() ([]map[string]interface{}, error) {
	cmd := exec.Command("sudo", "pdbedit", "-L")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取Samba用户列表失败: %w", err)
	}

	var users []map[string]interface{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// pdbedit -L 输出格式通常是: username:uid:
		parts := strings.Split(line, ":")
		if len(parts) >= 1 {
			users = append(users, map[string]interface{}{
				"username": parts[0],
			})
		}
	}

	return users, nil
}

// AddSambaUser 添加Samba用户
func (s *SambaService) AddSambaUser(username, password string) error {
	// 验证用户名格式
	if !isValidUsername(username) {
		return fmt.Errorf("无效的用户名格式")
	}

	// 检查系统用户是否已存在
	cmd := exec.Command("id", username)
	if err := cmd.Run(); err != nil {
		// 系统用户不存在，先创建系统用户
		if err := s.createSystemUser(username); err != nil {
			return fmt.Errorf("创建系统用户失败: %w", err)
		}
	}

	// 使用smbpasswd添加Samba用户
	cmd = exec.Command("sudo", "smbpasswd", "-a", username)
	cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("添加Samba用户失败: %w", err)
	}

	return nil
}

// createSystemUser 创建系统用户（禁止登录）
func (s *SambaService) createSystemUser(username string) error {
	// 使用useradd创建用户，设置为nologin shell禁止登录
	// -m: 创建home目录
	// -s /usr/sbin/nologin: 设置shell为nologin，禁止用户登录系统
	// -c: 设置用户注释
	cmd := exec.Command("sudo", "useradd", "-m", "-s", "/usr/sbin/nologin", "-c", "Samba user", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建系统用户失败: %w", err)
	}

	return nil
}

// DeleteSambaUser 删除Samba用户
func (s *SambaService) DeleteSambaUser(username string) error {
	// 验证用户名格式
	if !isValidUsername(username) {
		return fmt.Errorf("无效的用户名格式")
	}

	// 检查Samba用户是否存在
	cmd := exec.Command("sudo", "pdbedit", "-L", "-u", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("samba用户 %s 不存在", username)
	}

	// 删除Samba用户
	cmd = exec.Command("sudo", "pdbedit", "-x", "-u", username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除Samba用户失败: %w, 输出: %s", err, string(output))
	}

	// 可选：同时删除系统用户（如果是由系统创建的）
	// 注意：这里只删除Samba用户，不删除系统用户，避免意外删除重要的系统用户
	// 如果需要删除系统用户，可以单独提供接口

	return nil
}

// DeleteSambaUserCompletely 完全删除Samba用户（包括系统用户）
func (s *SambaService) DeleteSambaUserCompletely(username string) error {
	// 验证用户名格式
	if !isValidUsername(username) {
		return fmt.Errorf("无效的用户名格式")
	}

	// 先删除Samba用户
	if err := s.DeleteSambaUser(username); err != nil {
		// 如果Samba用户不存在，继续尝试删除系统用户
		if !strings.Contains(err.Error(), "不存在") {
			return err
		}
	}

	// 检查系统用户是否存在
	cmd := exec.Command("id", username)
	if err := cmd.Run(); err != nil {
		// 系统用户不存在，直接返回成功
		return nil
	}

	// 删除系统用户及其home目录
	// -r: 删除用户的home目录和邮件池
	cmd = exec.Command("sudo", "userdel", "-r", username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 即使删除home目录失败，也尝试强制删除用户
		cmd = exec.Command("sudo", "userdel", username)
		if err2 := cmd.Run(); err2 != nil {
			return fmt.Errorf("删除系统用户失败: %w, 输出: %s", err, string(output))
		}
	}

	return nil
}

// ChangeSambaUserPassword 修改Samba用户密码
func (s *SambaService) ChangeSambaUserPassword(username, password string) error {
	// 验证用户名格式
	if !isValidUsername(username) {
		return fmt.Errorf("无效的用户名格式")
	}

	// 检查用户是否存在
	cmd := exec.Command("sudo", "pdbedit", "-L", "-u", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("用户 %s 不存在", username)
	}

	// 使用smbpasswd -a修改密码
	cmd = exec.Command("sudo", "smbpasswd", "-a", username)
	cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("修改Samba用户密码失败: %w, 输出: %s", err, string(output))
	}

	return nil
}

// isValidUsername 验证用户名格式
func isValidUsername(username string) bool {
	// 简单的用户名验证：只允许字母、数字、下划线，长度3-20
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]{3,20}$", username)
	return matched
}

// GetDirOwner 获取目录所有者和组信息
func (s *SambaService) GetDirOwner(path string) (string, string, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 获取文件的系统信息
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return "", "", fmt.Errorf("无法获取系统文件信息")
	}

	// 获取用户信息
	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	userInfo, err := user.LookupId(uid)
	if err != nil {
		return "", "", fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 获取组信息
	gid := strconv.FormatUint(uint64(stat.Gid), 10)
	groupInfo, err := user.LookupGroupId(gid)
	if err != nil {
		return "", "", fmt.Errorf("获取组信息失败: %w", err)
	}

	return userInfo.Username, groupInfo.Name, nil
}
