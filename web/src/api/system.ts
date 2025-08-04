import ajax, { Result } from '@/api/ajax'

// 统计数据接口
export interface DashboardStats {
    projectCount: number;      // 项目数
    webdavCount: number;       // WebDAV服务数
    externalStorageCount: number; // 外部存储数
    scheduledTaskCount: number;   // 计划任务数
}

// 任务类型统计
export interface TaskTypeStats {
    type: string;   // 任务类型
    count: number;  // 数量
    label: string;  // 显示标签
}

// 任务执行统计
export interface TaskExecutionStats {
    success: number;  // 成功数
    failed: number;   // 失败数
    running: number;  // 运行中数
    total: number;    // 总数
}

// 任务执行趋势数据
export interface TaskTrendData {
    date: string;     // 日期
    success: number;  // 成功数
    failed: number;   // 失败数
    total: number;    // 总数
}

// 项目任务统计
export interface TaskProjectStats {
    projectId: string;   // 项目ID
    projectName: string; // 项目名称
    taskCount: number;   // 任务数量
}

const baseUrl = '/system';
export class SystemApi {
    checkState() {
        return ajax.get<any>(baseUrl+'/check-state')
    }

    init(user: any) {
        return ajax.post<Result<any>>(baseUrl+'/init', user)
    }

    version() {
        return ajax.get<any>(baseUrl+'/version')
    }

    // 获取仪表板统计数据
    getDashboardStats() {
        return ajax.get<Result<DashboardStats>>(baseUrl + '/dashboard/stats')
    }

    // 获取任务类型统计
    getTaskTypeStats() {
        return ajax.get<Result<TaskTypeStats[]>>(baseUrl + '/dashboard/task-types')
    }

    // 获取任务执行统计
    getTaskExecutionStats() {
        return ajax.get<Result<TaskExecutionStats>>(baseUrl + '/dashboard/task-execution')
    }

    // 获取任务执行趋势（近一个月）
    getTaskTrend(days: number = 30) {
        return ajax.get<Result<TaskTrendData[]>>(baseUrl + '/dashboard/task-trend', { days })
    }

    // 获取项目任务统计
    getTaskProjectStats() {
        return ajax.get<Result<TaskProjectStats[]>>(baseUrl + '/dashboard/task-projects')
    }
}

export default new SystemApi
