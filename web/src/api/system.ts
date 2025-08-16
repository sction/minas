import ajax, { Result } from '@/api/ajax'

// 系统版本信息接口
export interface SystemVersion {
    version: string;
    goVersion: string;
    test: string;
}

// 系统环境信息接口
export interface SystemEnvironment {
    isContainer: boolean;
    nfsServiceAvailable: boolean;
    os: string;
    arch: string;
    goVersion: string;
    version: string;
    containerType?: string;
    message?: string;
}

// 系统状态检查接口
export interface SystemState {
    fresh: boolean;
}

const baseUrl = '/system';

export class SystemApi {

    // 获取系统环境信息
    getEnvironment() {
        return ajax.get<Result<SystemEnvironment>>(baseUrl + '/environment')
    }

    // 检查系统状态
    checkState() {
        return ajax.get<Result<SystemState>>(baseUrl + '/check-state')
    }

    // 系统初始化
    init(user: any) {
        return ajax.post<Result<any>>(baseUrl + '/init', user)
    }
}

export default new SystemApi()
