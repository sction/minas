// package log 定义了计划任务日志相关的结构和方法
package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"server/core/app/request"
	"server/core/db"
	"server/data"
	"server/service/scheduled/job"
	"server/utils"
	"server/utils/global"
	"server/utils/logger"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/fileutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SchLog 计划任务日志结构体
// status 0 执行中 1完成 -1失败
type SchLog struct {
	ID        uint         `gorm:"primary_key" json:"id" mapstructure:"id"`   // 主键ID
	TaskId    uint         `gorm:"comment:'任务ID'" json:"task_id"`             // 关联的任务ID
	Status    int          `gorm:"default:0;comment:'状态'" json:"status"`      // 任务状态：0-执行中，1-完成，-1-失败
	LogPath   string       `gorm:"comment:'日志文件' default:''" json:"log_path"` // 日志文件路径
	LogText   string       `gorm:"comment:'日志内容' default:''" json:"log_text"` // 日志文本内容
	StartTime db.LocalTime `gorm:"comment:'开始时间'" json:"start_time"`          // 任务开始时间
	EndTime   db.LocalTime `gorm:"comment:'结束时间'" json:"end_time"`            // 任务结束时间
}

// TableName 指定数据库表名
func (SchLog) TableName() string {
	return "sch_task_log" // 返回数据库中对应的表名
}

// Start 开始记录任务执行日志
// 创建一条新的日志记录，标记任务开始执行
func (entity *SchLog) Start(job job.SchJob) error {
	logger.LOG.Infof("RUN: taskId:%d, taskName:%s, taskType:%s", job.TaskId, job.TaskName, job.TaskType)
	entity.TaskId = job.TaskId                          // 设置任务ID
	entity.StartTime = db.LocalTime{}.Now()             // 设置开始时间为当前时间
	entity.Status = 0                                   // 设置状态为执行中(0)
	return global.DB.Model(entity).Create(entity).Error // 创建日志记录并返回可能的错误
}

// StartLogFile 开始记录任务执行日志并创建日志文件
// 创建一条新的日志记录，并创建对应的日志文件
// 返回工作目录、日志文件路径和可能的错误
func (entity *SchLog) StartLogFile(job job.SchJob) (string, string, error) {
	// 构建工作目录和日志文件路径
	workDir := filepath.Join(data.WorkDir, "job")
	logPath := filepath.Join(workDir, "logs", fmt.Sprintf("%d", job.TaskId), fmt.Sprintf("%s.log", time.Now().Format("20060102_150405")))
	// 确保日志目录存在
	os.MkdirAll(filepath.Dir(logPath), 0777)
	logger.LOG.Infof("RUN: taskId:%d, taskName:%s, taskType:%s", job.TaskId, job.TaskName, job.TaskType)
	// 设置日志记录的基本信息
	entity.TaskId = job.TaskId
	entity.StartTime = db.LocalTime{}.Now()
	entity.Status = 0
	entity.LogPath = logPath
	// 创建日志记录并返回工作目录、日志路径和可能的错误
	return workDir, logPath, global.DB.Model(entity).Create(entity).Error
}

// Log 写入日志内容
// 更新日志记录的文本内容
func (entity *SchLog) Log(logText []string) error {
	// 将日志文本数组合并为字符串，以换行符分隔
	entity.LogText = strings.Join(logText, "\n")
	// 更新数据库中的日志文本字段
	return global.DB.Model(entity).Select("log_text").Updates(entity).Error
}

// Success 标记任务执行成功
// 更新日志记录，设置状态为成功(1)，记录结束时间和日志内容
func (entity *SchLog) Success(job job.SchJob, logText []string) error {
	// 合并日志文本
	entity.LogText = strings.Join(logText, "\n")
	// 设置结束时间为当前时间
	entity.EndTime = db.LocalTime{}.Now()
	// 设置状态为成功(1)
	entity.Status = 1
	logger.LOG.Infof("SUCCESS: taskId:%d, taskName:%s, taskType:%s", job.TaskId, job.TaskName, job.TaskType)
	// 更新数据库中的日志记录
	return global.DB.Model(entity).Select("log_text", "end_time", "status").Updates(entity).Error
}

// Error 标记任务执行失败
// 更新日志记录，设置状态为失败(-1)，记录结束时间和日志内容
func (entity *SchLog) Error(job job.SchJob, logText []string) error {
	// 合并日志文本
	entity.LogText = strings.Join(logText, "\n")
	// 设置结束时间为当前时间
	entity.EndTime = db.LocalTime{}.Now()
	// 设置状态为失败(-1)
	entity.Status = -1
	logger.LOG.Errorf("ERROR: taskId:%d, taskName:%s, taskType:%s", job.TaskId, job.TaskName, job.TaskType)
	// 更新数据库中的日志记录
	return global.DB.Model(entity).Select("log_text", "end_time", "status").Updates(entity).Error
}

// Load 按主键查询日志记录
// 根据ID加载日志记录，并尝试读取关联的日志文件内容
func (entity SchLog) Load(id any) (SchLog, error) {
	// 根据ID从数据库加载日志记录
	err := global.DB.Model(entity).Take(&entity, id).Error
	// 如果有日志文件路径且文件存在，则读取文件内容
	if entity.LogPath != "" && fileutil.IsExist(entity.LogPath) {
		logs, err := fileutil.ReadFileToString(entity.LogPath)
		if err != nil {
			logs = "读取日志文件时发生错误：" + err.Error()
		}
		// 将文件内容添加到日志文本前面
		entity.LogText = logs + "\n" + entity.LogText
	}
	// 处理错误，但忽略记录不存在的错误
	if err != nil && err != gorm.ErrRecordNotFound {
		return entity, err
	}
	return entity, nil
}

// CountQuery 计算符合查询条件的记录总数
// 根据查询条件统计日志记录的数量
func (entity SchLog) CountQuery(query request.PageQuery) (int64, error) {
	// 创建数据库查询对象
	countDb := global.DB.Model(entity)
	countDb = countDb.Select("count(ID)")
	// 确保过滤条件不为空
	if utils.IsEmpty(query.Filters) {
		query.Filters = []request.QueryFilter{}
	}
	var count int64
	// 应用查询条件
	countDb = db.QueryWhere(countDb, query.Filters)
	// 执行计数查询
	err := countDb.Count(&count).Error
	return count, err
}

// ListQuery 根据查询条件获取日志记录列表
// 应用分页、排序和过滤条件查询日志记录
func (entity SchLog) ListQuery(query request.PageQuery) (list []SchLog, err error) {
	log.Printf("query:%v\n", query)
	// 创建数据库查询对象
	modelDb := global.DB.Model(entity)

	// 设置查询字段
	if utils.IsEmpty(query.Columns) {
		modelDb = modelDb.Select("*") // 如果未指定字段，则查询所有字段
	} else {
		modelDb = modelDb.Select(query.Columns) // 查询指定字段
	}

	// 确保过滤条件不为空
	if utils.IsEmpty(query.Filters) {
		query.Filters = []request.QueryFilter{}
	}
	// 应用查询条件
	modelDb = db.QueryWhere(modelDb, query.Filters)

	// 应用排序条件
	if !utils.IsEmpty(query.Sorts) {
		for _, sort := range query.Sorts {
			if sort[0:1] == "-" {
				// 降序排序（字段名前有'-'符号）
				modelDb = modelDb.Order(clause.OrderByColumn{Column: clause.Column{Name: sort[1:]}, Desc: true})
			} else {
				// 升序排序
				modelDb = modelDb.Order(sort)
			}
		}
	}

	// 应用分页条件
	if query.Page < 1 {
		query.Page = 1 // 确保页码至少为1
	}
	if query.Size > 0 {
		modelDb = modelDb.Limit(query.Size).Offset((query.Page - 1) * query.Size)
	}

	// 执行查询
	err = modelDb.Find(&list).Error
	return list, err
}

// List 分页查询日志记录
// 执行分页查询并返回记录列表和总记录数
func (entity SchLog) List(query request.PageQuery) (list []SchLog, count int64, err error) {
	// 先获取符合条件的记录列表
	list, err = entity.ListQuery(query)
	if err == nil {
		// 如果查询成功，再获取符合条件的记录总数
		count, err = entity.CountQuery(query)
		return list, count, err
	} else {
		// 如果查询失败，返回错误
		return list, 0, err
	}
}

// CountByStatus 根据任务状态统计数量
func (entity SchLog) CountByStatus(status int) (int64, error) {
	var count int64
	err := global.DB.Model(&entity).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountByDateAndStatus 根据日期和状态统计数量
func (entity SchLog) CountByDateAndStatus(date string, status int) (int64, error) {
	var count int64
	err := global.DB.Model(&entity).
		Where("status = ? AND DATE(start_time) = ?", status, date).
		Count(&count).Error
	return count, err
}

// GetStatusStatistics 获取所有状态的统计信息
func (entity SchLog) GetStatusStatistics() (map[int]int64, error) {
	results := make(map[int]int64)

	// 查询各种状态的任务数量
	rows, err := global.DB.Model(&entity).
		Select("status, COUNT(*) as count").
		Group("status").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status int
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		results[status] = count
	}

	return results, nil
}
