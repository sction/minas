package depops

import (
	"server/core/app/request"
	"server/core/app/response"
	"server/core/app/webapi"
	"server/service/depops"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NodeApp 节点应用结构体
type NodeApp struct {
	webapi.BaseApp[depops.Node]
}

// AddRoutes 添加节点管理相关的路由配置
func (app *NodeApp) AddRoutes(parentGroup *gin.RouterGroup) {
	// 创建节点管理路由组
	group := parentGroup.Group("/node")

	// 设置节点更新操作允许修改的字段
	app.UpdateFields = []string{
		"name", "ip_address", "ssh_port", "username", "password",
		"private_key", "private_key_pass", "tags", "hostname", "remark",
	}

	// 添加基础CRUD路由
	webapi.AddBaseRoutes(group, app)

	// 添加自定义路由
	group.POST("/test-connection/:id", app.TestConnection)
	group.POST("/test-connection-data", app.TestConnectionData)
	group.GET("/tags", app.GetAllTags)
	group.POST("/execute-command/:id", app.ExecuteCommand)
}

// List 获取节点列表
func (app *NodeApp) List(ctx *gin.Context) {
	query := request.GetPageQuery(ctx)
	var entity depops.Node
	list, count, err := entity.List(query)
	if err == nil {
		response.List(ctx, "", count, list)
	} else {
		response.Error(ctx, err)
	}
}

// TestConnection 测试节点连接
func (app *NodeApp) TestConnection(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(ctx, "无效的节点ID")
		return
	}

	var node depops.Node
	node, err = node.Load(uint(id))
	if err != nil {
		response.NotFound(ctx, "节点不存在")
		return
	}

	err = node.TestConnection()
	if err != nil {
		response.Error(ctx, err)
		return
	}

	err = node.GetSystemInfo()
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "连接测试成功")
}

// TestConnectionData 使用节点数据测试连接（无需保存）
func (app *NodeApp) TestConnectionData(ctx *gin.Context) {
	var nodeData depops.Node
	if err := ctx.ShouldBindJSON(&nodeData); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	err := nodeData.TestConnectionOnly()
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "连接测试成功")
}

// GetAllTags 获取所有标签
func (app *NodeApp) GetAllTags(ctx *gin.Context) {
	tags, err := depops.GetAllTags()
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Data(ctx, "", tags)
}

// ExecuteCommand 执行命令
func (app *NodeApp) ExecuteCommand(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(ctx, "无效的节点ID")
		return
	}

	type CommandRequest struct {
		Command string `json:"command" binding:"required"`
	}

	var req CommandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	var node depops.Node
	node, err = node.Load(uint(id))
	if err != nil {
		response.NotFound(ctx, "节点不存在")
		return
	}

	output, err := node.ExecuteCommand(req.Command)
	if err != nil {
		response.Data(ctx, "命令执行完成（有错误）", map[string]interface{}{
			"output": output,
			"error":  err.Error(),
		})
		return
	}

	response.Data(ctx, "命令执行成功", map[string]interface{}{
		"output": output,
	})
}
