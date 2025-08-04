import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { store } from "../store";
import { router } from "../router/router";
import { Mutations } from "@/store/mutations";
import { t, te } from '@/locales';

// export interface AjaxOptions {
// }

export interface Result<T> {
    code: number;
    msg?: string;
    data?: T;
}

export interface SearchResult<T> {
    code: number;
    msg: string;
    data: T[];
    total: number;
}

export interface SearchArgs {
    filters?: string;
    columns?: string
    sorts?: string
    page: number;
    size: number;
}
export interface SetStatusArgs {
    id: string;
    status: number;
}
class Ajax {
    private ajax: AxiosInstance;

    constructor() {
        this.ajax = axios.create({
            baseURL: import.meta.env.VITE_API_URL,
            timeout: 30000,
            // withCredentials: true,
        })

        this.ajax.interceptors.request.use(
            (config: any) => {
                if (store.state.user?.token) {
                    config.headers.Authorization = "Bearer " + store.state.user.token
                }
                // store.commit(Mutations.SetAjaxLoading, true);
                return config;
            },
            (error: any) => {
                return Promise.reject(error);
            }
        )

        this.ajax.interceptors.response.use(
            (response: any) => {
                // store.commit(Mutations.SetAjaxLoading, false);
                if (response.data.code == null || response.data.code == undefined) {
                    console.log(response)
                    console.log(response.data)
                    window.message.error("接口：" + response.config.url + "，响应数据格式错误，请联系管理员！", { duration: 0, closable: true, })
                }
                
                // 检查是否为仪表板统计API，如果是则不显示成功提示
                const isDashboardAPI = response.config.url && response.config.url.includes('/dashboard/');
                
                if (response.data.msg != null && response.data.msg != undefined && response.data.msg != "") {
                    //处理提示信息
                    if (response.data.code == 200) {
                        // 仪表板统计API不显示成功提示
                        if (!isDashboardAPI) {
                            window.message.success(response.data.msg)
                        }
                    } else if (response.data.code == 4000 || response.data.code == 400 || response.data.code == 401 || response.data.code == 403 || response.data.code == 404 || response.data.code == 500) {
                        window.message.error(response.data.msg, { duration: 10000, closable: true, })
                    } else {
                        window.message.info(response.data.msg, { duration: 6000, closable: true, })
                    }
                } else {
                    // 根据状态码处理提示信息
                    if (response.data.code == 4002) {
                        //存在关联数据 一般用于删除数据时
                        window.message.info("存在关联数据(" + response.data.data + "条)，请先删除关联数据后再尝试删除当前数据！", { duration: 6000, closable: false, })
                    } else if (response.data.code == 4401) {
                        window.message.info(response.data.msg, { duration: 6000, closable: false, })
                    } else if (response.data.code > 4000) {
                        console.warn("未处理的错误：", response.data)
                    }
                }
                return response;
            },
            (error: any) => {
                if (this.handleError(error)) {
                    // Stop Promise chain
                    return new Promise(() => { })
                } else {
                    return Promise.reject(error)
                }
            }
        )
    }

    private handleError(error: any): boolean {
        if (error.response) {
            switch (error.response.status) {
                case 401:
                    store.commit(Mutations.Logout);
                    if (error.config.method === "get") {
                        router.replace({
                            name: 'login',
                            query: {
                                redirect: router.currentRoute.value.fullPath
                            }
                        });
                    } else {
                        this.showError(error)
                    }
                    return true
                case 403:
                    router.replace("/403");
                    return true
                case 404:
                    router.replace("/404");
                    return true
                case 500:
                    this.showError(error)
                    return true
            }
        } else {
            window.message.error(error.message, { duration: 0, closable: true, });
        }
        return false
    }

    private showError(error: any) {
        const code = error.response.data?.code || 1;
        const msg = te('errors.' + code) ? t('errors.' + code) : error.response.data?.msg || error.message;
        window.message.error(msg, { duration: 0, closable: true, });
    }

    async get<T>(url: string, args?: any, config?: AxiosRequestConfig): Promise<Result<T>> {
        config = { ...config, params: args }
        const r = await this.ajax.get<Result<T>>(url, config);
        return r.data;
    }
    async search<T>(url: string, args?: any, config?: AxiosRequestConfig): Promise<SearchResult<T>> {
        config = { ...config, params: args }
        const r = await this.ajax.get<SearchResult<T>>(url, config);
        return r.data;
    }
    async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<Result<T>> {
        config = { ...config, headers: { 'Content-Type': 'application/json' } }
        const r = await this.ajax.post<Result<T>>(url, data, config);
        return r.data;
    }
    async request<T>(config: AxiosRequestConfig): Promise<Result<T>> {
        const r = await this.ajax.request<Result<T>>(config);
        return r.data;
    }
    async orequest(config: AxiosRequestConfig): Promise<AxiosResponse> {
        // 确保有正确的this上下文
        if (!this.ajax) {
            throw new Error('Ajax实例未初始化');
        }
        return this.ajax.request(config);
    }
}

// 创建单例实例并绑定方法上下文
const ajaxInstance = new Ajax();
// 绑定所有方法的上下文
Object.getOwnPropertyNames(Ajax.prototype).forEach(method => {
    if (method !== 'constructor' && typeof ajaxInstance[method as keyof Ajax] === 'function') {
        (ajaxInstance[method as keyof Ajax] as Function) = (ajaxInstance[method as keyof Ajax] as Function).bind(ajaxInstance);
    }
});

export default ajaxInstance;