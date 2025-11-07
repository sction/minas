import ajax, { Result, SearchArgs, SearchResult } from '@/api/ajax'

export interface NodeItem {
  id: number
  name: string
  ip_address: string
  ssh_port: number
  username: string
  password?: string
  private_key?: string
  private_key_pass?: string
  tags: string
  os_type: string
  os_version: string
  hostname: string
  timezone: string
  last_connected?: string
  status: number
  remark: string
  is_disable: number
  created_at: string
  updated_at: string
  created_by_name?: string
  updated_by_name?: string
}



export interface NodeTag {
  name: string
  color: string
}

export interface BatchOperationResult {
  node_id: number
  name: string
  success: boolean
  output?: string
  error?: string
  data?: any
}

const baseUrl = '/depops/node'
// 节点管理API
export class NodeApi {

  // 获取节点列表
  search(args: SearchArgs) {
    return ajax.search<NodeItem>(`${baseUrl}/list`, args)
  }

  // 获取节点详情
  find(id: number) {
    return ajax.get<NodeItem>(`${baseUrl}/load/${id}`)
  }

  // 保存节点
  save(data: Partial<NodeItem>) {
    return ajax.post<any>(`${baseUrl}/save`, data)
  }

  // 删除节点
  delete(id: number) {
    return ajax.post<any>(`${baseUrl}/delete/${id}`)
  }

  // 启用节点
  enable(id: number) {
    return ajax.post<any>(`${baseUrl}/enable/${id}`)
  }

  // 禁用节点
  disable(id: number) {
    return ajax.post<any>(`${baseUrl}/disable/${id}`)
  }

  // 测试节点连接
  testConnection(id: number) {
    return ajax.post<any>(`${baseUrl}/test-connection/${id}`)
  }

  // 测试节点连接（使用数据）
  testConnectionData(data: Partial<NodeItem>) {
    return ajax.post<any>(`${baseUrl}/test-connection-data`, data)
  }


  // 获取所有标签
  getTags() {
    return ajax.get<string[]>(`${baseUrl}/tags`)
  }

  // 执行命令
  executeCommand(id: number, command: string) {
    return ajax.post<{ output: string; error?: string }>(`${baseUrl}/execute-command/${id}`, { command })
  }
}

// 批量操作API
export class BatchApi {
  private basePath = '/depops/batch'

  // 批量重启节点
  restartNodes(nodeIds: (string | number)[]) {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/restart`, { node_ids: nodeIds })
  }

  // 批量管理认证公钥
  manageKeys(nodeIds: (string | number)[], publicKey: string, action: 'add' | 'remove') {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/manage-keys`, {
      node_ids: nodeIds,
      public_key: publicKey,
      action
    })
  }

  // 批量设置时区
  setTimezone(nodeIds: (string | number)[], timezone: string) {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/set-timezone`, {
      node_ids: nodeIds,
      timezone
    })
  }

  // 批量管理防火墙规则
  manageFirewall(nodeIds: (string | number)[], action: string, ports?: number[], ips?: string[]) {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/manage-firewall`, {
      node_ids: nodeIds,
      action,
      ports,
      ips
    })
  }

  // 批量执行命令
  executeCommand(nodeIds: (string | number)[], command: string) {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/execute-command`, {
      node_ids: nodeIds,
      command
    })
  }

  // 批量上传文件
  uploadFile(nodeIds: (string | number)[], localPath: string, remotePath: string, permissions?: string) {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/upload-file`, {
      node_ids: nodeIds,
      local_path: localPath,
      remote_path: remotePath,
      permissions
    })
  }

  // 批量安装软件包
  installPackages(nodeIds: (string | number)[], packages: string[], action: 'install' | 'remove') {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/install-packages`, {
      node_ids: nodeIds,
      packages,
      action
    })
  }

  // 批量发送Docker镜像
  deployDockerImage(nodeIds: (string | number)[], imagePath: string, imageName?: string) {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/deploy-docker-image`, {
      node_ids: nodeIds,
      image_path: imagePath,
      image_name: imageName
    })
  }

  // 扫描目录中的Docker镜像文件
  scanDockerImages(directoryPath: string) {
    return ajax.post<string[]>(`${this.basePath}/scan-docker-images`, {
      directory_path: directoryPath
    })
  }

  // 批量发送目录中的Docker镜像文件
  deployDockerImagesFromDirectory(nodeIds: (string | number)[], directoryPath: string) {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/deploy-docker-images-from-directory`, {
      node_ids: nodeIds,
      directory_path: directoryPath
    })
  }

  // 批量执行脚本
  executeScript(nodeIds: (string | number)[], scriptPath: string, scriptType: string = 'bash') {
    return ajax.post<BatchOperationResult[]>(`${this.basePath}/execute-script`, {
      node_ids: nodeIds,
      script_path: scriptPath,
      script_type: scriptType
    })
  }

 
}

export const nodeApi = new NodeApi()
export const batchApi = new BatchApi()

export default {
  node: nodeApi,
  batch: batchApi,
}
