import ajax, { Result, SearchArgs, SearchResult, SetStatusArgs } from '@/api/ajax'


export interface FileItem {
    name: string
    isDir: boolean
    size?: number
    path: string
    createdAt?: string
    isLink: boolean
    linkTo?: string
}


const baseUrl = '/fs';
export class FileSystemApi {
    mkdir(path: string, name: string) {
        return ajax.post<Result<any>>(baseUrl + '/mkdir', { path, name })
    }

    // 获取文件列表
    list(path: string = '/', fstype: string = '', suffix: string = '') {
        return ajax.get<FileItem[]>(`${baseUrl}/list?path=${encodeURIComponent(path)}&fstype=${fstype}&suffix=${suffix}`)
    }

}

export default new FileSystemApi