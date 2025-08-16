package system

import (
	"runtime"
	"server/core/app/response"
	"server/service/basic"
	"server/utils/config"
	"server/utils/data"
	"server/utils/simple"
	"server/utils/system"

	"github.com/gin-gonic/gin"
)

// SystemCheckState 检查系统初始化状态
// 用于检测系统是否为首次安装（没有用户记录）
// 参数 ctx: Gin上下文
func SystemCheckState(ctx *gin.Context) {
	user := basic.User{}
	count, err := user.Count()
	if err == nil {
		response.Data(ctx, "", data.Map{"fresh": count == 0})
	} else {
		response.Error(ctx, simple.NewSimpleErrorMessage("读取系统状态错误！"))
		return
	}
}

// SystemInit 系统初始化
// 用于系统首次安装时创建管理员账号
// 参数 ctx: Gin上下文
func SystemInit(ctx *gin.Context) {
	user := basic.User{}
	count, err := user.Count()
	if err == nil {
		if count == 0 {
			user := basic.User{}
			if err := ctx.BindJSON(&user); err != nil {
				response.BadRequest(ctx, "参数错误！")
				return
			}
			user.IsAdmin = 1
			user.SetOperator(&user, ctx)
			err := user.Create(&user)
			if err == nil {
				response.Success(ctx, "初始化成功！")
			} else {
				response.Error(ctx, simple.NewSimpleErrorMessage("初始化失败！"))
			}
		} else {
			response.Error(ctx, simple.NewSimpleErrorMessage("系统已初始化完成！"))
		}
	} else {
		response.Error(ctx, simple.NewSimpleErrorMessage("初始化前读取系统状态错误！"))
	}
}

// EnableUI 启用UI
// 开启系统的Web界面功能
// 参数 ctx: Gin上下文
func EnableUI(ctx *gin.Context) {
	config.CONF.App.EnableUI = true
}

// DisableUI 禁用UI
// 关闭系统的Web界面功能
// 参数 ctx: Gin上下文
func DisableUI(ctx *gin.Context) {
	config.CONF.App.EnableUI = false
}

// GetSystemEnvironment 获取系统环境信息
func GetSystemEnvironment(ctx *gin.Context) {
	envInfo := data.Map{
		"isContainer":         system.IsRunningInContainer(),
		"nfsServiceAvailable": system.IsNfsServiceAvailable(),
		"os":                  runtime.GOOS,
		"arch":                runtime.GOARCH,
		"goVersion":           runtime.Version(),
		"version":             config.Version,
	}

	// 添加容器相关信息
	if system.IsRunningInContainer() {
		envInfo["containerType"] = "docker"
		envInfo["message"] = "系统运行在容器环境中，某些系统级功能可能不可用"
	}

	response.Data(ctx, "", envInfo)
}
