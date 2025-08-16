package system

import (
	"os"
	"os/exec"
	"strings"
)

// IsRunningInContainer 检测是否在Docker容器内运行
func IsRunningInContainer() bool {
	// 方法1: 检查/.dockerenv文件（Docker容器特有）
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// 方法2: 检查/proc/1/cgroup文件内容
	if content, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		contentStr := string(content)
		// 如果包含docker或containerd相关路径，说明在容器内
		if strings.Contains(contentStr, "docker") ||
			strings.Contains(contentStr, "containerd") ||
			strings.Contains(contentStr, "/system.slice/docker") {
			return true
		}
	}

	// 方法3: 检查容器运行时环境变量
	if os.Getenv("container") != "" ||
		os.Getenv("DOCKER_CONTAINER") != "" {
		return true
	}

	// 方法4: 检查init进程
	if content, err := os.ReadFile("/proc/1/comm"); err == nil {
		init := strings.TrimSpace(string(content))
		// 如果init进程不是systemd或init，可能在容器内
		if init != "systemd" && init != "init" {
			return true
		}
	}

	return false
}

// IsNfsServiceAvailable 检查NFS服务功能是否可用
// 在容器内运行时返回false，禁用NFS服务管理功能
func IsNfsServiceAvailable() bool {
	// 如果在容器内运行，禁用NFS服务管理
	if IsRunningInContainer() {
		return false
	}

	// 检查systemctl命令是否可用
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false
	}

	// 检查是否有基本的NFS工具
	if _, err := exec.LookPath("exportfs"); err != nil {
		return false
	}

	return true
}
