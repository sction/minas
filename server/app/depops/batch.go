package depops

import (
	"server/core/app/response"
	"server/service/depops"

	"github.com/gin-gonic/gin"
)

// BatchOperationApp 批量操作应用结构体
type BatchOperationApp struct{}

// NewBatchOperationApp 创建批量操作应用实例
func NewBatchOperationApp() *BatchOperationApp {
	return &BatchOperationApp{}
}

// AddRoutes 添加批量操作相关路由
func (app *BatchOperationApp) AddRoutes(parentGroup *gin.RouterGroup) {
	group := parentGroup.Group("/batch")

	// 批量操作路由
	group.POST("/restart", app.RestartNodes)
	group.POST("/manage-keys", app.ManageKeys)
	group.POST("/set-timezone", app.SetTimezone)
	group.POST("/manage-firewall", app.ManageFirewall)
	group.POST("/execute-command", app.ExecuteCommand)
	group.POST("/upload-file", app.UploadFile)
	group.POST("/install-packages", app.InstallPackages)
	group.POST("/deploy-docker-image", app.DeployDockerImage)
	group.POST("/scan-docker-images", app.ScanDockerImages)
	group.POST("/deploy-docker-images-from-directory", app.DeployDockerImagesFromDirectory)
	group.POST("/execute-script", app.ExecuteScript)
	group.GET("/list-files", app.ListFiles)
}

// BatchRequest 批量操作请求基础结构
type BatchRequest struct {
	NodeIDs []int `json:"node_ids" binding:"required"`
}

// RestartRequest 重启节点请求
type RestartRequest struct {
	BatchRequest
}

// RestartNodes 批量重启节点
func (app *BatchOperationApp) RestartNodes(ctx *gin.Context) {
	var req RestartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.RestartNodesRequest{
		NodeIDs: nodeIDs,
	}

	results := service.RestartNodes(serviceReq)
	response.Data(ctx, "批量重启操作完成", results)
}

// ManageKeysRequest 管理密钥请求
type ManageKeysRequest struct {
	BatchRequest
	PublicKey string `json:"public_key" binding:"required"`
	Action    string `json:"action" binding:"required"`
}

// ManageKeys 批量管理SSH密钥
func (app *BatchOperationApp) ManageKeys(ctx *gin.Context) {
	var req ManageKeysRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.ManageKeysRequest{
		NodeIDs:   nodeIDs,
		PublicKey: req.PublicKey,
		Action:    req.Action,
	}

	results := service.ManageAuthKeys(serviceReq)
	response.Data(ctx, "批量密钥管理操作完成", results)
}

// SetTimezoneRequest 设置时区请求
type SetTimezoneRequest struct {
	BatchRequest
	Timezone string `json:"timezone" binding:"required"`
}

// SetTimezone 批量设置时区
func (app *BatchOperationApp) SetTimezone(ctx *gin.Context) {
	var req SetTimezoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.SetTimezoneRequest{
		NodeIDs:  nodeIDs,
		Timezone: req.Timezone,
	}

	results := service.SetTimezone(serviceReq)
	response.Data(ctx, "批量时区设置操作完成", results)
}

// ManageFirewallRequest 管理防火墙请求
type ManageFirewallRequest struct {
	BatchRequest
	Action string   `json:"action" binding:"required"`
	Ports  []int    `json:"ports,omitempty"`
	IPs    []string `json:"ips,omitempty"`
}

// ManageFirewall 批量管理防火墙
func (app *BatchOperationApp) ManageFirewall(ctx *gin.Context) {
	var req ManageFirewallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.FirewallRuleRequest{
		NodeIDs: nodeIDs,
		Action:  req.Action,
		Ports:   req.Ports,
		IPs:     req.IPs,
	}

	results := service.ManageFirewall(serviceReq)
	response.Data(ctx, "批量防火墙管理操作完成", results)
}

// ExecuteCommandRequest 执行命令请求
type ExecuteCommandRequest struct {
	BatchRequest
	Command string `json:"command" binding:"required"`
}

// ExecuteCommand 批量执行命令
func (app *BatchOperationApp) ExecuteCommand(ctx *gin.Context) {
	var req ExecuteCommandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.ExecuteCommandRequest{
		NodeIDs: nodeIDs,
		Command: req.Command,
	}

	results := service.ExecuteCommand(serviceReq)
	response.Data(ctx, "批量命令执行操作完成", results)
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	BatchRequest
	LocalPath   string `json:"local_path" binding:"required"`
	RemotePath  string `json:"remote_path" binding:"required"`
	Permissions string `json:"permissions,omitempty"`
}

// UploadFile 批量上传文件
func (app *BatchOperationApp) UploadFile(ctx *gin.Context) {
	var req UploadFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.UploadFileRequest{
		NodeIDs:     nodeIDs,
		LocalPath:   req.LocalPath,
		RemotePath:  req.RemotePath,
		Permissions: req.Permissions,
	}

	results := service.UploadFile(serviceReq)
	response.Data(ctx, "批量文件上传操作完成", results)
}

// InstallPackagesRequest 安装软件包请求
type InstallPackagesRequest struct {
	BatchRequest
	Packages []string `json:"packages" binding:"required"`
	Action   string   `json:"action" binding:"required"`
}

// InstallPackages 批量安装软件包
func (app *BatchOperationApp) InstallPackages(ctx *gin.Context) {
	var req InstallPackagesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.InstallPackageRequest{
		NodeIDs:  nodeIDs,
		Packages: req.Packages,
		Action:   req.Action,
	}

	results := service.InstallPackages(serviceReq)
	response.Data(ctx, "批量软件包管理操作完成", results)
}

// DeployDockerImageRequest 部署Docker镜像请求
type DeployDockerImageRequest struct {
	BatchRequest
	ImagePath string `json:"image_path" binding:"required"`
	ImageName string `json:"image_name,omitempty"`
}

// DeployDockerImage 批量部署Docker镜像
func (app *BatchOperationApp) DeployDockerImage(ctx *gin.Context) {
	var req DeployDockerImageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.DockerImageRequest{
		NodeIDs:   nodeIDs,
		ImagePath: req.ImagePath,
		ImageName: req.ImageName,
	}

	results := service.DeployDockerImage(serviceReq)
	response.Data(ctx, "批量Docker镜像部署操作完成", results)
}

// ExecuteScriptRequest 执行脚本请求
type ExecuteScriptRequest struct {
	BatchRequest
	ScriptPath string `json:"script_path" binding:"required"`
	ScriptType string `json:"script_type" binding:"required"`
}

// ExecuteScript 批量执行脚本
func (app *BatchOperationApp) ExecuteScript(ctx *gin.Context) {
	var req ExecuteScriptRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.ExecuteScriptRequest{
		NodeIDs:    nodeIDs,
		ScriptPath: req.ScriptPath,
		ScriptType: req.ScriptType,
	}

	results := service.ExecuteScript(serviceReq)
	response.Data(ctx, "批量脚本执行操作完成", results)
}

// ScanDockerImagesRequest 扫描Docker镜像请求
type ScanDockerImagesRequest struct {
	DirectoryPath string `json:"directory_path" binding:"required"`
}

// ScanDockerImages 扫描目录中的Docker镜像文件
func (app *BatchOperationApp) ScanDockerImages(ctx *gin.Context) {
	var req ScanDockerImagesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	imageFiles := service.ScanDockerImages(req.DirectoryPath)
	response.Data(ctx, "扫描完成", imageFiles)
}

// DeployDockerImagesFromDirectoryRequest 从目录部署Docker镜像请求
type DeployDockerImagesFromDirectoryRequest struct {
	BatchRequest
	DirectoryPath string `json:"directory_path" binding:"required"`
}

// DeployDockerImagesFromDirectory 批量部署目录中的Docker镜像
func (app *BatchOperationApp) DeployDockerImagesFromDirectory(ctx *gin.Context) {
	var req DeployDockerImagesFromDirectoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	// 转换节点ID类型
	nodeIDs := make([]uint, len(req.NodeIDs))
	for i, id := range req.NodeIDs {
		nodeIDs[i] = uint(id)
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	serviceReq := depops.DockerImagesFromDirectoryRequest{
		NodeIDs:       nodeIDs,
		DirectoryPath: req.DirectoryPath,
	}

	results := service.DeployDockerImagesFromDirectory(serviceReq)
	response.Data(ctx, "批量Docker镜像部署操作完成", results)
}

// FileItem 文件项
type FileItem struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size,omitempty"`
	Path  string `json:"path"`
}

// ListFiles 列出指定路径下的文件和目录
func (app *BatchOperationApp) ListFiles(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		path = "/"
	}

	// 创建服务实例并调用服务方法
	service := &depops.BatchOperationService{}
	files := service.ListFiles(path)
	response.Data(ctx, "", files)
}
