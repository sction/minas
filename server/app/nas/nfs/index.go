//go:build linux
// +build linux

// Package nfs 提供NFS服务管理相关的API接口
package nfs

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
	"server/utils/system"
	"strings"

	"github.com/gin-gonic/gin"
)

// NfsApp 定义NFS管理应用
// 继承自BaseApp，提供基本的CRUD操作
type NfsApp struct {
	webapi.BaseApp[nas.NfsShare] // 使用nas.NfsShare模型作为数据类型
}

// AddRoutes 注册NFS管理相关的路由
// 参数:
//   - parentGroup: 父路由组
func AddRoutes(parentGroup *gin.RouterGroup) {
	// 创建NFS路由组
	group := parentGroup.Group("/nfs")

	// 创建NFS应用实例
	app := NfsApp{}

	// 设置允许更新的字段
	app.UpdateFields = []string{"name", "path", "client_ip", "permission", "sync", "root_squash", "subtree_check", "options", "remark", "is_disable"}

	// 注册自定义路由
	group.GET("/service/status", app.GetServiceStatus)   // 获取NFS服务状态
	group.POST("/service/start", app.StartService)       // 启动NFS服务
	group.POST("/service/stop", app.StopService)         // 停止NFS服务
	group.POST("/service/restart", app.RestartService)   // 重启NFS服务
	group.POST("/service/enable", app.EnableAutoStart)   // 启用开机自启
	group.POST("/service/disable", app.DisableAutoStart) // 禁用开机自启
	group.POST("/reload-exports", app.ReloadExports)     // 导出配置到系统
	group.GET("/list-dir", app.ListDir)                  // 列出目录内容
	group.GET("/sync-status", app.GetSyncStatus)         // 获取同步状态
	group.POST("/sync-from-server", app.SyncFromServer)  // 从服务器同步到数据库
	group.POST("/sync-to-server", app.SyncToServer)      // 从数据库同步到服务器
	group.GET("/exports-config", app.GetExportsConfig)   // 获取exports配置文件内容
	group.POST("/backup-exports", app.BackupExports)     // 备份exports配置文件

	// 注册基本的CRUD路由
	webapi.AddBaseRoutes(group, &app)
}

// List 实现NFS共享列表查询
// 参数:
//   - ctx: Gin上下文
func (app *NfsApp) List(ctx *gin.Context) {
	log.Println("List nfs shares")

	// 创建NFS共享数据库对象
	nfsShare := nas.NfsShare{}

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
	list, count, err := nfsShare.List(query)

	// 返回查询结果
	if err == nil {
		response.List(ctx, "", count, list) // 返回列表数据和总数
	} else {
		response.NoContent(ctx, "无数据！") // 返回空数据提示
	}
}

// GetServiceStatus 获取NFS服务状态
func (app *NfsApp) GetServiceStatus(ctx *gin.Context) {
	nfsService := &nas.NfsService{}

	// 检查NFS服务是否可用
	if !system.IsNfsServiceAvailable() {
		// 在容器内运行，返回不可用状态
		status := map[string]interface{}{
			"installed": false,
			"running":   false,
			"enabled":   false,
			"available": false,
			"message":   "NFS服务在容器环境中不可用",
		}
		response.Data(ctx, "", status)
		return
	}

	status := nfsService.GetStatus()
	status["available"] = true
	response.Data(ctx, "", status)
}

// StartService 启动NFS服务
func (app *NfsApp) StartService(ctx *gin.Context) {
	nfsService := &nas.NfsService{}

	// 检查NFS服务是否可用
	if !system.IsNfsServiceAvailable() {
		response.Error(ctx, fmt.Errorf("NFS服务在容器环境中不可用，请在宿主机环境中运行"))
		return
	}

	if err := nfsService.Start(); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "NFS服务启动成功")
}

// StopService 停止NFS服务
func (app *NfsApp) StopService(ctx *gin.Context) {
	nfsService := &nas.NfsService{}

	// 检查NFS服务是否可用
	if !system.IsNfsServiceAvailable() {
		response.Error(ctx, fmt.Errorf("NFS服务在容器环境中不可用，请在宿主机环境中运行"))
		return
	}

	if err := nfsService.Stop(); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "NFS服务停止成功")
}

// RestartService 重启NFS服务
func (app *NfsApp) RestartService(ctx *gin.Context) {
	nfsService := &nas.NfsService{}

	// 检查NFS服务是否可用
	if !system.IsNfsServiceAvailable() {
		response.Error(ctx, fmt.Errorf("NFS服务在容器环境中不可用，请在宿主机环境中运行"))
		return
	}

	if err := nfsService.Restart(); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "NFS服务重启成功")
}

// EnableAutoStart 启用NFS服务开机自启
func (app *NfsApp) EnableAutoStart(ctx *gin.Context) {
	nfsService := &nas.NfsService{}

	// 检查NFS服务是否可用
	if !system.IsNfsServiceAvailable() {
		response.Error(ctx, fmt.Errorf("NFS服务在容器环境中不可用，请在宿主机环境中运行"))
		return
	}

	if err := nfsService.Enable(); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "")
}

// DisableAutoStart 禁用NFS服务开机自启
func (app *NfsApp) DisableAutoStart(ctx *gin.Context) {
	nfsService := &nas.NfsService{}

	// 检查NFS服务是否可用
	if !system.IsNfsServiceAvailable() {
		response.Error(ctx, fmt.Errorf("NFS服务在容器环境中不可用，请在宿主机环境中运行"))
		return
	}

	if err := nfsService.Disable(); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "")
}

// EnableShare 启用NFS共享
func (app NfsApp) Enable(ctx *gin.Context) {
	// 获取路径参数中的ID
	id := ctx.Param("id")

	// 查询共享信息
	var share nas.NfsShare
	if err := global.DB.Where("id = ?", id).First(&share).Error; err != nil {
		response.Error(ctx, err)
		return
	}

	// 更新共享状态为启用
	share.IsDisable = 0
	if err := global.DB.Save(&share).Error; err != nil {
		response.Error(ctx, err)
		return
	}

	// 自动导出配置到exports文件
	app.autoExportShares(ctx)

	response.Success(ctx, "NFS共享启用成功")
}

// DisableShare 禁用NFS共享
func (app NfsApp) Disable(ctx *gin.Context) {
	// 获取路径参数中的ID
	id := ctx.Param("id")

	// 查询共享信息
	var share nas.NfsShare
	if err := global.DB.Where("id = ?", id).First(&share).Error; err != nil {
		response.Error(ctx, err)
		return
	}

	// 更新共享状态为禁用
	share.IsDisable = 1
	if err := global.DB.Save(&share).Error; err != nil {
		response.Error(ctx, err)
		return
	}

	// 自动导出配置到exports文件（只导出启用的共享）
	app.autoExportShares(ctx)

	response.Success(ctx, "NFS共享禁用成功")
}

// Save 重写保存方法，保存后自动导出配置
func (app *NfsApp) Save(ctx *gin.Context) {
	// 调用基础保存方法
	app.BaseApp.Save(ctx)

	// 如果保存成功，自动导出配置
	if ctx.Writer.Status() == 200 {
		app.autoExportShares(ctx)
	}
}

// Delete 重写删除方法，删除后自动导出配置
func (app *NfsApp) Delete(ctx *gin.Context) {
	// 调用基础删除方法
	app.BaseApp.Delete(ctx)

	// 如果删除成功，自动导出配置
	if ctx.Writer.Status() == 200 {
		app.autoExportShares(ctx)
	}
}

// reloadExports 自动重新加载NFS配置
func (app *NfsApp) ReloadExports(ctx *gin.Context) {
	// 导出配置
	nfsService := &nas.NfsService{}
	// 重新加载配置
	if err := nfsService.ReloadExports(); err != nil {
		log.Printf("重新加载NFS配置失败: %v", err)
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "重新加载NFS配置成功")
}

// autoExportShares 自动导出NFS共享配置
func (app *NfsApp) autoExportShares(ctx *gin.Context) {
	// 查询所有启用的NFS共享 (is_disable = 0 表示启用)
	var shares []nas.NfsShare
	if err := global.DB.Where("is_disable = ?", 0).Find(&shares).Error; err != nil {
		log.Printf("自动导出NFS配置失败，查询共享出错: %v", err)
		return
	}

	// 导出配置
	nfsService := &nas.NfsService{}
	if err := nfsService.ExportShares(shares); err != nil {
		log.Printf("自动导出NFS配置失败: %v", err)
		return
	}

	// 重新加载配置
	if err := nfsService.ReloadExports(); err != nil {
		log.Printf("自动重新加载NFS配置失败: %v", err)
		return
	}

	log.Println("NFS配置自动导出成功")
}

// ListDir 列出指定路径下的目录，用于选择NFS的共享目录
// 参数:
//   - ctx: Gin上下文
func (app *NfsApp) ListDir(ctx *gin.Context) {
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

// GetSyncStatus 获取同步状态
func (app *NfsApp) GetSyncStatus(ctx *gin.Context) {
	nfsService := &nas.NfsService{}
	status, err := nfsService.GetSyncStatus()
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Data(ctx, "", status)
}

// SyncFromServer 从服务器同步到数据库
func (app *NfsApp) SyncFromServer(ctx *gin.Context) {
	nfsService := &nas.NfsService{}
	count, err := nfsService.SyncFromServer()
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, fmt.Sprintf("从服务器同步了 %d 条配置到数据库", count))
}

// SyncToServer 从数据库同步到服务器
func (app *NfsApp) SyncToServer(ctx *gin.Context) {
	nfsService := &nas.NfsService{}
	if err := nfsService.SyncToServer(); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "同步数据到服务器成功")
}

// GetExportsConfig 获取exports配置文件内容
func (app *NfsApp) GetExportsConfig(ctx *gin.Context) {
	nfsService := &nas.NfsService{}
	content, err := nfsService.GetExportsConfig()
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Data(ctx, "", content)
}

// BackupExports 备份exports配置文件
func (app *NfsApp) BackupExports(ctx *gin.Context) {
	nfsService := &nas.NfsService{}
	if err := nfsService.BackupExports(); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "备份exports配置文件成功")
}
