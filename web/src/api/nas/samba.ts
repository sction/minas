import ajax, { Result, SearchArgs, SearchResult } from '@/api/ajax'

// Samba 共享配置接口
export interface SambaShare {
    id: number;             // ID
    name: string;           // 共享名称
    path: string;           // 共享路径
    comment: string;        // 共享描述
    browseable: boolean;    // 是否可浏览
    readOnly: boolean;      // 是否只读
    guestOk: boolean;       // 是否允许来宾访问
    validUsers: string;     // 有效用户列表
    writeList: string;      // 写权限用户列表
    readList: string;       // 读权限用户列表
    createMask: string;     // 创建文件权限掩码
    directoryMask: string;  // 创建目录权限掩码
    forceUser: string;      // 强制用户
    forceGroup: string;     // 强制组
    availableSpace: string; // 可用空间限制
    options: string;        // 其他选项
    remark: string;         // 备注
    is_disable: number;     // 是否禁用：0-启用，1-禁用
    created_at: string;
    updated_at: string;
    created_by: number;
    updated_by: number;
}

// Samba 服务状态接口
export interface SambaServiceStatus {
    installed: boolean;     // 是否已安装
    running: boolean;       // 是否正在运行
    enabled: boolean;       // 是否开机自启
    available?: boolean;    // 是否可用
    message?: string;       // 状态消息
}

const baseUrl = '/nas/samba';

export class SambaApi {
    // 共享管理
    save(share: SambaShare) {
        return ajax.post<Result<any>>(baseUrl + '/save', share)
    }

    search(args: SearchArgs) {
        return ajax.get<SearchResult<SambaShare>>(baseUrl + '/list', args)
    }

    load(id: number) {
        return ajax.get<SambaShare>(baseUrl + '/load/' + id)
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
        return ajax.get<SambaServiceStatus>(baseUrl + '/service/status')
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

    // 配置管理
    reloadConfig() {
        return ajax.post<Result<any>>(baseUrl + '/reload-config')
    }

    // 目录选择
    listDir(path: string) {
        return ajax.get<any>(baseUrl + '/list-dir', { path })
    }

    // 获取目录所有者
    getDirOwner(path: string) {
        return ajax.get<any>(baseUrl + '/dir-owner', { path })
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

    // smb.conf配置管理
    getSmbConfig() {
        return ajax.get<Result<string>>(baseUrl + '/smb-config')
    }

    backupSmbConfig() {
        return ajax.post<Result<any>>(baseUrl + '/backup-smb')
    }

    // 测试配置
    testConfig() {
        return ajax.post<Result<any>>(baseUrl + '/test-config')
    }

    // 用户管理
    listSambaUsers() {
        return ajax.get<SearchResult<any>>(baseUrl + '/users')
    }

    addSambaUser(username: string, password: string) {
        return ajax.post<Result<any>>(baseUrl + '/users/add', { username, password })
    }

    deleteSambaUser(username: string) {
        return ajax.post<Result<any>>(baseUrl + '/users/delete', { username })
    }

    changeSambaUserPassword(username: string, password: string) {
        return ajax.post<Result<any>>(baseUrl + '/users/password', { username, password })
    }
}

export default new SambaApi()