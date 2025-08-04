package system

import (
	"runtime"
	"server/core/app/response"
	"server/service/basic"
	"server/service/nas"
	"server/service/scheduled"
	"server/service/scheduled/log"
	"server/utils"
	"server/utils/config"
	"server/utils/data"
	"server/utils/global"
	"server/utils/simple"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// DashboardStats 仪表板统计数据结构
type DashboardStats struct {
	ProjectCount         int64 `json:"projectCount"`         // 项目数
	WebdavCount          int64 `json:"webdavCount"`          // WebDAV服务数
	ExternalStorageCount int64 `json:"externalStorageCount"` // 外部存储数
	ScheduledTaskCount   int64 `json:"scheduledTaskCount"`   // 计划任务数
}

// TaskTypeStats 任务类型统计
type TaskTypeStats struct {
	Type  string `json:"type"`  // 任务类型
	Count int64  `json:"count"` // 数量
	Label string `json:"label"` // 显示标签
}

// TaskExecutionStats 任务执行统计
type TaskExecutionStats struct {
	Success int64 `json:"success"` // 成功数
	Failed  int64 `json:"failed"`  // 失败数
	Running int64 `json:"running"` // 运行中数
	Total   int64 `json:"total"`   // 总数
}

// TaskTrendData 任务执行趋势数据
type TaskTrendData struct {
	Date    string `json:"date"`    // 日期
	Success int64  `json:"success"` // 成功数
	Failed  int64  `json:"failed"`  // 失败数
	Total   int64  `json:"total"`   // 总数
}

// TaskProjectStats 项目任务统计
type TaskProjectStats struct {
	ProjectID   string `json:"projectId"`   // 项目ID
	ProjectName string `json:"projectName"` // 项目名称
	TaskCount   int64  `json:"taskCount"`   // 任务数量
}

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

// SystemVersion 获取系统版本信息
// 返回当前系统版本、Go版本等信息
// 参数 ctx: Gin上下文
func SystemVersion(ctx *gin.Context) {
	response.Data(ctx, "", data.Map{
		"version":   config.Version,
		"goVersion": runtime.Version(),
		"test":      strings.ReplaceAll(utils.GetWorkDirectory(), "/", "_"),
	})
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

// GetDashboardStats 获取仪表板统计数据
func GetDashboardStats(ctx *gin.Context) {
	stats := DashboardStats{}

	// 获取项目目录数量
	projectDir := basic.ProjectDir{}
	projectCount, err := projectDir.Count()
	if err == nil {
		stats.ProjectCount = projectCount
	}

	// 获取WebDAV服务数量
	webdav := nas.Webdav{}
	webdavCount, err := webdav.Count()
	if err == nil {
		stats.WebdavCount = webdavCount
	}

	// 获取外部存储数量
	externalNas := nas.ExternalNas{}
	externalCount, err := externalNas.Count()
	if err == nil {
		stats.ExternalStorageCount = externalCount
	}

	// 获取计划任务数量
	schTask := scheduled.SchTask{}
	taskCount, err := schTask.Count()
	if err == nil {
		stats.ScheduledTaskCount = taskCount
	}

	response.Data(ctx, "获取统计数据成功", stats)
}

// GetTaskTypeStats 获取任务类型统计
func GetTaskTypeStats(ctx *gin.Context) {
	schTask := scheduled.SchTask{}

	// 获取各种类型的任务数量
	typeStats := []TaskTypeStats{}

	// 脚本任务
	shellCount, _ := schTask.CountByType("SHELL")
	if shellCount > 0 {
		typeStats = append(typeStats, TaskTypeStats{
			Type:  "SHELL",
			Count: shellCount,
			Label: "脚本任务",
		})
	}

	// 文件备份任务
	backupCount, _ := schTask.CountByType("FILE_BACKUP")
	if backupCount > 0 {
		typeStats = append(typeStats, TaskTypeStats{
			Type:  "FILE_BACKUP",
			Count: backupCount,
			Label: "文件备份",
		})
	}

	// 文件清理任务
	cleanCount, _ := schTask.CountByType("FILE_CLEAN")
	if cleanCount > 0 {
		typeStats = append(typeStats, TaskTypeStats{
			Type:  "FILE_CLEAN",
			Count: cleanCount,
			Label: "文件清理",
		})
	}

	// 作业任务
	jobCount, _ := schTask.CountByType("JOB_TASK")
	if jobCount > 0 {
		typeStats = append(typeStats, TaskTypeStats{
			Type:  "JOB_TASK",
			Count: jobCount,
			Label: "作业任务",
		})
	}

	response.Data(ctx, "获取任务类型统计成功", typeStats)
}

// GetTaskExecutionStats 获取任务执行统计
func GetTaskExecutionStats(ctx *gin.Context) {
	schLog := log.SchLog{}

	stats := TaskExecutionStats{}

	// 获取成功任务数
	successCount, _ := schLog.CountByStatus(1)
	stats.Success = successCount

	// 获取失败任务数
	failedCount, _ := schLog.CountByStatus(-1)
	stats.Failed = failedCount

	// 获取运行中任务数
	runningCount, _ := schLog.CountByStatus(0)
	stats.Running = runningCount

	// 计算总数
	stats.Total = stats.Success + stats.Failed + stats.Running

	response.Data(ctx, "获取任务执行统计成功", stats)
}

// GetTaskTrend 获取任务执行趋势
func GetTaskTrend(ctx *gin.Context) {
	days := 30
	if d := ctx.Query("days"); d != "" {
		// 可以从查询参数获取天数，这里保持默认30天
	}

	schLog := log.SchLog{}
	trendData := []TaskTrendData{}

	// 生成近30天的数据
	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		// 获取当天的成功和失败任务数
		successCount, _ := schLog.CountByDateAndStatus(dateStr, 1)
		failedCount, _ := schLog.CountByDateAndStatus(dateStr, -1)

		trendData = append(trendData, TaskTrendData{
			Date:    dateStr,
			Success: successCount,
			Failed:  failedCount,
			Total:   successCount + failedCount,
		})
	}

	response.Data(ctx, "获取任务执行趋势成功", trendData)
}

// GetTaskProjectStats 获取各项目下任务数量统计
func GetTaskProjectStats(ctx *gin.Context) {
	// 查询项目目录和对应的任务数量
	var results []TaskProjectStats

	// 使用原生SQL查询，关联项目目录表和任务表
	rows, err := global.DB.Raw(`
		SELECT 
			COALESCE(CAST(pd.id AS TEXT), '0') as project_id,
			COALESCE(pd.name, '未分类') as project_name,
			COUNT(st.id) as task_count
		FROM project_dirs pd
		LEFT JOIN sch_task st ON CAST(pd.id AS TEXT) = st.project_dir_id
		WHERE pd.deleted_at IS NULL AND pd.is_disable = 0
		GROUP BY pd.id, pd.name
		UNION ALL
		SELECT 
			'0' as project_id,
			'未分类' as project_name,
			COUNT(st.id) as task_count
		FROM sch_task st
		WHERE (st.project_dir_id = '' OR st.project_dir_id = '0' OR st.project_dir_id IS NULL)
		  AND st.deleted_at IS NULL
		HAVING COUNT(st.id) > 0
		ORDER BY task_count DESC
		LIMIT 10
	`).Rows()

	if err != nil {
		response.Error(ctx, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stat TaskProjectStats
		if err := rows.Scan(&stat.ProjectID, &stat.ProjectName, &stat.TaskCount); err != nil {
			continue
		}
		if stat.TaskCount > 0 {
			results = append(results, stat)
		}
	}

	response.Data(ctx, "获取项目任务统计成功", results)
}
