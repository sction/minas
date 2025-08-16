//go:build linux
// +build linux

package samba

import (
	"fmt"
	"log"
	"os"
	"server/core/app/request"
	"server/core/app/response"
	"server/core/app/webapi"
	"server/service/nas"
	"server/utils/data"
	"server/utils/global"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// SambaApp Samba应用控制器
type SambaApp struct {
	webapi.BaseApp[nas.SambaShare] // 使用nas.SambaShare模型作为数据类型
}

// AddRoutes 注册Samba管理相关的路由
// 参数:
//   - parentGroup: 父路由组
func AddRoutes(parentGroup *gin.RouterGroup) {
	// 创建Samba路由组
	group := parentGroup.Group("/samba")

	// 创建Samba应用实例
	app := SambaApp{}

	// 设置允许更新的字段
	app.UpdateFields = []string{"name", "path", "comment", "browseable", "read_only", "guest_ok", "valid_users", "write_list", "read_list", "create_mask", "directory_mask", "force_user", "force_group", "available_space", "options", "remark", "is_disable"}

	// 服务管理路由
	group.GET("/service/status", app.GetServiceStatus) // 获取Samba服务状态
	group.POST("/service/start", app.StartService)     // 启动Samba服务
	group.POST("/service/stop", app.StopService)       // 停止Samba服务
	group.POST("/service/restart", app.RestartService) // 重启Samba服务
	group.POST("/service/enable", app.EnableService)   // 启用开机自启
	group.POST("/service/disable", app.DisableService) // 禁用开机自启
	group.POST("/reload-config", app.ReloadConfig)     // 重新加载配置
	group.GET("/list-dir", app.ListDir)                // 列出目录内容
	group.GET("/dir-owner", app.GetDirOwner)           // 获取目录所有者

	// 同步功能路由
	group.GET("/sync-status", app.GetSyncStatus)        // 获取同步状态
	group.POST("/sync-from-server", app.SyncFromServer) // 从服务器同步到数据库
	group.POST("/sync-to-server", app.SyncToServer)     // 从数据库同步到服务器

	// 配置管理路由
	group.GET("/smb-config", app.GetSmbConfig)     // 获取smb.conf配置文件内容
	group.POST("/backup-smb", app.BackupSmbConfig) // 备份smb.conf配置文件
	group.POST("/test-config", app.TestConfig)     // 测试配置

	// 用户管理路由
	group.GET("/users", app.ListSambaUsers)                    // 获取Samba用户列表
	group.POST("/users/add", app.AddSambaUser)                 // 添加Samba用户
	group.POST("/users/delete", app.DeleteSambaUser)           // 删除Samba用户
	group.POST("/users/password", app.ChangeSambaUserPassword) // 修改用户密码

	// 注册基本的CRUD路由
	webapi.AddBaseRoutes(group, &app)
}

// List 获取Samba共享列表
func (app *SambaApp) List(ctx *gin.Context) {
	log.Println("List samba shares")

	// 创建Samba共享数据库对象
	sambaShare := nas.SambaShare{}

	// 获取查询参数
	name := ctx.Query("name") // 名称过滤
	path := ctx.Query("path") // 路径过滤

	// 构建查询条件
	query := request.GetPageQuery(ctx)
	query.AddFilter(
		request.NewLikeFilter("name", name), // 名称模糊查询
		request.NewLikeFilter("path", path), // 路径模糊查询
	)

	// 执行分页查询
	list, count, err := sambaShare.List(query)

	// 返回查询结果
	if err == nil {
		response.List(ctx, "", count, list) // 返回列表数据和总数
	} else {
		response.NoContent(ctx, "无数据！") // 返回空数据提示
	}
}

// ListDir 列出指定路径下的目录，用于选择NFS的共享目录
// 参数:
//   - ctx: Gin上下文
func (app *SambaApp) ListDir(ctx *gin.Context) {
	// 获取系统路径分隔符（Windows为\，Linux/Mac为/）
	PthSep := string(os.PathSeparator)

	// 获取要列出的路径
	path := ctx.Query("path")

	// 确保路径以分隔符结尾，方便后续拼接文件名
	if !strings.HasSuffix(path, PthSep) {
		path += PthSep
	}

	// 检查路径是否存在且是目录
	fileInfo, err := os.Stat(path)
	if err != nil || !fileInfo.IsDir() {
		// 路径不存在或不是目录，返回空列表
		response.Data(ctx, "", []data.Map{})
		return
	}

	// 初始化结果列表
	list := []data.Map{}

	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		// 读取目录失败，返回空列表
		response.Data(ctx, "", list)
		return
	}

	// 列出目录中的所有条目
	for _, entry := range entries {
		// 获取文件信息
		info, err := entry.Info()
		if err != nil {
			// 获取信息失败，跳过当前条目
			continue
		}

		// 获取文件模式
		mode := info.Mode()

		// 只添加目录到列表中，忽略普通文件
		if mode.IsDir() {
			// 添加目录到结果列表，key为完整路径，label为目录名
			list = append(list, data.Map{"key": path + info.Name(), "label": info.Name()})
		}
	}

	// 返回目录列表结果
	response.Data(ctx, "", list)
}

// GetDirOwner 获取目录所有者信息
func (app *SambaApp) GetDirOwner(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		response.Error(ctx, fmt.Errorf("路径不能为空"))
		return
	}

	// 检查路径是否存在
	fileInfo, err := os.Stat(path)
	if err != nil || !fileInfo.IsDir() {
		response.Error(ctx, fmt.Errorf("路径不存在或不是目录"))
		return
	}

	sambaService := nas.NewSambaService()
	owner, group, err := sambaService.GetDirOwner(path)
	if err != nil {
		response.Error(ctx, fmt.Errorf("获取目录所有者失败: %s", err.Error()))
		return
	}

	result := map[string]string{
		"owner": owner,
		"group": group,
	}
	response.Data(ctx, "", result)
}

// Save 重写保存方法，保存后自动生成配置
func (app *SambaApp) Save(ctx *gin.Context) {
	// 调用基础保存方法
	app.BaseApp.Save(ctx)

	// 如果保存成功，自动生成配置
	if ctx.Writer.Status() == 200 {
		app.autoGenerateShares(ctx)
	}
}

// Delete 重写删除方法，删除后自动生成配置
func (app *SambaApp) Delete(ctx *gin.Context) {
	// 调用基础删除方法
	app.BaseApp.Delete(ctx)

	// 如果删除成功，自动生成配置
	if ctx.Writer.Status() == 200 {
		app.autoGenerateShares(ctx)
	}
}

// Enable 启用Samba共享
func (app *SambaApp) Enable(ctx *gin.Context) {
	// 调用基础Enable方法
	app.BaseApp.Enable(ctx)

	// 如果启用成功，自动生成配置
	if ctx.Writer.Status() == 200 {
		app.autoGenerateShares(ctx)
	}
}

// Disable 禁用Samba共享
func (app *SambaApp) Disable(ctx *gin.Context) {
	// 调用基础Disable方法
	app.BaseApp.Disable(ctx)

	// 如果禁用成功，自动生成配置
	if ctx.Writer.Status() == 200 {
		app.autoGenerateShares(ctx)
	}
}

// GetServiceStatus 获取Samba服务状态
func (app *SambaApp) GetServiceStatus(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	status := sambaService.GetStatus()
	response.Data(ctx, "", status)
}

// StartService 启动Samba服务
func (app *SambaApp) StartService(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.Start(); err != nil {
		response.Error(ctx, fmt.Errorf("启动Samba服务失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "Samba服务启动成功")
}

// StopService 停止Samba服务
func (app *SambaApp) StopService(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.Stop(); err != nil {
		response.Error(ctx, fmt.Errorf("停止Samba服务失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "Samba服务停止成功")
}

// RestartService 重启Samba服务
func (app *SambaApp) RestartService(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.Restart(); err != nil {
		response.Error(ctx, fmt.Errorf("重启Samba服务失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "Samba服务重启成功")
}

// EnableService 启用Samba服务开机自启
func (app *SambaApp) EnableService(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.Enable(); err != nil {
		response.Error(ctx, fmt.Errorf("启用Samba服务开机自启失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "Samba服务开机自启启用成功")
}

// DisableService 禁用Samba服务开机自启
func (app *SambaApp) DisableService(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.Disable(); err != nil {
		response.Error(ctx, fmt.Errorf("禁用Samba服务开机自启失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "Samba服务开机自启禁用成功")
}

// ReloadConfig 重新加载Samba配置
func (app *SambaApp) ReloadConfig(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.ReloadConfig(); err != nil {
		response.Error(ctx, fmt.Errorf("重新加载Samba配置失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "Samba配置重新加载成功")
}

// GetSyncStatus 获取同步状态
func (app *SambaApp) GetSyncStatus(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	status, err := sambaService.GetSyncStatus()
	if err != nil {
		response.Error(ctx, fmt.Errorf("获取同步状态失败: %s", err.Error()))
		return
	}
	response.Data(ctx, "", status)
}

// SyncFromServer 从服务器同步配置到数据库
func (app *SambaApp) SyncFromServer(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	count, err := sambaService.SyncFromServer()
	if err != nil {
		response.Error(ctx, fmt.Errorf("从服务器同步配置失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "同步成功，新增 "+strconv.Itoa(count)+" 个配置")
}

// SyncToServer 将数据库配置同步到服务器
func (app *SambaApp) SyncToServer(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.SyncToServer(); err != nil {
		response.Error(ctx, fmt.Errorf("同步配置到服务器失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "配置同步到服务器成功")
}

// GetSmbConfig 获取smb.conf配置文件内容
func (app *SambaApp) GetSmbConfig(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	config, err := sambaService.GetSmbConfig()
	if err != nil {
		response.Error(ctx, fmt.Errorf("获取smb.conf配置失败: %s", err.Error()))
		return
	}
	response.Data(ctx, "", config)
}

// BackupSmbConfig 备份smb.conf配置文件
func (app *SambaApp) BackupSmbConfig(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.BackupSmbConfig(); err != nil {
		response.Error(ctx, fmt.Errorf("备份smb.conf配置失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "smb.conf配置备份成功")
}

// TestConfig 测试Samba配置
func (app *SambaApp) TestConfig(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	if err := sambaService.TestConfig(); err != nil {
		response.Error(ctx, fmt.Errorf("配置测试失败: %s", err.Error()))
		return
	}
	response.Success(ctx, "配置测试通过")
}

// ListSambaUsers 获取Samba用户列表
func (app *SambaApp) ListSambaUsers(ctx *gin.Context) {
	sambaService := nas.NewSambaService()
	users, err := sambaService.ListSambaUsers()
	if err != nil {
		response.Error(ctx, fmt.Errorf("获取Samba用户列表失败: %s", err.Error()))
		return
	}
	response.Data(ctx, "", users)
}

// AddSambaUser 添加Samba用户
func (app *SambaApp) AddSambaUser(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, fmt.Errorf("参数错误: %s", err.Error()))
		return
	}

	sambaService := nas.NewSambaService()
	if err := sambaService.AddSambaUser(req.Username, req.Password); err != nil {
		response.Error(ctx, fmt.Errorf("添加Samba用户失败: %s", err.Error()))
		return
	}

	response.Success(ctx, "Samba用户添加成功")
}

// DeleteSambaUser 删除Samba用户
func (app *SambaApp) DeleteSambaUser(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, fmt.Errorf("参数错误: %s", err.Error()))
		return
	}

	sambaService := nas.NewSambaService()
	if err := sambaService.DeleteSambaUser(req.Username); err != nil {
		response.Error(ctx, fmt.Errorf("删除Samba用户失败: %s", err.Error()))
		return
	}

	response.Success(ctx, "Samba用户删除成功")
}

// ChangeSambaUserPassword 修改Samba用户密码
func (app *SambaApp) ChangeSambaUserPassword(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, fmt.Errorf("参数错误: %s", err.Error()))
		return
	}

	sambaService := nas.NewSambaService()
	if err := sambaService.ChangeSambaUserPassword(req.Username, req.Password); err != nil {
		response.Error(ctx, fmt.Errorf("修改Samba用户密码失败: %s", err.Error()))
		return
	}

	response.Success(ctx, "Samba用户密码修改成功")
}

// autoGenerateShares 自动生成Samba共享配置
func (app *SambaApp) autoGenerateShares(ctx *gin.Context) {
	// 查询所有启用的Samba共享 (is_disable = 0 表示启用)
	var shares []nas.SambaShare
	if err := global.DB.Where("is_disable = ?", 0).Find(&shares).Error; err != nil {
		log.Printf("自动生成Samba配置失败，查询共享出错: %v", err)
		return
	}

	// 生成配置
	sambaService := nas.NewSambaService()
	if err := sambaService.GenerateShares(shares); err != nil {
		log.Printf("自动生成Samba配置失败: %v", err)
		return
	}

	// 重新加载配置
	if err := sambaService.ReloadConfig(); err != nil {
		log.Printf("自动重新加载Samba配置失败: %v", err)
		return
	}

	log.Println("Samba配置自动生成成功")
}
