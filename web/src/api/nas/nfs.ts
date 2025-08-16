import ajax, { Result, SearchArgs, SearchResult } from '@/api/ajax'

// NFS 共享配置接口
export interface NfsShare {
    id: number;             // 修改为number类型以匹配后端uint类型
    name: string;           // 共享名称
    path: string;           // 共享路径
    clientIp: string;       // 客户端IP
    permission: string;     // 读写权限：ro/rw
    sync: string;           // 同步模式：sync/async
    rootSquash: string;     // Root权限映射
    subtreeCheck: string;   // 子树检查：subtree_check/no_subtree_check
    options: string;        // 其他选项
    remark: string;         // 备注
    is_disable: number;     // 是否禁用：0-启用，1-禁用
    created_at: string;
    updated_at: string;
    created_by: number;
    updated_by: number;
}

// NFS 服务状态接口
export interface NfsServiceStatus {
    installed: boolean;     // 是否已安装
    running: boolean;       // 是否正在运行
    enabled: boolean;       // 是否开机自启
    available?: boolean;    // 是否可用（新增字段）
    message?: string;       // 状态消息（新增字段）
}

const baseUrl = '/nas/nfs';

export class NfsApi {
    // 共享管理
    save(share: NfsShare) {
        return ajax.post<Result<any>>(baseUrl + '/save', share)
    }

    search(args: SearchArgs) {
        return ajax.get<SearchResult<NfsShare>>(baseUrl + '/list', args)
    }

    load(id: number) {
        return ajax.get<NfsShare>(baseUrl + '/load/' + id)
    }

    delete(id: number) {
        return ajax.post<Result<any>>(baseUrl + '/delete/' + id)
    }

    // 启用/禁用共享
    enable(id: number) {
        return ajax.post<Result<any>>(baseUrl + '/enable/' + id)
    }

    disable(id: number) {
        return ajax.post<Result<any>>(baseUrl + '/disable/' + id)
    }

    // 服务管理
    getServiceStatus() {
        return ajax.get<NfsServiceStatus>(baseUrl + '/service/status')
    }

    startService() {
        return ajax.post<Result<any>>(baseUrl + '/service/start')
    }

    stopService() {
        return ajax.post<Result<any>>(baseUrl + '/service/stop')
    }

    restartService() {
        return ajax.post<Result<any>>(baseUrl + '/service/restart')
    }

    enableAutoStart() {
        return ajax.post<Result<any>>(baseUrl + '/service/enable')
    }

    disableAutoStart() {
        return ajax.post<Result<any>>(baseUrl + '/service/disable')
    }

    // 配置导出
    reloadExports() {
        return ajax.post<Result<any>>(baseUrl + '/reload-exports')
    }

    // 目录选择
    listDir(path: string) {
        return ajax.get<any>(baseUrl + '/list-dir', { path })
    }

    // 同步功能
    getSyncStatus() {
        return ajax.get<Result<any>>(baseUrl + '/sync-status')
    }

    syncFromServer() {
        return ajax.post<Result<any>>(baseUrl + '/sync-from-server')
    }

    syncToServer() {
        return ajax.post<Result<any>>(baseUrl + '/sync-to-server')
    }

    // exports配置管理
    getExportsConfig() {
        return ajax.get<Result<string>>(baseUrl + '/exports-config')
    }

    backupExportsConfig() {
        return ajax.post<Result<any>>(baseUrl + '/backup-exports')
    }
}

export default new NfsApi()