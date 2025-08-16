package db

import (
	"log"
	"server/core/app/request" // 导入请求处理模块
	"server/utils"            // 导入工具函数
	"server/utils/global"     // 导入全局变量
	"server/utils/logger"
	"strings"

	"gorm.io/gorm"        // 导入GORM ORM库
	"gorm.io/gorm/clause" // 导入GORM子句处理
)

// SQL语句相关常量
var (
	SQL_SPACE = " "   // SQL空格
	SQL_AND   = "AND" // SQL与运算符
	SQL_OR    = "OR"  // SQL或运算符
	SQL_BEGIN = "("   // SQL开始括号
	SQL_END   = ")"   // SQL结束括号
)

// SetOperatorUID 设置操作人ID
func (baseModel *BaseModel[T]) SetOperatorUID(uid uint) {
	baseModel.UpdatedBy = uid
}

// Save 保存实体（自动判断是创建还是更新）
// 如果ID大于0且记录存在则更新，否则创建新记录
func (BaseModel[T]) Save(entity *T, columns ...string) error {
	id := (*entity).GetID()
	if id > 0 {
		var tmp T
		err := global.DB.Model(tmp).Take(&tmp, id).Error
		if tmp.GetID() == 0 || (err != nil && err == gorm.ErrRecordNotFound) {
			// 记录不存在，执行新增操作
			return (*entity).Create(entity)
		} else {
			logger.LOG.Debugf("更新记录ID: %d, columns: %v\n", id, columns)
			// 记录存在，执行更新操作
			return (*entity).Update(entity, columns...)
		}
	} else {
		// ID不存在，执行新增操作
		return (*entity).Create(entity)
	}
}

// Create 创建新记录
func (BaseModel[T]) Create(entity *T) error {
	return global.DB.Model(entity).Create(entity).Error
}

// Update 更新已有记录
// 可指定要更新的列，如果不指定则更新所有列
func (BaseModel[T]) Update(entity *T, columns ...string) error {
	if len(columns) > 0 {
		// 添加更新人字段到更新列表
		columns = append(columns, "updated_by")
		logger.LOG.Debugf("更新记录ID: %d, columns: %v\n", (*entity).GetID(), columns)
		return global.DB.Model(entity).Select(columns).Updates(entity).Error
	} else {
		// 更新所有字段
		return global.DB.Model(entity).Updates(entity).Error
	}
}

// Enable 启用记录（将is_disable字段设置为0）
func (BaseModel[T]) Enable(id any) error {
	var entity T
	return global.DB.Model(&entity).Where(ID_FIELD+" =?", id).Update(IS_DISABLE_FIELD, 0).Error
}

// Disable 禁用记录（将is_disable字段设置为1）
func (BaseModel[T]) Disable(id any) error {
	var entity T
	return global.DB.Model(&entity).Where(ID_FIELD+" =?", id).Update(IS_DISABLE_FIELD, 1).Error
}

// Delete 根据ID删除记录（软删除）
func (BaseModel[T]) Delete(id any) error {
	var entity T
	return global.DB.Model(entity).Delete(&entity, id).Error
}

// Load 根据ID加载单个记录
func (BaseModel[T]) Load(id any) (T, error) {
	var entity T
	err := global.DB.Model(entity).Take(&entity, id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return entity, err
	}
	return entity, nil
}

// Count 计算符合条件的记录总数
func (BaseModel[T]) Count() (int64, error) {
	var entity T
	var count int64
	err := global.DB.Model(entity).Count(&count).Error
	return count, err
}
func (BaseModel[T]) CountEnable() (int64, error) {
	var entity T
	var count int64
	err := global.DB.Model(entity).Where(IS_DISABLE_FIELD+" = ?", 0).Count(&count).Error
	return count, err
}

// CountQuery 根据查询条件计算记录数
func (BaseModel[T]) CountQuery(query request.PageQuery) (int64, error) {
	var entity T
	countDb := global.DB.Model(entity)
	countDb = countDb.Select("count(ID)")
	if utils.IsEmpty(query.Filters) {
		query.Filters = []request.QueryFilter{}
	}
	var count int64
	// 应用查询条件
	countDb = QueryWhere(countDb, query.Filters)
	err := countDb.Count(&count).Error
	return count, err
}

// ListQuery 根据查询条件获取列表
func (BaseModel[T]) ListQuery(query request.PageQuery) (list []T, err error) {
	log.Printf("query:%v\n", query)
	var entity T
	modelDb := global.DB.Model(entity)

	// 设置查询字段
	if utils.IsEmpty(query.Columns) {
		modelDb = modelDb.Select("*")
	} else {
		modelDb = modelDb.Select(query.Columns)
	}

	// 设置查询条件
	if utils.IsEmpty(query.Filters) {
		query.Filters = []request.QueryFilter{}
	}
	modelDb = QueryWhere(modelDb, query.Filters)

	// 设置排序
	if !utils.IsEmpty(query.Sorts) {
		for _, sort := range query.Sorts {
			if sort[0:1] == "-" {
				// 降序排序，格式如 "-created_at"
				modelDb = modelDb.Order(clause.OrderByColumn{Column: clause.Column{Name: sort[1:]}, Desc: true})
			} else {
				// 升序排序
				modelDb = modelDb.Order(sort)
			}
		}
	}

	// 设置分页
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Size > 0 {
		modelDb = modelDb.Limit(query.Size).Offset((query.Page - 1) * query.Size)
	}

	// 执行查询
	err = modelDb.Find(&list).Error
	return list, err
}

// List 执行分页查询并返回数据和总数
func (BaseModel[T]) List(query request.PageQuery) (list []T, count int64, err error) {
	var entity T
	// 获取数据列表
	list, err = entity.ListQuery(query)
	if err == nil {
		// 获取数据总数
		count, err = entity.CountQuery(query)
		return list, count, err
	} else {
		return list, 0, err
	}
}

// QueryWhere 根据过滤条件构建查询条件
// 支持条件组合和复杂查询条件
func QueryWhere(db *gorm.DB, filters []request.QueryFilter) *gorm.DB {
	if utils.IsEmpty(filters) {
		return db
	} else {
		for _, filter := range filters {
			if isValid(filter) {
				// 处理条件组
				if !utils.IsEmpty(filter.Filters) {
					if filter.Or {
						// 使用OR连接多个条件
						db.Or(QueryWhere(db, filter.Filters))
					} else {
						// 使用AND连接多个条件
						db.Where(QueryWhere(db, filter.Filters))
					}
				} else {
					// 处理单个条件
					column := filter.Column
					operator := filter.Operator
					value := filter.Value

					if operator == "$nil" {
						// 处理空值条件
						if filter.Or {
							db.Or(column)
						} else {
							db.Where(column)
						}
					} else if strings.Contains(operator, "$column") {
						// 处理列与列比较的条件
						operator = strings.ReplaceAll(operator, "\\$column", column)
						if filter.Or {
							db.Or(column+SQL_SPACE+operator, value)
						} else {
							db.Where(column+SQL_SPACE+operator, value)
						}
					} else {
						// 处理普通条件
						if filter.Or {
							db.Or(column+SQL_SPACE+operator, value)
						} else {
							db.Where(column+SQL_SPACE+operator, value)
						}
					}
				}
			}
		}
		return db
	}
}

// isValid 检查过滤条件是否有效
func isValid(filter request.QueryFilter) bool {
	if !utils.IsEmpty(filter.Column) && !utils.IsEmpty(filter.Operator) {
		// 单个条件有效
		return true
	} else if !utils.IsEmpty(filter.Filters) { // 检查条件组
		for _, filter := range filter.Filters {
			// 任何子条件有效则整体有效
			if isValid(filter) {
				return true
			}
		}
		return false
	} else {
		return false
	}
}
