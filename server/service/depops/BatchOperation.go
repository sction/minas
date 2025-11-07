package depops

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"server/utils/logger"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// BatchOperationService 批量操作服务
type BatchOperationService struct{}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	NodeID  uint        `json:"node_id"`
	Name    string      `json:"name"`
	Success bool        `json:"success"`
	Output  string      `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// RestartNodesRequest 重启节点请求
type RestartNodesRequest struct {
	NodeIDs []uint `json:"node_ids"`
}

// ManageKeysRequest 管理密钥请求
type ManageKeysRequest struct {
	NodeIDs   []uint `json:"node_ids"`
	PublicKey string `json:"public_key"`
	Action    string `json:"action"` // add, remove
}

// SetTimezoneRequest 设置时区请求
type SetTimezoneRequest struct {
	NodeIDs  []uint `json:"node_ids"`
	Timezone string `json:"timezone"`
}

// FirewallRuleRequest 防火墙规则请求
type FirewallRuleRequest struct {
	NodeIDs []uint   `json:"node_ids"`
	Action  string   `json:"action"` // add_port, remove_port, add_ip, remove_ip
	Ports   []int    `json:"ports,omitempty"`
	IPs     []string `json:"ips,omitempty"`
}

// ExecuteCommandRequest 执行命令请求
type ExecuteCommandRequest struct {
	NodeIDs []uint `json:"node_ids"`
	Command string `json:"command"`
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	NodeIDs     []uint `json:"node_ids"`
	LocalPath   string `json:"local_path"`
	RemotePath  string `json:"remote_path"`
	Permissions string `json:"permissions,omitempty"`
}

// InstallPackageRequest 安装软件包请求
type InstallPackageRequest struct {
	NodeIDs  []uint   `json:"node_ids"`
	Packages []string `json:"packages"`
	Action   string   `json:"action"` // install, remove
}

// DockerImageRequest Docker镜像请求
type DockerImageRequest struct {
	NodeIDs   []uint `json:"node_ids"`
	ImagePath string `json:"image_path"`
	ImageName string `json:"image_name"`
}

// ExecuteScriptRequest 执行脚本请求
type ExecuteScriptRequest struct {
	NodeIDs    []uint `json:"node_ids"`
	ScriptPath string `json:"script_path"`
	ScriptType string `json:"script_type"` // bash, python, etc.
}

// DockerImagesFromDirectoryRequest 从目录部署Docker镜像请求
type DockerImagesFromDirectoryRequest struct {
	NodeIDs       []uint `json:"node_ids"`
	DirectoryPath string `json:"directory_path"`
}

// RestartNodes 批量重启节点
func (s *BatchOperationService) RestartNodes(req RestartNodesRequest) []BatchOperationResult {
	return s.executeCommandOnNodes(req.NodeIDs, "sudo reboot", "重启节点")
}

// ManageAuthKeys 管理认证公钥
func (s *BatchOperationService) ManageAuthKeys(req ManageKeysRequest) []BatchOperationResult {
	var command string
	var description string

	switch req.Action {
	case "add":
		command = fmt.Sprintf("mkdir -p ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && chmod 700 ~/.ssh", req.PublicKey)
		description = "添加公钥"
	case "remove":
		command = fmt.Sprintf("sed -i '\\|%s|d' ~/.ssh/authorized_keys", req.PublicKey)
		description = "删除公钥"
	default:
		results := make([]BatchOperationResult, len(req.NodeIDs))
		for i, nodeID := range req.NodeIDs {
			node, _ := s.getNodeByID(nodeID)
			results[i] = BatchOperationResult{
				NodeID:  nodeID,
				Name:    node.Name,
				Success: false,
				Error:   "不支持的操作类型",
			}
		}
		return results
	}

	return s.executeCommandOnNodes(req.NodeIDs, command, description)
}

// SetTimezone 批量设置时区
func (s *BatchOperationService) SetTimezone(req SetTimezoneRequest) []BatchOperationResult {
	command := fmt.Sprintf("sudo timedatectl set-timezone %s", req.Timezone)
	return s.executeCommandOnNodes(req.NodeIDs, command, "设置时区")
}

// ManageFirewall 管理防火墙规则
func (s *BatchOperationService) ManageFirewall(req FirewallRuleRequest) []BatchOperationResult {
	switch req.Action {
	case "add_port", "remove_port", "add_ip", "remove_ip":
		// 支持的操作类型
	default:
		results := make([]BatchOperationResult, len(req.NodeIDs))
		for i, nodeID := range req.NodeIDs {
			node, _ := s.getNodeByID(nodeID)
			results[i] = BatchOperationResult{
				NodeID:  nodeID,
				Name:    node.Name,
				Success: false,
				Error:   "不支持的防火墙操作类型",
			}
		}
		return results
	}

	// 根据每个节点的包管理类型分别执行防火墙操作
	results := make([]BatchOperationResult, 0)

	for _, nodeID := range req.NodeIDs {
		result := BatchOperationResult{
			NodeID: nodeID,
		}

		node, err := s.getNodeByID(nodeID)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Name = fmt.Sprintf("Node_%d", nodeID)
			results = append(results, result)
			continue
		}

		result.Name = node.Name

		fmt.Printf("Processing node: %s (ID: %d) pm: %s\n", node.Name, node.ID, node.PackageManager)
		// 根据节点的包管理类型生成相应的防火墙命令
		command := s.generateFirewallCommandForNode(req, node.PackageManager)

		output, err := node.ExecuteCommand(command)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Output = output
		} else {
			result.Success = true
			result.Output = output
		}

		results = append(results, result)
	}

	return results
}

// ExecuteCommand 批量执行命令
func (s *BatchOperationService) ExecuteCommand(req ExecuteCommandRequest) []BatchOperationResult {
	return s.executeCommandOnNodes(req.NodeIDs, req.Command, "执行命令")
}

// UploadFile 批量上传文件
func (s *BatchOperationService) UploadFile(req UploadFileRequest) []BatchOperationResult {
	results := make([]BatchOperationResult, 0)

	for _, nodeID := range req.NodeIDs {
		result := BatchOperationResult{
			NodeID: nodeID,
		}

		node, err := s.getNodeByID(nodeID)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Name = fmt.Sprintf("Node_%d", nodeID)
			results = append(results, result)
			continue
		}

		result.Name = node.Name

		// 建立SSH连接
		client, err := node.GetSSHClient()
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("SSH连接失败: %v", err)
			results = append(results, result)
			continue
		}
		defer client.Close()

		// 使用SCP上传文件
		err = s.scpUpload(client, req.LocalPath, req.RemotePath)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("文件上传失败: %v", err)
			results = append(results, result)
			continue
		}

		// 设置文件权限
		if req.Permissions != "" {
			command := fmt.Sprintf("chmod %s %s", req.Permissions, req.RemotePath)
			_, err = node.ExecuteCommand(command)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("设置权限失败: %v", err)
				results = append(results, result)
				continue
			}
		}

		result.Success = true
		result.Output = fmt.Sprintf("文件上传成功: %s -> %s", req.LocalPath, req.RemotePath)
		results = append(results, result)
	}

	return results
}

// InstallPackages 批量安装软件包
func (s *BatchOperationService) InstallPackages(req InstallPackageRequest) []BatchOperationResult {
	var command string
	var description string

	packages := strings.Join(req.Packages, " ")

	switch req.Action {
	case "install":
		// 检测系统类型并使用相应的包管理器
		command = fmt.Sprintf(`
			if command -v apt-get >/dev/null 2>&1; then
				sudo apt-get update && sudo apt-get install -y %s
			elif command -v yum >/dev/null 2>&1; then
				sudo yum install -y %s
			elif command -v dnf >/dev/null 2>&1; then
				sudo dnf install -y %s
			else
				echo "未找到支持的包管理器"
				exit 1
			fi
		`, packages, packages, packages)
		description = "安装软件包"
	case "remove":
		command = fmt.Sprintf(`
			if command -v apt-get >/dev/null 2>&1; then
				sudo apt-get remove -y %s
			elif command -v yum >/dev/null 2>&1; then
				sudo yum remove -y %s
			elif command -v dnf >/dev/null 2>&1; then
				sudo dnf remove -y %s
			else
				echo "未找到支持的包管理器"
				exit 1
			fi
		`, packages, packages, packages)
		description = "卸载软件包"
	default:
		results := make([]BatchOperationResult, len(req.NodeIDs))
		for i, nodeID := range req.NodeIDs {
			node, _ := s.getNodeByID(nodeID)
			results[i] = BatchOperationResult{
				NodeID:  nodeID,
				Name:    node.Name,
				Success: false,
				Error:   "不支持的包管理操作",
			}
		}
		return results
	}

	return s.executeCommandOnNodes(req.NodeIDs, command, description)
}

// DeployDockerImage 批量发送并加载Docker镜像
func (s *BatchOperationService) DeployDockerImage(req DockerImageRequest) []BatchOperationResult {
	results := make([]BatchOperationResult, 0)

	for _, nodeID := range req.NodeIDs {
		result := BatchOperationResult{
			NodeID: nodeID,
		}

		node, err := s.getNodeByID(nodeID)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Name = fmt.Sprintf("Node_%d", nodeID)
			results = append(results, result)
			continue
		}

		result.Name = node.Name

		// 建立SSH连接
		client, err := node.GetSSHClient()
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("SSH连接失败: %v", err)
			results = append(results, result)
			continue
		}
		defer client.Close()

		// 上传Docker镜像文件
		remotePath := fmt.Sprintf("/tmp/%s", filepath.Base(req.ImagePath))
		err = s.scpUpload(client, req.ImagePath, remotePath)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("镜像文件上传失败: %v", err)
			results = append(results, result)
			continue
		}

		// 加载Docker镜像
		command := fmt.Sprintf("docker load -i %s", remotePath)
		output, err := node.ExecuteCommand(command)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Docker镜像加载失败: %v", err)
			result.Output = output
			results = append(results, result)
			continue
		}

		// 清理临时文件
		cleanupCommand := fmt.Sprintf("rm -f %s", remotePath)
		node.ExecuteCommand(cleanupCommand)

		result.Success = true
		result.Output = fmt.Sprintf("Docker镜像加载成功: %s", output)
		results = append(results, result)
	}

	return results
}

// ExecuteScript 批量执行脚本
func (s *BatchOperationService) ExecuteScript(req ExecuteScriptRequest) []BatchOperationResult {
	results := make([]BatchOperationResult, 0)

	for _, nodeID := range req.NodeIDs {
		result := BatchOperationResult{
			NodeID: nodeID,
		}

		node, err := s.getNodeByID(nodeID)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Name = fmt.Sprintf("Node_%d", nodeID)
			results = append(results, result)
			continue
		}

		result.Name = node.Name

		// 建立SSH连接
		client, err := node.GetSSHClient()
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("SSH连接失败: %v", err)
			results = append(results, result)
			continue
		}
		defer client.Close()

		// 上传脚本文件
		remotePath := fmt.Sprintf("/tmp/script_%d_%s", time.Now().Unix(), filepath.Base(req.ScriptPath))
		err = s.scpUpload(client, req.ScriptPath, remotePath)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("脚本文件上传失败: %v", err)
			results = append(results, result)
			continue
		}

		// 设置脚本权限
		chmodCommand := fmt.Sprintf("chmod +x %s", remotePath)
		node.ExecuteCommand(chmodCommand)

		// 执行脚本
		var command string
		switch req.ScriptType {
		case "bash":
			command = fmt.Sprintf("bash %s", remotePath)
		case "python":
			command = fmt.Sprintf("python %s", remotePath)
		case "python3":
			command = fmt.Sprintf("python3 %s", remotePath)
		default:
			command = remotePath
		}

		output, err := node.ExecuteCommand(command)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("脚本执行失败: %v", err)
			result.Output = output
		} else {
			result.Success = true
			result.Output = output
		}

		// 清理临时文件
		cleanupCommand := fmt.Sprintf("rm -f %s", remotePath)
		node.ExecuteCommand(cleanupCommand)

		results = append(results, result)
	}

	return results
}

// executeCommandOnNodes 在多个节点上执行命令的通用方法
func (s *BatchOperationService) executeCommandOnNodes(nodeIDs []uint, command, description string) []BatchOperationResult {
	results := make([]BatchOperationResult, 0)

	for _, nodeID := range nodeIDs {
		result := BatchOperationResult{
			NodeID: nodeID,
		}

		node, err := s.getNodeByID(nodeID)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Name = fmt.Sprintf("Node_%d", nodeID)
			results = append(results, result)
			continue
		}

		result.Name = node.Name

		output, err := node.ExecuteCommand(command)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Output = output
		} else {
			result.Success = true
			result.Output = output
		}

		results = append(results, result)
	}

	return results
}

// getNodeByID 根据ID获取节点
func (s *BatchOperationService) getNodeByID(id uint) (Node, error) {
	var node Node
	return node.Load(id)
}

// scpUpload 使用SCP上传文件
func (s *BatchOperationService) scpUpload(client *ssh.Client, localPath, remotePath string) error {
	// 检查本地文件是否存在
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// 使用SCP协议上传文件
	// 格式: scp -t remotePath
	command := fmt.Sprintf("scp -t %s", remotePath)

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	err = session.Start(command)
	if err != nil {
		return err
	}

	// 等待初始响应
	response := make([]byte, 1)
	_, err = stdout.Read(response)
	if err != nil {
		return err
	}
	if response[0] != 0 {
		return fmt.Errorf("SCP初始化失败")
	}

	// 发送文件头信息
	fileName := filepath.Base(localPath)
	header := fmt.Sprintf("C0644 %d %s\n", fileInfo.Size(), fileName)
	_, err = stdin.Write([]byte(header))
	if err != nil {
		return err
	}

	// 等待确认
	_, err = stdout.Read(response)
	if err != nil {
		return err
	}
	if response[0] != 0 {
		return fmt.Errorf("SCP头信息发送失败")
	}

	// 发送文件内容
	_, err = io.Copy(stdin, file)
	if err != nil {
		return err
	}

	// 发送结束标记
	_, err = stdin.Write([]byte{0})
	if err != nil {
		return err
	}

	// 等待最终确认
	_, err = stdout.Read(response)
	if err != nil {
		return err
	}
	if response[0] != 0 {
		return fmt.Errorf("SCP文件传输失败")
	}

	stdin.Close()
	return session.Wait()
}

// generateFirewallCommandForNode 根据节点的包管理类型生成防火墙命令
func (s *BatchOperationService) generateFirewallCommandForNode(req FirewallRuleRequest, packageManager string) string {
	// 根据包管理类型决定使用哪种防火墙工具
	if packageManager == "deb" {
		// Debian/Ubuntu 系统使用 ufw
		return s.generateUfwCommand(req)
	} else if packageManager == "rpm" {
		// RPM 系统使用 firewall-cmd
		return s.generateFirewallCmdCommand(req)
	} else {
		// 其他系统使用智能检测
		return s.generateFirewallCommand(req)
	}
}

// generateUfwCommand 生成 ufw 防火墙命令
func (s *BatchOperationService) generateUfwCommand(req FirewallRuleRequest) string {
	var commands []string

	switch req.Action {
	case "add_port":
		for _, port := range req.Ports {
			commands = append(commands, fmt.Sprintf("sudo ufw allow %d", port))
		}
	case "remove_port":
		for _, port := range req.Ports {
			commands = append(commands, fmt.Sprintf("sudo ufw delete allow %d", port))
		}
	case "add_ip":
		for _, ip := range req.IPs {
			commands = append(commands, fmt.Sprintf("sudo ufw allow from %s", ip))
		}
	case "remove_ip":
		for _, ip := range req.IPs {
			commands = append(commands, fmt.Sprintf("sudo ufw delete allow from %s", ip))
		}
	}

	commandStr := strings.Join(commands, " && ")

	return fmt.Sprintf(`
# 检查并执行ufw防火墙操作
if command -v ufw >/dev/null 2>&1; then
	# 检查ufw是否已启用
	if sudo ufw status | grep -q "Status: active"; then
		%s
		echo "使用ufw执行防火墙操作成功"
	else
		echo "错误: ufw防火墙未启用，请先执行 'sudo ufw enable'"
		exit 1
	fi
else
	echo "错误: 未找到ufw防火墙工具，请先安装ufw"
	exit 1
fi
	`, commandStr)
}

// generateFirewallCmdCommand 生成 firewall-cmd 防火墙命令
func (s *BatchOperationService) generateFirewallCmdCommand(req FirewallRuleRequest) string {
	var commands []string

	switch req.Action {
	case "add_port":
		for _, port := range req.Ports {
			commands = append(commands, fmt.Sprintf("sudo firewall-cmd --permanent --add-port=%d/tcp && sudo firewall-cmd --permanent --add-port=%d/udp", port, port))
		}
	case "remove_port":
		for _, port := range req.Ports {
			commands = append(commands, fmt.Sprintf("sudo firewall-cmd --permanent --remove-port=%d/tcp && sudo firewall-cmd --permanent --remove-port=%d/udp", port, port))
		}
	case "add_ip":
		for _, ip := range req.IPs {
			commands = append(commands, fmt.Sprintf("sudo firewall-cmd --permanent --add-source=%s", ip))
		}
	case "remove_ip":
		for _, ip := range req.IPs {
			commands = append(commands, fmt.Sprintf("sudo firewall-cmd --permanent --remove-source=%s", ip))
		}
	}

	commandStr := strings.Join(commands, " && ")
	// 添加重新加载防火墙规则
	if len(commands) > 0 {
		commandStr += " && sudo firewall-cmd --reload"
	}

	return fmt.Sprintf(`
# 检查并执行firewall-cmd防火墙操作
if command -v firewall-cmd >/dev/null 2>&1; then
	# 检查firewalld是否在运行
	if sudo systemctl is-active firewalld >/dev/null 2>&1; then
		%s
		echo "使用firewall-cmd执行防火墙操作成功"
	else
		echo "错误: firewalld服务未运行，请先执行 'sudo systemctl start firewalld'"
		exit 1
	fi
else
	echo "错误: 未找到firewall-cmd防火墙工具，请先安装firewalld"
	exit 1
fi
	`, commandStr)
}

// generateFirewallCommand 生成智能检测防火墙系统的命令
func (s *BatchOperationService) generateFirewallCommand(req FirewallRuleRequest) string {
	var ufwCommands []string
	var firewallCmdCommands []string

	// 构建ufw命令
	switch req.Action {
	case "add_port":
		for _, port := range req.Ports {
			ufwCommands = append(ufwCommands, fmt.Sprintf("sudo ufw allow %d", port))
			firewallCmdCommands = append(firewallCmdCommands, fmt.Sprintf("sudo firewall-cmd --permanent --add-port=%d/tcp && sudo firewall-cmd --permanent --add-port=%d/udp", port, port))
		}
	case "remove_port":
		for _, port := range req.Ports {
			ufwCommands = append(ufwCommands, fmt.Sprintf("sudo ufw delete allow %d", port))
			firewallCmdCommands = append(firewallCmdCommands, fmt.Sprintf("sudo firewall-cmd --permanent --remove-port=%d/tcp && sudo firewall-cmd --permanent --remove-port=%d/udp", port, port))
		}
	case "add_ip":
		for _, ip := range req.IPs {
			ufwCommands = append(ufwCommands, fmt.Sprintf("sudo ufw allow from %s", ip))
			firewallCmdCommands = append(firewallCmdCommands, fmt.Sprintf("sudo firewall-cmd --permanent --add-source=%s", ip))
		}
	case "remove_ip":
		for _, ip := range req.IPs {
			ufwCommands = append(ufwCommands, fmt.Sprintf("sudo ufw delete allow from %s", ip))
			firewallCmdCommands = append(firewallCmdCommands, fmt.Sprintf("sudo firewall-cmd --permanent --remove-source=%s", ip))
		}
	}

	ufwCommandStr := strings.Join(ufwCommands, " && ")
	firewallCmdCommandStr := strings.Join(firewallCmdCommands, " && ")

	// 添加重新加载防火墙规则
	if len(firewallCmdCommands) > 0 {
		firewallCmdCommandStr += " && sudo firewall-cmd --reload"
	}

	// 智能检测并执行防火墙命令的脚本
	return fmt.Sprintf(`
# 检测并执行防火墙操作
if command -v ufw >/dev/null 2>&1; then
	# 检查ufw是否已启用
	if sudo ufw status | grep -q "Status: active"; then
		%s
		echo "使用ufw执行防火墙操作成功"
	else
		echo "错误: ufw防火墙未启用，请先执行 'sudo ufw enable'"
		exit 1
	fi
elif command -v firewall-cmd >/dev/null 2>&1; then
	# 检查firewalld是否在运行
	if sudo systemctl is-active firewalld >/dev/null 2>&1; then
		%s
		echo "使用firewall-cmd执行防火墙操作成功"
	else
		echo "错误: firewalld服务未运行，请先执行 'sudo systemctl start firewalld'"
		exit 1
	fi
else
	echo "错误: 未找到支持的防火墙工具(ufw或firewall-cmd)，请先安装相应的防火墙软件"
	exit 1
fi
	`, ufwCommandStr, firewallCmdCommandStr)
}

// ScanDockerImages 扫描目录中的Docker镜像文件
func (s *BatchOperationService) ScanDockerImages(directoryPath string) []string {
	var imageFiles []string

	// 判断是文件还是目录
	if fileInfo, err := os.Stat(directoryPath); err == nil && !fileInfo.IsDir() {
		// 如果是文件，直接检查是否是.tar文件
		if strings.HasSuffix(strings.ToLower(fileInfo.Name()), ".tar") {
			return []string{directoryPath}
		}
		return imageFiles
	}

	// 检查目录是否存在
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		return imageFiles
	}

	// 读取目录中的文件
	entries, err := os.ReadDir(directoryPath)
	if err != nil {
		return imageFiles
	}

	// 查找.tar文件（Docker镜像文件）
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".tar") {
			imageFiles = append(imageFiles, entry.Name())
		}
	}

	return imageFiles
}

// DeployDockerImagesFromDirectory 批量部署目录中的Docker镜像
func (s *BatchOperationService) DeployDockerImagesFromDirectory(req DockerImagesFromDirectoryRequest) []BatchOperationResult {
	results := make([]BatchOperationResult, 0)

	for _, nodeID := range req.NodeIDs {
		result := BatchOperationResult{
			NodeID: nodeID,
		}

		node, err := s.getNodeByID(nodeID)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Name = fmt.Sprintf("Node_%d", nodeID)
			results = append(results, result)
			continue
		}

		result.Name = node.Name

		// 建立SSH连接
		client, err := node.GetSSHClient()
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("SSH连接失败: %v", err)
			results = append(results, result)
			continue
		}
		defer client.Close()

		var successCount int
		var errors []string
		var outputs []string

		// 扫描目录中的Docker镜像文件
		var imageFiles = s.ScanDockerImages(req.DirectoryPath)
		// 逐个处理每个镜像文件
		for _, imageFile := range imageFiles {
			localPath := imageFile
			// 从localppath中获取文件名
			if !filepath.IsAbs(imageFile) {
				localPath = filepath.Join(req.DirectoryPath, imageFile)
			}
			remotePath := fmt.Sprintf("/tmp/%s", filepath.Base(imageFile))
			fileIsExist := false
			// 判断文件是否存在，判断文件md5是否一致，一致则跳过上传
			if _, err := os.Stat(localPath); err == nil {
				// 获取本地文件的MD5值
				localMD5, err := s.calculateFileMD5(localPath)
				logger.LOG.Debugf("本地文件MD5值: %s, 错误: %v", localMD5, err)
				if err == nil {
					// 检查远程文件是否存在并获取其MD5值
					cmdstr := fmt.Sprintf("if [ -f %s ]; then md5sum %s | cut -d' ' -f1; else echo 'file_not_exist'; fi", remotePath, remotePath)

					logger.LOG.Debugf("cmdstr: %s", cmdstr)
					md5Output, err := node.ExecuteCommand(cmdstr)
					if err == nil {
						remoteMD5 := strings.TrimSpace(md5Output)
						logger.LOG.Debugf("远程文件MD5值: %s, 错误: %v", remoteMD5, err)
						// 如果远程文件存在且MD5相同，则跳过上传
						if remoteMD5 != "file_not_exist" && remoteMD5 == localMD5 {
							fileIsExist = true
							outputs = append(outputs, fmt.Sprintf("文件 %s 已存在且内容相同，跳过上传", imageFile))
						}
					}
				}
			}

			if !fileIsExist {
				// 上传Docker镜像文件
				err = s.scpUpload(client, localPath, remotePath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("上传 %s 失败: %v", imageFile, err))
					continue
				}
				outputs = append(outputs, fmt.Sprintf("成功上传 %s", imageFile))
			}

			// 加载Docker镜像
			command := fmt.Sprintf("docker load -i %s", remotePath)
			output, err := node.ExecuteCommand(command)
			if err != nil {
				errors = append(errors, fmt.Sprintf("加载 %s 失败: %v", imageFile, err))
			} else {
				successCount++
				outputs = append(outputs, fmt.Sprintf("成功加载 %s: %s", imageFile, output))
				// 清理临时文件
				cleanupCommand := fmt.Sprintf("rm -f %s", remotePath)
				node.ExecuteCommand(cleanupCommand)
			}

		}

		// 设置结果
		if successCount == len(imageFiles) {
			result.Success = true
			result.Output = strings.Join(outputs, "\n")
		} else if successCount > 0 {
			result.Success = false
			result.Output = strings.Join(outputs, "\n")
			result.Error = fmt.Sprintf("部分成功：%d/%d，失败原因：%s", successCount, len(imageFiles), strings.Join(errors, "; "))
		} else {
			result.Success = false
			result.Error = strings.Join(errors, "; ")
		}

		results = append(results, result)
	}

	return results
}

// FileItem 文件项
type FileItem struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size,omitempty"`
	Path  string `json:"path"`
}

// ListFiles 列出指定路径下的文件和目录
func (s *BatchOperationService) ListFiles(path string) []FileItem {
	var files []FileItem

	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return files
	}

	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return files
	}

	// 处理每个条目
	for _, entry := range entries {
		itemPath := filepath.Join(path, entry.Name())

		// 使用 os.Stat 而不是 entry.Info() 来正确处理符号链接
		// os.Stat 会跟随符号链接获取目标文件的信息
		info, err := os.Stat(itemPath)
		if err != nil {
			// 如果符号链接指向不存在的文件，跳过该条目
			continue
		}

		item := FileItem{
			Name:  entry.Name(),
			IsDir: info.IsDir(), // 使用真实文件信息判断是否为目录
			Path:  itemPath,
		}

		if !info.IsDir() {
			item.Size = info.Size()
		}

		files = append(files, item)
	}

	return files
}

// calculateFileMD5 计算文件的MD5值
func (s *BatchOperationService) calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
