package database

import (
	"bytes"
	"server/data"
	"server/utils/config"
	"server/utils/global"
	"server/utils/logger"
	"time"

	gormLogger "gorm.io/gorm/logger"

	"server/service/basic"
	"server/service/nas"
	"server/service/scheduled"
	"server/service/scheduled/log"
	"server/service/sflow"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Init 初始化数据库连接和相关配置
// 该函数负责建立数据库连接、设置连接池参数、迁移数据表结构
// 根据配置文件选择不同的数据库驱动(MySQL、PostgreSQL或SQLite)
func Init() {
	// 创建自定义GORM日志记录器
	newLogger := gormLogger.New(
		logger.LOG, // io writer（日志输出的地方，这里是标准输出）
		gormLogger.Config{
			LogLevel: gormLogger.Info, // 日志级别
			Colorful: false,           // 彩色打印
		},
	)
	var db *gorm.DB

	// 根据配置选择不同的数据库类型进行连接
	if config.CONF.Db.Type == "mysql" {
		// MySQL数据库连接字符串构造
		configStr := bytes.Buffer{}
		configStr.WriteString(config.CONF.Db.Username + ":")
		configStr.WriteString(config.CONF.Db.Password + "@tcp(")
		configStr.WriteString(config.CONF.Db.Host + ":" + config.CONF.Db.Port + ")/" + config.CONF.Db.DbName)
		configStr.WriteString(config.CONF.Db.Config)
		// 打开MySQL连接
		db, _ = gorm.Open(mysql.Open(configStr.String()), &gorm.Config{
			Logger: newLogger, PrepareStmt: true, SkipDefaultTransaction: true,
		})
	} else if config.CONF.Db.Type == "postgres" {
		// PostgreSQL数据库连接字符串构造
		configStr := bytes.Buffer{}
		configStr.WriteString("host=" + config.CONF.Db.Host + " ")
		configStr.WriteString("port=" + config.CONF.Db.Port + " ")
		configStr.WriteString("user=" + config.CONF.Db.Username + " ")
		configStr.WriteString("password=" + config.CONF.Db.Password + " ")
		configStr.WriteString("dbname=" + config.CONF.Db.DbName + " ")
		configStr.WriteString("" + config.CONF.Db.Config + " ")
		// 打开PostgreSQL连接
		db, _ = gorm.Open(postgres.Open(configStr.String()), &gorm.Config{
			Logger: newLogger, PrepareStmt: true, SkipDefaultTransaction: true,
		})
	} else {
		// 默认使用SQLite数据库
		db, _ = gorm.Open(sqlite.Open(data.DbPath), &gorm.Config{
			Logger: newLogger, PrepareStmt: true, SkipDefaultTransaction: true,
		})
	}

	// 获取底层SQL数据库连接
	dbSql, err := db.DB()
	if err != nil {
		logger.LOG.Errorf("获取数据库连接失败: %s", err.Error())
		return
	}
	logger.LOG.Debug("database connection successfully")

	// 设置连接池参数
	dbSql.SetMaxIdleConns(config.CONF.Db.MaxIdleConns)                                // 设置最大空闲连接数
	dbSql.SetMaxOpenConns(config.CONF.Db.MaxOpenConns)                                // 设置最大打开连接数
	dbSql.SetConnMaxLifetime(time.Duration(config.CONF.Db.MaxLifeTime) * time.Second) // 设置连接最大生存时间

	// 自动迁移数据表结构，确保模型对应的数据表存在且结构正确
	db.AutoMigrate(
		&basic.User{},        // 用户表
		&basic.ProjectDir{},  // 项目目录表
		&sflow.SFlow{},       // 流程配置表
		&sflow.SFlowLog{},    // 流程日志表
		&nas.Webdav{},        // WebDAV配置表
		&nas.SambaShare{},    // Samba共享配置表
		&nas.ExternalNas{},   // 外部存储配置表
		&nas.NfsShare{},      // NFS共享配置表
		&scheduled.SchTask{}, // 计划任务表
		&log.SchLog{},        // 计划任务日志表
	)
	logger.LOG.Debug("database AutoMigrate successfully")

	// 下面是被注释掉的回调钩子注册代码
	// db.Callback().Create().Before("gorm:create").Register("created_at", createdAt)
	// db.Callback().Create().Before("gorm:create").Register("updated_at", updatedAt)
	// db.Callback().Create().Before("gorm:update").Register("updated_at", updatedAt)
	// db.Callback().Query().Before("gorm:query").Register("my_plugin:before_query", beforeQuery)

	// 初始化gplus
	// gplus.Init(db)

	// 将数据库连接保存到全局变量中，方便其他包使用
	global.DB = db
	logger.LOG.Info("init database successfully")
}

// 以下是被注释掉的时间戳自动更新函数
// func createdAt(db *gorm.DB) {
//   db.Statement.SetColumn("created_at", base.LocalTime{}.Now())
// }
// func updatedAt(db *gorm.DB) {
//   db.Statement.SetColumn("updated_at", base.LocalTime{}.Now())
// }
// func beforeQuery(db *gorm.DB) {
//   db.Statement.SetColumn("updated_at", base.LocalTime{}.Now())
// }
