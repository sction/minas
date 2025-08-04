package route

import (
	"fmt"
	"server/app/basic/projectdir"
	"server/app/basic/system"
	"server/app/basic/user"
	"server/app/nas/external"
	"server/app/nas/webdav"
	"server/app/schtask"
	"server/app/sflow"
	"server/app/term"
	"server/dagflow/api"
	"server/middleware"
	"server/utils/config"

	"github.com/gin-gonic/gin"
)

// Init 初始化系统路由
// 此函数负责设置所有API路由和对应的处理函数，组织整个应用程序的API结构
// 参数:
//   - app: Gin引擎实例，用于注册路由
func Init(app *gin.Engine) {
	// 输出应用上下文路径信息
	fmt.Println("----------------------:" + config.CONF.App.ContextPath)
	// 创建API v1版本的路由组，所有API都在此路由组下
	v1 := app.Group(config.CONF.App.ContextPath + "/v1")
	{
		// 白名单路由组，这些路由无需验证令牌即可访问
		WhiteList := v1.Group("")
		{
			// 验证码获取接口
			WhiteList.GET("/get_captcha", middleware.GetCaptchaHandler)
			// 用户登录接口
			WhiteList.POST("/login", middleware.LoginHandler)
			// 用户登出接口
			WhiteList.GET("/password_login_out", middleware.PasswordLoginOut)
			// 获取用户信息接口
			WhiteList.GET("/userinfo", middleware.GetUserInfo)
			// 获取一次性令牌接口，用于安全验证
			WhiteList.GET("/once/token", middleware.GetOnceToken)
			// 系统相关接口子路由组
			System := WhiteList.Group("/system")
			{
				// 获取系统版本信息
				System.GET("/version", system.SystemVersion)
				// 检查系统状态
				System.GET("/check-state", system.SystemCheckState)
				// 系统初始化接口
				System.POST("/init", system.SystemInit)
				// 仪表板统计相关接口
				System.GET("/dashboard/stats", system.GetDashboardStats)
				System.GET("/dashboard/task-types", system.GetTaskTypeStats)
				System.GET("/dashboard/task-execution", system.GetTaskExecutionStats)
				System.GET("/dashboard/task-trend", system.GetTaskTrend)
				System.GET("/dashboard/task-projects", system.GetTaskProjectStats)
			}
		}

		// 基础模块路由组，需要认证中间件保护
		Basic := v1.Group("/basic", middleware.AuthMiddleware)
		{
			// 添加用户管理相关路由
			user.AddRoutes(Basic)
			println("------------------")
			// 添加项目目录相关路由
			projectdir.AddRoutes(Basic)
		}

		// 终端管理路由组
		Term := v1.Group("/term")
		{
			// 添加终端管理相关路由
			term.AddRoutes(Term)
		}

		// 计划任务路由组，需要认证中间件保护
		SchSystem := v1.Group("/sch", middleware.AuthMiddleware)
		{
			// 添加计划任务相关路由
			schtask.SchTaskApp{}.AddRoutes(SchSystem)
			// 添加计划任务日志相关路由
			schtask.SchLogApp{}.AddRoutes(SchSystem)
		}

		// 作业流程路由组，需要认证中间件保护
		SFlowSystem := v1.Group("/sflow", middleware.AuthMiddleware)
		{
			// 添加作业流程相关路由
			sflow.SFlowApp{}.AddRoutes(SFlowSystem)
			sflow.SFlowLogApp{}.AddRoutes(SFlowSystem)
		}

		// DAG流程路由组，需要认证中间件保护
		RagFlowSystem := v1.Group("/dagflow", middleware.AuthMiddleware)
		{
			// 添加DAG流程相关路由
			ragFlowAPI := &api.DAGFlowAPI{}
			ragFlowAPI.AddRoutes(RagFlowSystem)
		}

		// 文件存储(NAS)路由组，需要认证中间件保护
		NasSystem := v1.Group("/nas", middleware.AuthMiddleware)
		{
			// 添加WebDAV相关路由，用于文件访问
			webdav.AddRoutes(NasSystem)
			// 添加外部存储相关路由
			external.AddRoutes(NasSystem)
		}

		// 配置管理路由组，通过API密钥认证保护
		Config := v1.Group("/config", middleware.AuthApiKeyMiddleware)
		{
			// 启用UI接口
			Config.GET("/ui/enable", system.EnableUI)
			// 禁用UI接口
			Config.GET("/ui/disable", system.DisableUI)
		}
	}
}
