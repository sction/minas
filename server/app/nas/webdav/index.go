// Package webdav 提供WebDAV管理相关的API接口
package webdav

import (
	"log"
	"os"
	"server/core/app/request"
	"server/core/app/response"
	"server/core/app/webapi"
	"server/service/nas"
	"server/utils/data"
	"strings"

	"github.com/gin-gonic/gin"
)

// WebDavApp 定义WebDAV管理应用
// 继承自BaseApp，提供基本的CRUD操作
type WebDavApp struct {
	webapi.BaseApp[nas.Webdav] // 使用nas.Webdav模型作为数据类型
}

// AddRoutes 注册WebDAV管理相关的路由
// 参数:
//   - parentGroup: 父路由组
func AddRoutes(parentGroup *gin.RouterGroup) {
	// 创建WebDAV路由组
	group := parentGroup.Group("/webdav")

	// 创建WebDAV应用实例
	app := WebDavApp{}

	// 注册基本的CRUD路由
	webapi.AddBaseRoutes(group, &app)

	// 设置允许更新的字段
	app.UpdateFields = []string{"name", "home", "account", "perms", "remark"}

	// 注册自定义路由
	group.GET("/list-dir", app.ListDir) // 列出目录内容
}

// List 实现WebDAV账号列表查询
// 参数:
//   - ctx: Gin上下文
func (app *WebDavApp) List(ctx *gin.Context) {
	log.Println("List webdav")

	// 创建WebDAV数据库对象
	webdav := nas.Webdav{}

	// 获取查询参数
	name := ctx.Query("name")       // 名称过滤
	account := ctx.Query("account") // 账号过滤

	// 构建查询条件
	query := request.GetPageQuery(ctx)
	query.AddFilter(
		request.NewLikeFilter("name", name),       // 名称模糊查询
		request.NewLikeFilter("account", account), // 账号模糊查询
	)

	// 执行分页查询
	list, count, err := webdav.List(query)

	// 返回查询结果
	if err == nil {
		response.List(ctx, "", count, list) // 返回列表数据和总数
	} else {
		response.NoContent(ctx, "无数据！") // 返回空数据提示
	}
}

// ListDir 列出指定路径下的目录，用于选择WebDAV的主目录
// 参数:
//   - ctx: Gin上下文
func (app *WebDavApp) ListDir(ctx *gin.Context) {
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
