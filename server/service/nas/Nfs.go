//go:build linux
// +build linux

// Package nas 提供网络存储相关的服务和模型
package nas

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"server/core/db"
	"server/utils/global"
	"server/utils/system"
	"strings"
	"time"

	"gorm.io/gorm"
)

// NfsShare 定义NFS共享配置模型
// 用于存储NFS共享信息，包括路径、权限、访问控制等
type NfsShare struct {
	db.BaseModel[NfsShare] // 继承基础模型，提供通用字段和方法

	// 基本信息
	Name string `gorm:"comment:'共享名称'" json:"name"` // NFS共享名称
	Path string `gorm:"comment:'共享路径'" json:"path"` // NFS共享目录路径

	// 访问控制
	ClientIP   string `gorm:"comment:'客户端IP'" json:"clientIp"`  // 允许访问的客户端IP，支持网段和通配符
	Permission string `gorm:"comment:'读写权限'" json:"permission"` // 权限：ro(只读), rw(读写)

	// 高级选项
	Sync         string `gorm:"comment:'同步模式'" json:"sync"`           // 同步模式：sync(同步), async(异步)
	RootSquash   string `gorm:"comment:'Root权限映射'" json:"rootSquash"` // root_squash, no_root_squash, all_squash
	SubtreeCheck string `gorm:"comment:'子树检查'" json:"subtreeCheck"`   // subtree_check, no_subtree_check

	// 其他选项
	Options string `gorm:"comment:'其他选项'" json:"options"` // 其他自定义NFS选项
	Remark  string `gorm:"comment:'备注'" json:"remark"`    // 备注信息
}

// TableName 返回NFS共享表名
// 实现gorm的Tabler接口
func (NfsShare) TableName() string {
	return "nfs_shares" // 数据库表名
}

// BeforeCreate 创建记录前的钩子函数
func (n *NfsShare) BeforeCreate(tx *gorm.DB) (err error) {
	// 调用父类的创建前函数
	n.SupperBeforeCreate()
	// 设置默认状态为启用（IsDisable = 0表示启用）
	if n.IsDisable == 0 {
		n.IsDisable = 0 // 默认启用
	}
	return
}

// AfterFind 查询记录后的钩子函数
func (n *NfsShare) AfterFind(tx *gorm.DB) (err error) {
	// 调用父类的查询后函数
	n.SupperAfterFind()
	return
}

// GetExportOptions 获取完整的exports选项字符串
func (n *NfsShare) GetExportOptions() string {
	options := []string{}

	// 添加读写权限
	if n.Permission != "" {
		options = append(options, n.Permission)
	} else {
		options = append(options, "ro") // 默认只读
	}

	// 添加同步模式
	if n.Sync != "" {
		options = append(options, n.Sync)
	} else {
		options = append(options, "sync") // 默认同步
	}

	// 添加root权限映射（只有在明确设置时才添加）
	if n.RootSquash != "" {
		options = append(options, n.RootSquash)
	}

	// 添加子树检查选项（只有在明确设置时才添加）
	if n.SubtreeCheck != "" {
		options = append(options, n.SubtreeCheck)
	}

	// 添加其他自定义选项
	if n.Options != "" {
		customOptions := strings.Split(n.Options, ",")
		for _, opt := range customOptions {
			opt = strings.TrimSpace(opt)
			if opt != "" {
				options = append(options, opt)
			}
		}
	}

	return strings.Join(options, ",")
}

// GetExportLine 获取完整的exports配置行
func (n *NfsShare) GetExportLine() string {
	clientIP := n.ClientIP
	if clientIP == "" {
		clientIP = "*" // 默认允许所有IP
	}

	return fmt.Sprintf("%s %s(%s)", n.Path, clientIP, n.GetExportOptions())
}

// NfsService NFS服务管理
type NfsService struct {
	// 缓存检测到的NFS服务名称
	serviceName string
}

// detectNfsService 检测当前系统的NFS服务名称
func (s *NfsService) detectNfsService() string {
	if s.serviceName != "" {
		return s.serviceName
	}

	// 按优先级检测NFS服务名称
	serviceNames := []string{
		"nfs-server",        // CentOS/RHEL/Rocky Linux
		"nfs-kernel-server", // Ubuntu/Debian
		"nfs",               // 旧版本系统
		"nfsd",              // 某些系统变体
	}

	for _, serviceName := range serviceNames {
		cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
		if err := cmd.Run(); err == nil {
			s.serviceName = serviceName
			fmt.Println("使用服务名称:", s.serviceName)
			return serviceName
		}
	}

	// 如果都没找到，默认使用nfs-server
	s.serviceName = "nfs-server"
	fmt.Println("警告: 未检测到已知的NFS服务，使用默认服务名称:", s.serviceName)
	return s.serviceName
}

// IsInstalled 检查NFS服务是否已安装
func (s *NfsService) IsInstalled() bool {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	err := cmd.Run()
	return err == nil
}

// IsRunning 检查NFS服务是否正在运行
func (s *NfsService) IsRunning() bool {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "active"
}

// IsEnabled 检查NFS服务是否已设置开机自启
func (s *NfsService) IsEnabled() bool {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "is-enabled", serviceName)
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "enabled"
}

// Start 启动NFS服务
func (s *NfsService) Start() error {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动NFS服务失败: %w", err)
	}
	return nil
}

// Stop 停止NFS服务
func (s *NfsService) Stop() error {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "stop", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("停止NFS服务失败: %w", err)
	}
	return nil
}

// Restart 重启NFS服务
func (s *NfsService) Restart() error {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "restart", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("重启NFS服务失败: %w", err)
	}
	return nil
}

// Enable 启用NFS服务开机自启
func (s *NfsService) Enable() error {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "enable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置NFS开机自启失败: %w", err)
	}
	return nil
}

// Disable 禁用NFS服务开机自启
func (s *NfsService) Disable() error {
	serviceName := s.detectNfsService()
	cmd := exec.Command("systemctl", "disable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("禁用NFS开机自启失败: %w", err)
	}
	return nil
}

// GetStatus 获取NFS服务状态
func (s *NfsService) GetStatus() map[string]interface{} {
	// 检查是否在容器中运行
	if system.IsRunningInContainer() {
		return map[string]interface{}{
			"installed": false,
			"running":   false,
			"enabled":   false,
			"available": false,
			"message":   "NFS服务在容器环境中不可用",
		}
	}

	return map[string]interface{}{
		"installed": s.IsInstalled(),
		"running":   s.IsRunning(),
		"enabled":   s.IsEnabled(),
		"available": true,
	}
}

// hasWritePermission 检查当前用户是否有写入/etc/exports的权限
func (s *NfsService) hasWritePermission() bool {
	// 检查/etc目录的写权限
	info, err := os.Stat("/etc/exports")
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，检查/etc目录的写权限
			return s.canWriteToDir("/etc")
		}
		return false
	}

	// 文件存在，检查是否可以写入
	file, err := os.OpenFile("/etc/exports", os.O_WRONLY|os.O_APPEND, info.Mode())
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// canWriteToDir 检查是否可以在指定目录写入文件
func (s *NfsService) canWriteToDir(dir string) bool {
	// 尝试创建临时文件来测试写权限
	tmpFile, err := os.CreateTemp(dir, ".test_write_*")
	if err != nil {
		return false
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name())
	return true
}

// ExportShares 导出NFS共享配置到/etc/exports文件
func (s *NfsService) ExportShares(shares []NfsShare) error {
	// 检查权限，如果没有写权限则使用sudo
	needSudo := !s.hasWritePermission()

	// 准备配置内容
	configLines, err := s.prepareExportConfig(shares, needSudo)
	if err != nil {
		return err
	}

	// 根据权限情况写入文件
	if needSudo {
		return s.writeExportsFileWithSudo(configLines)
	} else {
		return s.writeExportsFile(configLines)
	}
}

// prepareExportConfig 准备exports配置内容
func (s *NfsService) prepareExportConfig(shares []NfsShare, useSudo bool) ([]string, error) {
	// 读取现有的exports文件
	var existingLines []string
	var err error

	if useSudo {
		existingLines, err = s.readExportsFileWithSudo()
	} else {
		existingLines, err = s.readExportsFile()
	}

	if err != nil {
		return nil, err
	}

	// 创建当前要导出的配置映射，用于精确重复检测
	currentShareKeys := make(map[string]bool)
	for _, share := range shares {
		if share.IsDisable == 0 { // 只考虑启用的共享 (IsDisable=0表示启用)
			key := s.generateShareKey(share.Path, share.ClientIP)
			currentShareKeys[key] = true
		}
	}

	// 过滤掉由本系统管理的配置行和重复的配置
	filteredLines := []string{}
	inMinasSection := false

	for _, line := range existingLines {
		trimmedLine := strings.TrimSpace(line)

		// 检查是否是Minas管理的标记
		if trimmedLine == "# NFS shares managed by Minas" {
			inMinasSection = true
			continue // 跳过这个标记行
		}

		// 如果在Minas管理的区域内，跳过非注释行
		if inMinasSection {
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
				// 遇到空行或新的注释，结束Minas区域
				if trimmedLine != "" {
					inMinasSection = false
					filteredLines = append(filteredLines, line)
				}
			}
			// 跳过Minas管理的配置行
			continue
		}

		// 检查当前行是否与我们要导出的配置重复
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
			// 尝试解析当前行的路径和客户端IP
			if existingShares, err := s.parseExportLine(trimmedLine); err == nil && len(existingShares) > 0 {
				// 检查是否有任何客户端IP与当前要导出的配置重复
				for _, existingShare := range existingShares {
					existingKey := s.generateShareKey(existingShare.Path, existingShare.ClientIP)
					if currentShareKeys[existingKey] {
						// 存在重复配置（路径+客户端IP相同），跳过现有条目
						fmt.Printf("跳过重复的exports配置: %s %s\n", existingShare.Path, existingShare.ClientIP)
						goto skipLine
					}
				}
			}
		}

		// 保留非重复的配置
		filteredLines = append(filteredLines, line)
		continue

	skipLine:
		// 跳过这一行，继续处理下一行
		continue
	}

	// 对要导出的配置进行去重处理
	uniqueShares := s.deduplicateShares(shares)

	// 添加系统管理的共享配置
	if len(uniqueShares) > 0 {
		filteredLines = append(filteredLines, "# NFS shares managed by Minas")
		for _, share := range uniqueShares {
			if share.IsDisable == 0 { // 只导出启用的共享 (IsDisable=0表示启用)
				filteredLines = append(filteredLines, share.GetExportLine())
			}
		}
	}

	return filteredLines, nil
}

// deduplicateShares 对要导出的共享配置进行去重
func (s *NfsService) deduplicateShares(shares []NfsShare) []NfsShare {
	shareMap := make(map[string]NfsShare)
	var uniqueShares []NfsShare

	for _, share := range shares {
		if share.IsDisable == 0 { // 只处理启用的共享 (IsDisable=0表示启用)
			key := s.generateShareKey(share.Path, share.ClientIP)

			// 如果已存在相同的key，检查是否需要更新
			if existingShare, exists := shareMap[key]; exists {
				// 保留ID较大的（通常是更新的记录）
				if share.ID > existingShare.ID {
					shareMap[key] = share
				}
			} else {
				shareMap[key] = share
			}
		}
	}

	// 转换为切片
	for _, share := range shareMap {
		uniqueShares = append(uniqueShares, share)
	}

	return uniqueShares
}

// ReloadExports 重新加载NFS exports配置
func (s *NfsService) ReloadExports() error {
	// 首先清理可能存在的锁文件
	if err := s.cleanupNfsLocks(); err != nil {
		// 锁文件清理失败不应该阻止重新加载，只记录警告
		fmt.Printf("警告: 清理NFS锁文件失败: %v\n", err)
	}

	var cmd *exec.Cmd
	var err error

	// 首先尝试使用sudo执行，因为通常需要管理员权限
	cmd = exec.Command("sudo", "exportfs", "-ra")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// 如果sudo失败，尝试直接执行
		cmd = exec.Command("exportfs", "-ra")
		output2, err2 := cmd.CombinedOutput()

		if err2 != nil {
			// 两种方式都失败，返回详细错误信息
			return fmt.Errorf("重新加载NFS exports失败:\n使用sudo: %w, 输出: %s\n直接执行: %v, 输出: %s",
				err, string(output), err2, string(output2))
		}

		// 直接执行成功，但可能有警告
		if len(output2) > 0 {
			fmt.Printf("exportfs警告: %s\n", string(output2))
		}
	} else {
		// sudo执行成功，但可能有警告
		if len(output) > 0 {
			fmt.Printf("exportfs警告: %s\n", string(output))
		}
	}

	return nil
}

// cleanupNfsLocks 清理NFS锁文件
func (s *NfsService) cleanupNfsLocks() error {
	lockFiles := []string{
		"/var/lib/nfs/.etab.lock",
		"/var/lib/nfs/etab.lock",
	}

	for _, lockFile := range lockFiles {
		// 检查锁文件是否存在
		if _, err := os.Stat(lockFile); err == nil {
			// 尝试删除锁文件
			if err := os.Remove(lockFile); err != nil {
				// 如果普通权限删除失败，尝试使用sudo
				cmd := exec.Command("sudo", "rm", "-f", lockFile)
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("删除锁文件 %s 失败: %w", lockFile, err)
				}
			}
		}
	}

	return nil
}

// readExportsFile 读取/etc/exports文件内容
func (s *NfsService) readExportsFile() ([]string, error) {
	file, err := os.Open("/etc/exports")
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

// readExportsFileWithSudo 使用sudo读取/etc/exports文件内容
func (s *NfsService) readExportsFileWithSudo() ([]string, error) {
	cmd := exec.Command("sudo", "cat", "/etc/exports")
	output, err := cmd.Output()
	if err != nil {
		// 如果文件不存在，返回空切片
		if strings.Contains(err.Error(), "No such file") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("读取/etc/exports文件失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	// 移除最后一个空行（如果存在）
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

// writeExportsFile 写入/etc/exports文件
func (s *NfsService) writeExportsFile(lines []string) error {
	file, err := os.OpenFile("/etc/exports", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

// writeExportsFileWithSudo 使用sudo写入/etc/exports文件
func (s *NfsService) writeExportsFileWithSudo(lines []string) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "exports_*")
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

	// 使用sudo复制临时文件到/etc/exports
	cmd := exec.Command("sudo", "cp", tmpFile.Name(), "/etc/exports")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("复制文件到/etc/exports失败: %w", err)
	}

	// 设置正确的权限
	cmd = exec.Command("sudo", "chmod", "644", "/etc/exports")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置/etc/exports权限失败: %w", err)
	}

	return nil
}

// ParseExportsFile 解析/etc/exports文件，提取NFS共享配置
func (s *NfsService) ParseExportsFile() ([]NfsShare, error) {
	// 读取exports文件
	var lines []string
	var err error

	if s.hasWritePermission() {
		lines, err = s.readExportsFile()
	} else {
		lines, err = s.readExportsFileWithSudo()
	}

	if err != nil {
		return nil, fmt.Errorf("读取exports文件失败: %w", err)
	}

	var shares []NfsShare
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析exports行，现在支持多客户端配置
		parsedShares, err := s.parseExportLine(line)
		if err != nil {
			// 记录解析错误但继续处理其他行
			fmt.Printf("警告: 第%d行解析失败: %s, 错误: %v\n", lineNum+1, line, err)
			continue
		}

		// 添加所有解析出的共享配置
		for _, share := range parsedShares {
			if share != nil {
				shares = append(shares, *share)
			}
		}
	}

	return shares, nil
}

// parseExportLine 解析单行exports配置
func (s *NfsService) parseExportLine(line string) ([]*NfsShare, error) {
	// exports格式: /path client1(options) client2(options) ...
	// 支持一个路径对应多个客户端配置

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("无效的exports行格式")
	}

	path := parts[0]
	var shares []*NfsShare

	// 解析所有客户端配置
	for i := 1; i < len(parts); i++ {
		clientPart := parts[i]

		// 查找括号位置
		openParen := strings.Index(clientPart, "(")
		closeParen := strings.LastIndex(clientPart, ")")

		if openParen == -1 || closeParen == -1 || closeParen <= openParen {
			return nil, fmt.Errorf("无效的客户端配置格式: %s", clientPart)
		}

		clientIP := clientPart[:openParen]
		optionsStr := clientPart[openParen+1 : closeParen]

		// 创建共享配置
		share := &NfsShare{
			Path:     path,
			ClientIP: clientIP,
		}
		// 设置默认为启用状态（IsDisable = 0表示启用）
		share.IsDisable = 0

		// 从路径生成名称，添加客户端IP后缀以区分
		baseName := s.generateShareName(path)
		if clientIP == "*" {
			share.Name = baseName
		} else {
			// 清理IP地址中的特殊字符，用于生成名称
			cleanIP := strings.ReplaceAll(clientIP, ".", "_")
			cleanIP = strings.ReplaceAll(cleanIP, "/", "_")
			share.Name = fmt.Sprintf("%s_%s", baseName, cleanIP)
		}

		// 解析选项字符串
		s.parseOptions(share, optionsStr)

		shares = append(shares, share)
	}

	return shares, nil
}

// generateShareName 从路径生成共享名称
func (s *NfsService) generateShareName(path string) string {
	// 移除路径分隔符并取最后一部分作为名称
	name := strings.TrimSuffix(path, "/")
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		name = parts[len(parts)-1]
	}
	if name == "" {
		name = "root"
	}
	return name + "_share"
}

// parseOptions 解析NFS选项字符串
func (s *NfsService) parseOptions(share *NfsShare, optionsStr string) {
	options := strings.Split(optionsStr, ",")
	var customOptions []string

	for _, opt := range options {
		opt = strings.TrimSpace(opt)
		switch opt {
		case "ro":
			share.Permission = "ro"
		case "rw":
			share.Permission = "rw"
		case "sync":
			share.Sync = "sync"
		case "async":
			share.Sync = "async"
		case "root_squash":
			share.RootSquash = "root_squash"
		case "no_root_squash":
			share.RootSquash = "no_root_squash"
		case "all_squash":
			share.RootSquash = "all_squash"
		case "subtree_check":
			share.SubtreeCheck = "subtree_check"
		case "no_subtree_check":
			share.SubtreeCheck = "no_subtree_check"
		default:
			// 其他选项作为自定义选项
			if opt != "" {
				customOptions = append(customOptions, opt)
			}
		}
	}

	// 设置默认值 - 只为必需的字段设置默认值
	if share.Permission == "" {
		share.Permission = "ro"
	}
	if share.Sync == "" {
		share.Sync = "sync"
	}
	// 不再为RootSquash和SubtreeCheck设置默认值，保持空字符串表示使用NFS默认行为

	// 组合自定义选项
	if len(customOptions) > 0 {
		share.Options = strings.Join(customOptions, ",")
	}
}

// SyncFromServer 从服务器同步NFS配置到数据库
func (s *NfsService) SyncFromServer() (int, error) {
	// 解析服务器上的配置
	serverShares, err := s.ParseExportsFile()
	if err != nil {
		return 0, fmt.Errorf("解析服务器配置失败: %w", err)
	}

	if len(serverShares) == 0 {
		return 0, nil
	}

	// 获取数据库中现有的配置
	var dbShares []NfsShare
	if err := global.DB.Find(&dbShares).Error; err != nil {
		return 0, fmt.Errorf("查询数据库配置失败: %w", err)
	}

	// 创建数据库配置的路径+客户端IP组合映射，用于重复检测
	dbShareMap := make(map[string]NfsShare)
	for _, share := range dbShares {
		// 使用路径+客户端IP作为唯一标识
		key := s.generateShareKey(share.Path, share.ClientIP)
		dbShareMap[key] = share
	}

	syncCount := 0
	for _, serverShare := range serverShares {
		// 检查数据库中是否已存在相同路径和客户端IP的配置
		key := s.generateShareKey(serverShare.Path, serverShare.ClientIP)
		if _, exists := dbShareMap[key]; !exists {
			// 不存在重复，添加到数据库
			serverShare.Remark = "从服务器同步"
			if err := global.DB.Create(&serverShare).Error; err != nil {
				return syncCount, fmt.Errorf("保存配置失败 [%s]: %w", serverShare.Path, err)
			}
			syncCount++
		} else {
			// 存在重复，跳过同步
			fmt.Printf("跳过重复配置: %s %s\n", serverShare.Path, serverShare.ClientIP)
		}
	}

	// 同步完成后，立即重新导出配置以确保一致性
	// 这样可以将数据库中的所有配置（包括刚同步的）统一管理
	if syncCount > 0 {
		// 查询所有启用的NFS共享
		var allShares []NfsShare
		if err := global.DB.Where("is_disable = ?", 0).Find(&allShares).Error; err != nil {
			return syncCount, fmt.Errorf("重新查询配置失败: %w", err)
		}

		// 重新导出所有配置，这样可以确保exports文件由系统统一管理
		if err := s.ExportShares(allShares); err != nil {
			// 导出失败不影响同步结果，只记录警告
			fmt.Printf("警告: 同步后重新导出配置失败: %v\n", err)
		}
	}

	return syncCount, nil
}

// generateShareKey 生成共享配置的唯一标识
// 基于共享路径和客户端IP组合生成key，用于重复检测
func (s *NfsService) generateShareKey(path, clientIP string) string {
	// 标准化客户端IP，空值或*都表示允许所有IP
	normalizedClientIP := clientIP
	if normalizedClientIP == "" {
		normalizedClientIP = "*"
	}
	return fmt.Sprintf("%s||%s", path, normalizedClientIP)
}

// SyncToServer 将数据库配置同步到服务器
func (s *NfsService) SyncToServer() error {
	// 查询所有启用的NFS共享
	var shares []NfsShare
	if err := global.DB.Where("is_disable = ?", 0).Find(&shares).Error; err != nil {
		return fmt.Errorf("查询数据库配置失败: %w", err)
	}

	// 导出配置到服务器
	if err := s.ExportShares(shares); err != nil {
		return fmt.Errorf("导出配置到服务器失败: %w", err)
	}

	// 重新加载配置
	if err := s.ReloadExports(); err != nil {
		return fmt.Errorf("重新加载服务器配置失败: %w", err)
	}

	return nil
}

// GetSyncStatus 获取同步状态信息
func (s *NfsService) GetSyncStatus() (map[string]interface{}, error) {
	// 解析服务器配置
	serverShares, err := s.ParseExportsFile()
	if err != nil {
		return nil, fmt.Errorf("解析服务器配置失败: %w", err)
	}

	// 查询数据库配置
	var dbShares []NfsShare
	if err := global.DB.Find(&dbShares).Error; err != nil {
		return nil, fmt.Errorf("查询数据库配置失败: %w", err)
	}

	// 分析差异
	serverPaths := make(map[string]bool)
	for _, share := range serverShares {
		serverPaths[share.Path] = true
	}

	dbPaths := make(map[string]bool)
	enabledDbPaths := make(map[string]bool)
	for _, share := range dbShares {
		dbPaths[share.Path] = true
		if share.IsDisable == 0 { // IsDisable=0表示启用
			enabledDbPaths[share.Path] = true
		}
	}

	// 计算需要同步的数量
	serverOnlyCount := 0
	for path := range serverPaths {
		if !dbPaths[path] {
			serverOnlyCount++
		}
	}

	dbOnlyCount := 0
	for path := range enabledDbPaths {
		if !serverPaths[path] {
			dbOnlyCount++
		}
	}

	return map[string]interface{}{
		"server_shares_count":   len(serverShares),
		"db_shares_count":       len(dbShares),
		"enabled_db_count":      len(enabledDbPaths),
		"server_only_count":     serverOnlyCount,
		"db_only_count":         dbOnlyCount,
		"need_sync_from_server": serverOnlyCount > 0,
		"need_sync_to_server":   dbOnlyCount > 0,
	}, nil
}

// GetExportsConfig 获取exports配置文件内容
func (s *NfsService) GetExportsConfig() (string, error) {
	var lines []string
	var err error

	if s.hasWritePermission() {
		lines, err = s.readExportsFile()
	} else {
		lines, err = s.readExportsFileWithSudo()
	}

	if err != nil {
		return "", fmt.Errorf("读取exports配置文件失败: %w", err)
	}

	// 将所有行合并为一个字符串
	return strings.Join(lines, "\n"), nil
}

// BackupExports 备份exports配置文件
func (s *NfsService) BackupExports() error {
	// 生成备份文件名，格式：/etc/exports.backup.YYYYMMDD_HHMMSS
	now := time.Now()
	backupFileName := fmt.Sprintf("/etc/exports.backup.%s", now.Format("20060102_150405"))

	// 检查原文件是否存在
	if _, err := os.Stat("/etc/exports"); os.IsNotExist(err) {
		return fmt.Errorf("exports配置文件不存在，无法备份")
	}

	// 根据权限选择备份方式
	if s.hasWritePermission() {
		return s.copyFileWithPermission("/etc/exports", backupFileName)
	} else {
		return s.copyFileWithSudo("/etc/exports", backupFileName)
	}
}

// copyFileWithPermission 使用普通权限复制文件
func (s *NfsService) copyFileWithPermission(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	// 复制文件内容
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 确保数据写入磁盘
	return destFile.Sync()
}

// copyFileWithSudo 使用sudo权限复制文件
func (s *NfsService) copyFileWithSudo(src, dst string) error {
	cmd := exec.Command("sudo", "cp", src, dst)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("使用sudo复制文件失败: %w", err)
	}
	return nil
}
