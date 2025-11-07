package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"server/utils"
	"strings"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

// Program 结构体定义了服务程序
// exit 通道用于通知服务停止
type Program struct {
	exit chan struct{}
}

// Start 方法在服务启动时调用
// 参数 s 是服务实例
// 返回启动过程中的错误，如果没有错误则返回 nil
func (p *Program) Start(s service.Service) error {
	log.Println("启动服务...", service.Platform())
	p.exit = make(chan struct{}) // 初始化退出通道
	go p.run()                   // 在新的协程中运行服务主体逻辑
	log.Println("服务启动完成.")
	return nil
}

// run 方法包含服务的主要业务逻辑
func (p *Program) run() {
	log.Println("服务运行中...")
	InitConfig(nil) // 初始化配置
	StartServer()   // 启动服务器
}

// Stop 方法在服务停止时调用
// 参数 s 是服务实例
// 返回停止过程中的错误，如果没有错误则返回 nil
func (p *Program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	log.Println("停止服务.")
	close(p.exit) // 关闭退出通道，通知各个协程停止运行
	return nil
}

// init 函数在包初始化时执行，用于添加命令到根命令
func init() {
	// 添加命令
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(runCmd)
}

// runCmd 定义了运行服务的命令
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "启动服务",
	Long:  `启动服务.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 命令执行前的准备工作
	},
	Run: func(cmd *cobra.Command, args []string) {
		ControlService("run") // 控制服务执行运行操作
	},
}

// installCmd 定义了安装服务的命令
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装服务",
	Long:  `安装服务.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 命令执行前的准备工作
		if runtime.GOOS == "linux" {
			// 检查并配置 SELinux
			configureSELinux()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ControlService("install") // 控制服务执行安装操作
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		// 命令执行后的一些清理工作
		log.Println("安装服务完成，请使用 'minas start' 启动服务。")
		if runtime.GOOS == "linux" {
			// 检查并配置防火墙
			configureFirewall()
			log.Println("请确保防火墙规则已正确配置。")
		}
	},
}

// uninstallCmd 定义了卸载服务的命令
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载服务",
	Long:  `卸载服务.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 命令执行前的准备工作
	},
	Run: func(cmd *cobra.Command, args []string) {
		ControlService("uninstall") // 控制服务执行卸载操作
	},
}

// startCmd 定义了启动服务的命令
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动服务",
	Long:  `启动服务.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 命令执行前的准备工作
	},
	Run: func(cmd *cobra.Command, args []string) {
		ControlService("start") // 控制服务执行启动操作
	},
}

// stopCmd 定义了停止服务的命令
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止服务",
	Long:  `停止服务.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 命令执行前的准备工作
	},
	Run: func(cmd *cobra.Command, args []string) {
		ControlService("stop") // 控制服务执行停止操作
	},
}

// ControlService 根据指定的操作控制服务
// 参数 action 表示要执行的操作：run, install, uninstall, start, stop
func ControlService(action string) {
	log.Println("正在执行服务操作: " + action + "")
	s, err := initService() // 初始化服务
	if err != nil {
		log.Println("初始化服务失败：" + err.Error())
	}
	switch action {
	case "run":
		s.Run() // 直接运行服务
	default:
		// 使用 service 包的 Control 函数执行其他操作
		err := service.Control(s, action)
		if err != nil {
			log.Printf("%s not valid, Valid actions: %q\n%v", action, service.ControlAction, err)
		}
		return
	}
}

// initService 初始化并返回服务实例
// 返回服务实例和可能的错误
// configureFirewall 检查并配置防火墙规则
// configureSELinux 检查并配置 SELinux 权限
func configureSELinux() {
	// 检查是否安装了 SELinux 工具
	if _, err := os.Stat("/usr/sbin/sestatus"); err != nil {
		log.Println("未检测到 SELinux 工具，跳过 SELinux 配置")
		return
	}

	// 检查 SELinux 状态
	cmd := exec.Command("getenforce")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("获取 SELinux 状态失败: %v\n", err)
		return
	}

	status := strings.TrimSpace(string(output))
	if strings.ToLower(status) == "disabled" {
		log.Println("SELinux 已禁用，无需配置")
		return
	}

	log.Printf("检测到 SELinux 状态为: %s，正在配置程序权限...\n", status)

	// 获取程序路径
	exePath := utils.GetExeFileDirectory()
	if exePath == "" {
		log.Println("无法获取程序路径，跳过 SELinux 配置")
		return
	}

	// 设置可执行文件的 SELinux 上下文
	cmd = exec.Command("chcon", "-t", "bin_t", exePath+"/minas")
	if err := cmd.Run(); err != nil {
		log.Printf("设置程序 SELinux 上下文失败: %v\n", err)
	} else {
		log.Println("已设置程序 SELinux 上下文")
	}
}

func configureFirewall() {
	// 获取需要配置的端口
	httpPort := "8002"
	httpsPort := "8003"

	// 尝试配置 firewalld
	if configureFirewalld(httpPort, httpsPort) {
		return
	}

	// 如果 firewalld 未安装或未运行，尝试配置 ufw
	if configureUfw(httpPort, httpsPort) {
		return
	}

	log.Println("未检测到支持的防火墙服务（firewalld/ufw），无需配置防火墙规则")
}

// configureFirewalld 配置 firewalld 防火墙规则
func configureFirewalld(httpPort, httpsPort string) bool {
	// 检查是否安装了 firewalld
	if _, err := os.Stat("/usr/bin/firewall-cmd"); err == nil {
		// 检查 firewalld 是否运行
		cmd := exec.Command("systemctl", "is-active", "firewalld")
		if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) == "active" {
			log.Println("检测到 firewalld 正在运行，正在配置防火墙规则...")

			// 检查 firewall-cmd 是否可以执行
			testCmd := exec.Command("firewall-cmd", "--state")
			if err := testCmd.Run(); err != nil {
				log.Println("无法执行 firewall-cmd，可能需要以 root 权限运行")
				return false
			}
			// 添加 HTTP 端口
			httpPortRule := fmt.Sprintf("%s/tcp", httpPort)
			cmd = exec.Command("firewall-cmd", "--zone=public", "--add-port="+httpPortRule, "--permanent")
			if err := cmd.Run(); err != nil {
				log.Printf("添加 HTTP 端口失败: %v\n", err)
			} else {
				log.Printf("已添加 HTTP 端口 %s\n", httpPortRule)
			}

			// 添加 HTTPS 端口
			httpsPortRule := fmt.Sprintf("%s/tcp", httpsPort)
			cmd = exec.Command("firewall-cmd", "--zone=public", "--add-port="+httpsPortRule, "--permanent")
			if err := cmd.Run(); err != nil {
				log.Printf("添加 HTTPS 端口失败: %v\n", err)
			} else {
				log.Printf("已添加 HTTPS 端口 %s\n", httpsPortRule)
			}

			// 重新加载防火墙规则
			cmd = exec.Command("firewall-cmd", "--reload")
			if err := cmd.Run(); err != nil {
				log.Printf("重新加载防火墙规则失败: %v\n", err)
			} else {
				log.Println("防火墙规则已更新")
			}
			return true
		}
	}
	return false
}

// configureUfw 配置 ufw 防火墙规则
func configureUfw(httpPort, httpsPort string) bool {
	// 检查是否安装了 ufw
	if _, err := os.Stat("/usr/bin/ufw"); err == nil {
		// 检查 ufw 是否启用
		cmd := exec.Command("ufw", "status")
		if output, err := cmd.Output(); err == nil && strings.Contains(string(output), "Status: active") {
			log.Println("检测到 ufw 正在运行，正在配置防火墙规则...")

			// 添加 HTTP 端口
			cmd = exec.Command("ufw", "allow", httpPort+"/tcp")
			if err := cmd.Run(); err != nil {
				log.Printf("添加 HTTP 端口失败: %v\n", err)
			} else {
				log.Printf("已添加 HTTP 端口 %s/tcp\n", httpPort)
			}

			// 添加 HTTPS 端口
			cmd = exec.Command("ufw", "allow", httpsPort+"/tcp")
			if err := cmd.Run(); err != nil {
				log.Printf("添加 HTTPS 端口失败: %v\n", err)
			} else {
				log.Printf("已添加 HTTPS 端口 %s/tcp\n", httpsPort)
			}

			log.Println("ufw 防火墙规则已更新")
			return true
		} else {
			log.Println("检测到 ufw 已安装但未启用")
		}
	}
	return false
}

func initService() (service.Service, error) {
	os.Chdir(utils.GetExeFileDirectory()) // 切换到可执行文件所在目录

	// 服务的配置信息
	options := make(service.KeyValue)
	options["LogOutput"] = true
	options["HasOutputFileSupport"] = true
	options["WorkingDirectory"] = utils.GetWorkDirectory()
	options["CurrentDirectory"] = utils.GetWorkDirectory()

	dependencies := []string{}
	if runtime.GOOS == "windows" {
		// Windows 系统特定配置
	} else {
		// 非 Windows 系统配置
		options["Restart"] = "on-failure"
		options["SuccessExitStatus"] = "1 2 8 SIGKILL"
		options["SELinuxContext"] = "unconfined_u:object_r:bin_t:s0"
	}

	// 服务基本配置
	svcConfig := &service.Config{
		Name:         "minas",
		DisplayName:  "Minas服务",
		Description:  "支持webdav、文件备份、日志或备份清理、超级终端管理",
		Option:       options,
		Dependencies: dependencies,
		Arguments:    []string{"run"}, // 运行时的默认参数
	}

	if runtime.GOOS == "windows" {
		// Windows 系统特定服务配置
	} else {
		// 非 Windows 系统特定服务配置
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"}
		svcConfig.UserName = "root" // 使用 root 用户运行服务
	}

	pro := &Program{}                     // 创建程序实例
	s, err := service.New(pro, svcConfig) // 创建服务
	return s, err
}
