// package cmd 包含服务器命令行相关功能
package cmd

import (
	"log"
	"strings"

	"github.com/fsnotify/fsnotify" // 用于监控文件变化
	"github.com/spf13/cobra"       // 命令行工具库
	"github.com/spf13/viper"       // 配置管理库

	"crypto/tls"               // 添加TLS配置支持
	"io/fs"                    // 文件系统接口
	"net/http"                 // HTTP服务
	"os"                       // 操作系统功能
	"path/filepath"            // 文件路径处理
	"server/dagflow"           // DAGFlow服务
	"server/data"              // 数据相关
	"server/middleware"        // 中间件
	"server/route"             // 路由定义
	"server/service/database"  // 数据库服务
	"server/service/scheduled" // 定时任务服务
	"server/utils/config"      // 配置工具
	"server/utils/logger"      // 日志工具
	"server/utils/rclone"      // RClone工具
	"server/utils/tlscert"     // TLS证书工具
	"server/www"               // 前端资源

	"github.com/gin-gonic/gin" // Web框架
)

// CERT_CRT_PATH 证书文件路径
const CERT_CRT_PATH = "data/certs/server.crt"

// CERT_KEY_PATH 证书密钥文件路径
const CERT_KEY_PATH = "data/certs/server.key"

// serverCmd 表示服务器命令，用于启动Web服务
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "服务",
	Long:  `启动服务.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 在运行命令前初始化配置
		InitConfig(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 启动服务器
		StartServer()
	},
}

// init 初始化函数，在包被导入时自动执行
func init() {
	// 将server命令添加到根命令
	rootCmd.AddCommand(serverCmd)
	// 接收参数port（已注释掉，现在使用配置文件中的端口）

	// serverCmd.Flags().UintVarP(&httpsPort, "port", "p", 8003, "https端口号,默认8003")
	// serverCmd.Flags().UintVar(&httpPort, "http", 8002, "开启http协议端口，默认8002，0不开启")
}

// fixConfigFile 检查并创建配置文件（如果不存在）
func fixConfigFile() {
	// 检查config.yaml文件是否存在
	_, err := os.Stat(data.ConfigFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建父目录
			parentDir := filepath.Dir(data.ConfigFilePath)
			os.MkdirAll(parentDir, 0755)
			// 创建配置文件并写入默认配置
			f, err := os.Create(data.ConfigFilePath)
			if err == nil {
				f.Write(data.DefaultYamlData)
			}
		}
	}
}

// InitConfig 初始化配置系统
func InitConfig(cmd *cobra.Command) {

	// 如果配置文件不存在则创建
	fixConfigFile()

	// 设置环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MINAS") // 环境变量前缀
	// 设置环境变量名称转换规则（将.和-转换为_）
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	// 绑定命令行参数
	if cmd != nil {
		viper.BindPFlags(cmd.Flags())
	}
	// 设置配置文件路径
	viper.SetConfigFile(data.ConfigFilePath)

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("读取配置文件出错:  %v", err)
		os.Exit(1)
	}
	// 将配置映射到结构体
	err = viper.Unmarshal(&config.CONF)
	if err != nil {
		log.Fatalf("解析config.yaml出错: %v", err)
	}
	// 设置默认配置值
	initFixConfig()

	// 开始监听配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件变化时重新加载配置
		err = viper.Unmarshal(&config.CONF)
		if err != nil {
			log.Fatalf("解析config.yaml出错: %v", err)
		}
		initFixConfig()
	})

	// 输出使用的配置文件路径
	path, err := filepath.Abs(viper.ConfigFileUsed())
	if err != nil {
		path = viper.ConfigFileUsed()
	}
	log.Printf("使用配置文件: %s\n", path)
}

// initFixConfig 设置配置的默认值和修正值
func initFixConfig() {
	// 如果RClone命令超时时间未设置，则设置为24小时
	if config.CONF.RClone.CmdTimeOut == 0 {
		config.CONF.RClone.CmdTimeOut = 24
	}
	// 输出当前配置信息
	log.Printf("当前配置: %v\n", config.CONF)
}

// StartServer 启动Web服务器
func StartServer() {

	// 根据配置设置Gin的运行模式
	if config.CONF.Debug {
		// 开发模式 - 输出详细日志
		gin.SetMode(gin.DebugMode)
	} else {
		// 生产模式 - 优化性能
		gin.SetMode(gin.ReleaseMode)
	}
	// 初始化日志系统
	logger.Init()

	// 启动RClone服务（用于文件操作）
	pid, err := rclone.StartRClone()
	if err != nil {
		logger.LOG.Errorf("启动RClone错误：%s", err.Error())
		logger.LOG.Errorf("文件备份、文件浏览、远程存储等功能将无法正常使用！")
	} else {
		logger.LOG.Infof("RClone启动成功！ PID：%d", pid)
	}

	// 初始化数据库
	database.Init()
	// 启动定时任务
	scheduled.Start()
	if err != nil {
		logger.LOG.Errorf("启动定时任务错误：%s", err.Error())
	}

	// 初始化DAGFlow服务
	initDAGFlow()

	// 初始化Gin应用
	app := gin.Default()
	logger.LOG.Infof("Gin框架初始化成功，调试模式：%v", config.CONF.Debug)

	// 设置最大上传文件大小
	app.MaxMultipartMemory = config.CONF.FileUpload.MaxSize

	// 初始化WebDAV处理器（必须在其他中间件前注册，防止其他中间件影响WebDAV功能）
	app.Use(middleware.WebDavHandler())

	// 添加跨域支持中间件
	app.Use(middleware.Cors())

	// 初始化后端API路由
	route.Init(app)
	// 注册前端Vue路由
	registerVueRoutes(app)
	// 添加统一响应中间件（处理未明确处理的响应）
	app.Use(middleware.UnifiedResponseMiddleware())

	// 异常处理中间件（防止业务中断）
	// app.Use(gin.Recovery())

	// 根据配置启动HTTP和/或HTTPS服务
	if config.CONF.App.Port != "" && config.CONF.App.SslPort != "" {
		// 同时启动HTTP和HTTPS服务
		go httpsServer(app) // HTTPS在单独的goroutine中运行
		httpServer(app)     // HTTP在主goroutine中运行
	} else if config.CONF.App.Port != "" {
		// 仅启动HTTP服务
		httpServer(app)
	} else if config.CONF.App.SslPort != "" {
		// 仅启动HTTPS服务
		httpsServer(app)
	} else {
		// 配置错误，未指定端口
		logger.LOG.Error("配置文件中未设置HTTP或HTTPS端口")
	}
}

// registerVueRoutes 注册前端Vue应用的静态资源和路由
func registerVueRoutes(app *gin.Engine) {
	// 获取嵌入的静态资源目录
	staticDir, _ := fs.Sub(www.StaticFs, "dist/static")
	// 注册静态资源路由
	app.StaticFS(config.CONF.App.ContextPath+"/static", http.FS(staticDir))

	// 处理所有未匹配的路由（SPA应用需要）
	app.NoRoute(func(c *gin.Context) {
		if config.CONF.App.EnableUI {
			// UI启用时，返回主页面
			c.Writer.WriteHeader(http.StatusOK)
			_, _ = c.Writer.Write(www.IndexByte)
			c.Writer.Header().Add("Accept", "text/html")
			c.Writer.Flush()
		} else {
			// UI禁用时，返回禁用页面
			c.Writer.WriteHeader(http.StatusOK)
			_, _ = c.Writer.Write(www.NoUIndexByte)
			c.Writer.Header().Add("Accept", "text/html")
			c.Writer.Flush()
		}
	})
}

// httpServer 启动HTTP服务器
func httpServer(app *gin.Engine) {
	// 使用配置的主机和端口启动HTTP服务
	if err := app.Run(config.CONF.App.Host + ":" + config.CONF.App.Port); err != nil {
		panic(err) // 启动失败时触发panic
	}
}

// httpsServer 启动HTTPS服务器
func httpsServer(app *gin.Engine) {
	// 检查证书文件是否存在，如果不存在则自动生成自签名证书
	_, err := os.Stat(CERT_CRT_PATH)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建证书目录
			parentDir := filepath.Dir(CERT_CRT_PATH)
			os.MkdirAll(parentDir, 0755)
			// 生成自签名证书
			tlscert.GenerateSelfSignedCert(CERT_CRT_PATH, CERT_KEY_PATH, "", "")
		}
	}

	// 创建安全的TLS配置以修复CVE-2016-2183漏洞
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // 最低使用TLS 1.2
		MaxVersion: tls.VersionTLS13, // 最高使用TLS 1.3
		CipherSuites: []uint16{
			// TLS 1.2 安全密码套件
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true, // 优先使用服务器密码套件顺序
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}

	// 创建HTTP服务器并配置TLS
	server := &http.Server{
		Addr:      config.CONF.App.Host + ":" + config.CONF.App.SslPort,
		Handler:   app,
		TLSConfig: tlsConfig,
	}

	// 启动HTTPS服务
	if err := server.ListenAndServeTLS(CERT_CRT_PATH, CERT_KEY_PATH); err != nil {
		panic(err) // 启动失败时触发panic
	}
}

// initDAGFlow 初始化DAGFlow服务
func initDAGFlow() {
	// 初始化DAGFlow服务
	dagflow.Init()
	logger.LOG.Infof("DAGFlow服务初始化成功")
}
