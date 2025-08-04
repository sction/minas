package scheduled

// SchTask.go
// 该文件实现了计划任务的数据模型和相关操作方法
// 支持多种类型的计划任务，包括脚本执行、文件备份、文件清理等

import (
	"encoding/json"
	"fmt"
	"server/core/db"
	"server/service/scheduled/job"
	"server/service/scheduled/job/filebackup"
	"server/service/scheduled/job/fileclean"
	shell "server/service/scheduled/job/script"
	"server/service/scheduled/job/task"
	"server/service/scheduled/log"
	"server/utils"
	"server/utils/cron"
	"server/utils/global"
	"server/utils/logger"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
)

// 计划任务类型定义：
// - SHELL: Shell脚本执行任务
// - FILE_BACKUP: 文件备份任务
// - FILE_CLEAN: 文件清理任务
// - JOB_TASK: 作业任务

// SchTask 计划任务数据模型
// 用于存储和管理系统中的定时任务
// 支持多种类型的任务和Cron表达式调度
type SchTask struct {
	db.BaseModel[SchTask]        // 基础模型字段（ID、创建时间等）
	Name                  string `gorm:"comment:'名称' size:128" json:"name"`                 // 任务名称
	Type                  string `gorm:"comment:'类型' size:50 default:'shell'" json:"type"`  // 任务类型（SHELL/FILE_BACKUP/FILE_CLEAN/JOB_TASK）
	Cron                  string `gorm:"comment:'调度表达式' size:50" json:"cron"`               // Cron表达式
	LastStatus            uint   `gorm:"default:0;comment:'最近状态'" json:"last_status"`       // 任务最近执行状态
	LastRunTime           string `gorm:"comment:'上次执行时间'" json:"last_run_time"`             // 上次执行时间
	NextRunTime           string `gorm:"comment:'下次执行时间'" json:"next_run_time"`             // 下次执行时间
	LogKeepNum            uint   `gorm:"default:0;comment:'保留日志数量'" json:"log_keep_num"`    // 保留的日志数量
	Script                string `gorm:"comment:'任务' size:102400 default:''" json:"script"` // 任务配置脚本（JSON格式）
	Remark                string `gorm:"comment:'备注'" json:"remark"`                        // 任务备注
	ProjectDirID          string `gorm:"comment:'项目目录ID';default:0" json:"project_dir_id"`  // 项目目录ID，关联到项目目录
}

// TableName 返回数据库表名
func (SchTask) TableName() string {
	return "sch_task"
}

// Create 创建新的计划任务
// 首先将任务保存到数据库，然后将其添加到定时任务调度器中
func (schTask SchTask) Create(entity *SchTask) error {
	// 将任务保存到数据库
	err := global.DB.Model(entity).Create(entity).Error
	if err != nil {
		return err
	}

	// 将任务转换为定时任务并添加到调度器
	job, err := entity.toJob()
	if err != nil {
		return err
	}
	return cron.AddJobTask(job)
}

// Update 更新计划任务
// 首先更新数据库中的任务信息，然后更新调度器中的任务
func (schTask SchTask) Update(entity *SchTask, columns ...string) (err error) {
	// 更新数据库中的任务信息
	if len(columns) > 0 {
		// 如果指定了列，只更新指定列
		err = global.DB.Model(entity).Select(columns).Updates(entity).Error
	} else {
		// 否则更新所有列
		err = global.DB.Model(entity).Updates(entity).Error
	}
	if err != nil {
		return err
	}

	// 注意：这里有重复的更新操作，可能是一个错误
	// TODO: 考虑移除这部分重复代码
	if len(columns) > 0 {
		err = global.DB.Model(entity).Select(columns).Updates(entity).Error
	} else {
		err = global.DB.Model(entity).Updates(entity).Error
	}
	if err != nil {
		return err
	}

	// 更新调度器中的任务
	job, err := entity.toJob()
	if err != nil {
		return err
	}
	return cron.UpdateJobTask(job)
}

// Enable 启用指定的计划任务
// 将任务的禁用状态设置为启用，并将其添加到调度器中
func (entity SchTask) Enable(id any) error {
	// 在数据库中启用任务
	err := global.DB.Model(&entity).Where(db.ID_FIELD+" =?", id).Update(db.IS_DISABLE_FIELD, 0).Error
	if err != nil {
		return err
	}

	// 加载任务详细信息
	entity, err = entity.Load(id)
	if err != nil {
		return err
	}

	// 将任务添加到调度器
	job, err := entity.toJob()
	if err != nil {
		return err
	}
	cron.AddJobTask(job)
	return nil
}

// Disable 禁用指定的计划任务
// 将任务的状态设置为禁用，并从调度器中移除
func (entity SchTask) Disable(id any) error {
	// 在数据库中禁用任务
	err := global.DB.Model(&entity).Where(db.ID_FIELD+" =?", id).Update(db.IS_DISABLE_FIELD, 1).Error
	if err != nil {
		return err
	}

	// 从调度器中移除任务
	cron.RemoveJobTask(id)
	return nil
}

// Delete 删除指定的计划任务
// 从数据库中删除任务，并从调度器中移除
func (entity SchTask) Delete(id any) error {
	// 从数据库中删除任务
	err := global.DB.Model(entity).Delete(&entity, id).Error
	if err != nil {
		return err
	}

	// 从调度器中移除任务
	cron.RemoveJobTask(id)
	return nil
}

// Start 启动所有定时任务
// 在系统启动时调用，用于初始化并启动所有有效的计划任务
func Start() {
	// 修复异常退出的任务日志，将未完成的任务标记为异常终止
	fix := log.SchLog{}
	fix.Status = -2 // -2表示异常终止
	global.DB.Model(fix).Where("status=0").Updates(fix)

	// 查询需要启动的定时任务
	// 条件：未禁用、未删除、有有效的Cron表达式
	var list []SchTask
	modelDb := global.DB.Model(SchTask{}).Select("*").Where("is_disable=0 and deleted_at IS NULL and cron IS NOT NULL and cron <> ''")
	err := modelDb.Find(&list).Error
	if err != nil {
		logger.LOG.Errorf("获取定时任务列表失败:%s", err.Error())
	}

	// 将数据库中的任务转换为调度器任务
	jobs := make([]cron.SchJob, 0)
	for _, item := range list {
		job, err := item.toJob()
		if err == nil {
			jobs = append(jobs, job)
		}
	}

	// 启动调度器
	err = cron.Start(jobs)
	if err != nil {
		logger.LOG.Errorf("启动定时任务失败:%s", err.Error())
	}
}

// Exec 手动执行任务
// 将任务转换为可执行的任务并在新的goroutine中运行
func (entity SchTask) Exec() error {
	// 将任务转换为可执行的任务
	job, err := entity.toJob()
	if err != nil {
		return err
	}

	// 在新的goroutine中异步执行任务
	go job.Job.Run()
	return nil
}

// toShellJob 将计划任务转换为Shell脚本任务
// 解析Script字段中的JSON数据，提取shell脚本内容
func (entity SchTask) toShellJob() (shell.ShellJob, error) {
	// 解析JSON数据
	var obj = make(map[string]any)
	err := json.Unmarshal([]byte(entity.Script), &obj)
	if err != nil {
		return shell.ShellJob{}, err
	}

	// 创建Shell脚本任务
	mjob := shell.ShellJob{
		SchJob: job.SchJob{
			TaskId:   entity.ID,
			TaskName: entity.Name,
			TaskType: entity.Type,
		},
		Script: fmt.Sprint(obj["shell"]), // 提取shell脚本内容
	}
	return mjob, nil
}

// toFileBackupJob 将计划任务转换为文件备份任务
// 解析Script字段中的JSON数据，提取文件备份相关配置
func (entity SchTask) toFileBackupJob() (filebackup.FileBackupJob, error) {
	// 解析JSON数据
	var obj = make(map[string]any)
	err := json.Unmarshal([]byte(entity.Script), &obj)
	if err != nil {
		return filebackup.FileBackupJob{}, err
	}

	// 提取源和目标路径
	src := convertor.ToString(obj["source"])
	dst := convertor.ToString(obj["target"])
	srcId := convertor.ToString(obj["source_nas_id"])
	dstId := convertor.ToString(obj["target_nas_id"])
	createEmptySrcDirs, _ := convertor.ToBool(convertor.ToString(obj["is_create_dir"]))
	// 如果有外部存储ID，添加到路径前
	if srcId != "" {
		src = srcId + ":" + src
	}
	if dstId != "" {
		dst = dstId + ":" + dst
	}

	// 创建文件备份任务
	mjob := filebackup.FileBackupJob{
		SchJob: job.SchJob{
			TaskId:   entity.ID,
			TaskName: entity.Name,
			TaskType: entity.Type,
		},
		Type:               utils.GetUint(obj, "type", 0),                         // 备份类型（1备份 2镜像 3双向同步 4完整备份）
		Src:                src,                                                   // 源路径
		Dst:                dst,                                                   // 目标路径
		CreateEmptySrcDirs: createEmptySrcDirs,                                    // 是否创建空源目录
		ShowDebug:          utils.GetBool(obj, "show_debug", true),                // 是否显示调试信息
		Includes:           strings.Split(utils.GetString(obj, "includes"), "\n"), // 包含的文件模式
		Excludes:           strings.Split(utils.GetString(obj, "excludes"), "\n"), // 排除的文件模式
	}
	return mjob, nil
}

// toFileCleanJob 将计划任务转换为文件清理任务
// 解析Script字段中的JSON数据，提取文件清理相关配置
func (entity SchTask) toFileCleanJob() (fileclean.FileCleanJob, error) {
	// 解析JSON数据
	var obj = make(map[string]any)
	err := json.Unmarshal([]byte(entity.Script), &obj)
	if err != nil {
		return fileclean.FileCleanJob{}, err
	}

	// 提取工作目录和存储信息
	workDir := convertor.ToString(obj["work_dir"])
	storage := convertor.ToString(obj["storage"])

	// 如果有外部存储ID，添加到路径前
	if storage != "" {
		workDir = storage + ":" + workDir
	}

	// 创建文件清理任务
	mjob := fileclean.FileCleanJob{
		SchJob: job.SchJob{
			TaskId:   entity.ID,
			TaskName: entity.Name,
			TaskType: entity.Type,
		},
		WorkDir:   workDir,                                               // 工作目录
		Includes:  strings.Split(utils.GetString(obj, "includes"), "\n"), // 包含的文件模式
		Excludes:  strings.Split(utils.GetString(obj, "excludes"), "\n"), // 排除的文件模式
		KeepNum:   utils.GetInt(obj, "keep_num", 7),                      // 保留的文件数量
		OffsetDay: utils.GetInt(obj, "offset_day", 7),                    // 偏移天数
		IsTest:    utils.GetBool(obj, "is_test", true),                   // 是否仅测试不实际执行
		ShowDebug: utils.GetBool(obj, "show_debug", true),                // 是否显示调试信息
	}
	return mjob, nil
}

// toTaskJob 将计划任务转换为作业任务
// 解析Script字段中的JSON数据，提取作业脚本内容
func (entity SchTask) toTaskJob() (task.TaskJob, error) {
	// 解析JSON数据
	var obj = make(map[any]any)
	err := json.Unmarshal([]byte(entity.Script), &obj)
	if err != nil {
		return task.TaskJob{}, err
	}

	// 创建作业任务
	mjob := task.TaskJob{
		SchJob: job.SchJob{
			TaskId:   entity.ID,
			TaskName: entity.Name,
			TaskType: entity.Type,
		},
		Script: fmt.Sprint(obj["script"]), // 作业脚本内容
	}
	return mjob, nil
}

// toJob 将计划任务转换为调度器任务
// 根据任务类型调用相应的转换函数，生成可执行的任务
func (entity SchTask) toJob() (cron.SchJob, error) {
	// 根据任务类型选择相应的转换函数
	switch entity.Type {
	case "SHELL": // Shell脚本任务
		mjob, err := entity.toShellJob()
		if err != nil {
			return cron.SchJob{}, err
		}
		return cron.SchJob{Id: fmt.Sprint(entity.ID), Cron: entity.Cron, Job: mjob}, nil

	case "FILE_BACKUP": // 文件备份任务
		mjob, err := entity.toFileBackupJob()
		if err != nil {
			return cron.SchJob{}, err
		}
		return cron.SchJob{Id: fmt.Sprint(entity.ID), Cron: entity.Cron, Job: mjob}, nil

	case "FILE_CLEAN": // 文件清理任务
		mjob, err := entity.toFileCleanJob()
		if err != nil {
			return cron.SchJob{}, err
		}
		return cron.SchJob{Id: fmt.Sprint(entity.ID), Cron: entity.Cron, Job: mjob}, nil

	case "JOB_TASK": // 作业任务
		mjob, err := entity.toTaskJob()
		if err != nil {
			return cron.SchJob{}, err
		}
		return cron.SchJob{Id: fmt.Sprint(entity.ID), Cron: entity.Cron, Job: mjob}, nil

	default: // 未知的任务类型
		return cron.SchJob{}, fmt.Errorf("未知的任务类型:%s", entity.Type)
	}
}

// CountByType 根据任务类型统计数量
func (entity SchTask) CountByType(taskType string) (int64, error) {
	var count int64
	err := global.DB.Model(&entity).Where("type = ? AND deleted_at IS NULL", taskType).Count(&count).Error
	return count, err
}

// GetTypeStatistics 获取所有类型的统计信息
func (entity SchTask) GetTypeStatistics() (map[string]int64, error) {
	results := make(map[string]int64)

	// 查询各种类型的任务数量
	rows, err := global.DB.Model(&entity).
		Select("type, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("type").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskType string
		var count int64
		if err := rows.Scan(&taskType, &count); err != nil {
			continue
		}
		results[taskType] = count
	}

	return results, nil
}
