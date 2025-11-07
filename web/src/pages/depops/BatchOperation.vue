<template>
  <x-page-header>
    <template #action>
    </template>
  </x-page-header>

  <n-space class="page-body" vertical :size="12">
    <!-- 节点选择 -->
    <n-card title="选择节点" :bordered="false" size="small">
      <n-space vertical :size="12">
        <n-space :size="12">
          <n-input v-model:value="searchKeyword" placeholder="搜索节点" clearable style="width: 300px;" />
          <n-select v-model:value="selectedTags" placeholder="按标签筛选" multiple clearable :options="tagOptions"
            style="width: 200px;" /> 
        </n-space>

        <n-data-table :columns="nodeColumns" :data="availableNodes" :row-key="(row: any) => row.id" size="small"
          :checked-row-keys="selectedNodeIds" @update:checked-row-keys="handleNodeSelection" max-height="300px" />

        <n-alert v-if="selectedNodeIds.length > 0" type="info">
          已选择 {{ selectedNodeIds.length }} 个节点
        </n-alert>
      </n-space>
    </n-card>

    <!-- 操作选择 -->
    <n-card title="批量操作" :bordered="false" size="small">
      <n-tabs v-model:value="activeTab" type="card">
        <!-- 系统管理 -->
        <n-tab-pane name="system" tab="系统管理">
          <n-space vertical :size="16">
            <n-card title="重启节点" size="small" :bordered="false">
              <n-space vertical :size="12">
                <n-alert type="warning">
                  重启操作将使节点暂时不可用，请确认操作。
                </n-alert>
                <n-button type="warning" @click="() => executeBatchOperation('restart')"
                  :disabled="selectedNodeIds.length === 0" :loading="operationLoading">
                  重启选中节点
                </n-button>
              </n-space>
            </n-card>

            <n-card title="设置时区" size="small" :bordered="false">
              <n-space vertical :size="12">
                <n-select v-model:value="timezoneForm.timezone" :options="timezoneOptions" placeholder="选择时区"
                  style="width: 300px;" />
                <n-button type="primary" @click="() => executeBatchOperation('set_timezone')"
                  :disabled="selectedNodeIds.length === 0 || !timezoneForm.timezone" :loading="operationLoading">
                  设置时区
                </n-button>
              </n-space>
            </n-card>
          </n-space>
        </n-tab-pane>

        <!-- 安全管理 -->
        <n-tab-pane name="security" tab="安全管理">
          <n-space vertical :size="16">
            <n-card title="SSH密钥管理" size="small" :bordered="false">
              <n-space vertical :size="12">
                <n-radio-group v-model:value="keyForm.action">
                  <n-radio value="add">添加公钥</n-radio>
                  <n-radio value="remove">删除公钥</n-radio>
                </n-radio-group>
                <n-input v-model:value="keyForm.publicKey" type="textarea" placeholder="请输入SSH公钥内容" :rows="4" />
                <n-button type="primary" @click="() => executeBatchOperation('manage_keys')"
                  :disabled="selectedNodeIds.length === 0 || !keyForm.publicKey" :loading="operationLoading">
                  {{ keyForm.action === 'add' ? '添加' : '删除' }}公钥
                </n-button>
              </n-space>
            </n-card>

            <n-card title="防火墙管理" size="small" :bordered="false">
              <n-space vertical :size="12">
                <n-select v-model:value="firewallForm.action" :options="firewallActions" placeholder="选择防火墙操作"
                  style="width: 200px;" />
                <n-input v-if="firewallForm.action?.includes('port')" v-model:value="firewallForm.ports"
                  placeholder="端口号，多个用逗号分隔" style="width: 300px;" />
                <n-input v-if="firewallForm.action?.includes('ip')" v-model:value="firewallForm.ips"
                  placeholder="IP地址，多个用逗号分隔" style="width: 300px;" />
                <n-button type="primary" @click="() => executeBatchOperation('manage_firewall')"
                  :disabled="selectedNodeIds.length === 0 || !firewallForm.action" :loading="operationLoading">
                  执行防火墙操作
                </n-button>
              </n-space>
            </n-card>
          </n-space>
        </n-tab-pane>

        <!-- 软件管理 -->
        <n-tab-pane name="software" tab="软件管理">
          <n-space vertical :size="16">
            <n-card title="软件包管理" size="small" :bordered="false">
              <n-space vertical :size="12">
                <n-radio-group v-model:value="packageForm.action">
                  <n-radio value="install">安装软件包</n-radio>
                  <n-radio value="remove">卸载软件包</n-radio>
                </n-radio-group>
                <n-input v-model:value="packageForm.packages" placeholder="软件包名称，多个用空格分隔" style="width: 400px;" />
                <n-button type="primary" @click="() => executeBatchOperation('install_packages')"
                  :disabled="selectedNodeIds.length === 0 || !packageForm.packages" :loading="operationLoading">
                  {{ packageForm.action === 'install' ? '安装' : '卸载' }}软件包
                </n-button>
              </n-space>
            </n-card>

            <n-card title="Docker镜像部署" size="small" :bordered="false">
              <n-space vertical :size="12">
                <file-select v-model="dockerForm.imagePath" placeholder="点击选择文件或目录" />
                <n-button type="primary" @click="() => executeBatchOperation('deploy_docker_image')"
                  :disabled="isDockerDeployDisabled" :loading="operationLoading">
                  部署Docker镜像
                </n-button>
              </n-space>
            </n-card>
          </n-space>
        </n-tab-pane>

        <!-- 命令执行 -->
        <n-tab-pane name="command" tab="命令执行">
          <n-space vertical :size="16">
            <n-card title="执行命令" size="small" :bordered="false">
              <n-space vertical :size="12">
                <n-input v-model:value="commandForm.command" type="textarea" placeholder="请输入要执行的命令" :rows="4" />
                <n-button type="primary" @click="() => executeBatchOperation('execute_command')"
                  :disabled="selectedNodeIds.length === 0 || !commandForm.command" :loading="operationLoading">
                  执行命令
                </n-button>
              </n-space>
            </n-card>

            <n-card title="执行脚本" size="small" :bordered="false">
              <n-space vertical :size="12">
                <n-input v-model:value="scriptForm.scriptPath" placeholder="脚本文件路径" style="width: 400px;" />
                <n-select v-model:value="scriptForm.scriptType" :options="scriptTypes" placeholder="脚本类型"
                  style="width: 150px;" />
                <n-button type="primary" @click="() => executeBatchOperation('execute_script')"
                  :disabled="selectedNodeIds.length === 0 || !scriptForm.scriptPath" :loading="operationLoading">
                  执行脚本
                </n-button>
              </n-space>
            </n-card>
          </n-space>
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 执行结果 -->
    <n-card v-if="operationResults.length > 0" title="执行结果" :bordered="false" size="small">
      <n-data-table :columns="resultColumns" :data="operationResults" size="small" :pagination="false"
        max-height="400px" />
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NSpace,
  NCard,
  NButton,
  NIcon,
  NInput,
  NSelect,
  NDataTable,
  NAlert,
  NTabs,
  NTabPane,
  NRadio,
  NRadioGroup,
  useMessage,
  DataTableColumns,
} from 'naive-ui'
import {
  ArrowBackOutline as ArrowBackIcon,
  RefreshOutline as RefreshIcon,
} from '@vicons/ionicons5'
import XPageHeader from '@/components/PageHeader.vue'
import FileSelect from '@/components/fileselect/index.vue'
import depopsApi from '@/api/depops'
import type { NodeItem, BatchOperationResult } from '@/api/depops'
import { renderTag } from '@/utils/render'
import { xxtea } from '@/utils/xxtea'

const message = useMessage()
const router = useRouter();

// 节点选择相关
const searchKeyword = ref('')
const selectedTags = ref<string[]>([])
const availableNodes = ref<NodeItem[]>([])
const selectedNodeIds = ref<(string | number)[]>([])
const tagOptions = ref<{ label: string; value: string }[]>([])

// 操作相关
const activeTab = ref('system')
const operationLoading = ref(false)
const operationResults = ref<BatchOperationResult[]>([])


// 表单数据
const timezoneForm = ref({ timezone: '' })
const keyForm = ref({ action: 'add' as 'add' | 'remove', publicKey: '' })
const firewallForm = ref({ action: '', ports: '', ips: '' })
const packageForm = ref({ action: 'install' as 'install' | 'remove', packages: '' })
const dockerForm = ref({
  imagePath: '',
})
const commandForm = ref({ command: '' })
const scriptForm = ref({ scriptPath: '', scriptType: 'bash' })

// 节点表格列
const nodeColumns: DataTableColumns<NodeItem> = [
  { type: 'selection' },
  { title: 'ID', key: 'id', width: 60 },
  { title: '节点名称', key: 'name', width: 120 },
  { title: 'IP地址', key: 'ip_address', width: 120 },
  {
    title: '状态', key: 'status', width: 80, render: (row: NodeItem) => {
      const statusMap = { 0: { text: "未知", type: "default" }, 1: { text: "在线", type: "success" }, 2: { text: "离线", type: "error" } }
      const status = statusMap[row.status as keyof typeof statusMap]
      return renderTag(status.text, status.type as any)
    }
  },
  {
    title: '标签', key: 'tags', width: 150, render: (row: NodeItem) => {
      if (!row.tags) return '-'
      try {
        const tags = JSON.parse(row.tags)
        if (!Array.isArray(tags) || tags.length === 0) return '-'
        return tags.map(tag => renderTag(tag, 'info'))
      } catch {
        return row.tags || '-'
      }
    }
  },
  { title: '操作系统', key: 'os_info', render: (row: NodeItem) => row.os_type ? `${row.os_type} ${row.os_version}` : "未知" },
]

// 结果表格列
const resultColumns: DataTableColumns<BatchOperationResult> = [
  { title: '节点ID', key: 'node_id', width: 80 },
  { title: '节点名称', key: 'name', width: 150 },
  {
    title: '状态', key: 'success', width: 80, render: (row: BatchOperationResult) =>
      renderTag(row.success ? "成功" : "失败", row.success ? "success" : "error")
  },
  { title: '输出', key: 'output', ellipsis: { tooltip: true } },
  { title: '错误', key: 'error', ellipsis: { tooltip: true } },
]

// 选项数据
const timezoneOptions = [
  { label: 'Asia/Shanghai', value: 'Asia/Shanghai' },
  { label: 'Asia/Tokyo', value: 'Asia/Tokyo' },
  { label: 'Europe/London', value: 'Europe/London' },
  { label: 'America/New_York', value: 'America/New_York' },
  { label: 'UTC', value: 'UTC' },
]

const firewallActions = [
  { label: '开放端口', value: 'add_port' },
  { label: '关闭端口', value: 'remove_port' },
  { label: '添加信任IP', value: 'add_ip' },
  { label: '删除信任IP', value: 'remove_ip' },
]

const scriptTypes = [
  { label: 'Bash', value: 'bash' },
  { label: 'Python', value: 'python' },
  { label: 'Python3', value: 'python3' },
  { label: 'Shell', value: 'sh' },
]

// Docker部署验证
const isDockerDeployDisabled = computed(() => {
  console.log('dockerForm.value.imagePath', dockerForm.value.imagePath)
  return !dockerForm.value.imagePath
})

// 防抖定时器
let searchTimer: NodeJS.Timeout | null = null

// 防抖加载节点
const debouncedLoadNodes = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  searchTimer = setTimeout(() => {
    loadNodes()
  }, 300) // 300ms 防抖延迟
}


// 加载节点列表
const loadNodes = async () => {
  try {
    clearSelection();
    const params: any = { page: 1, size: 1000 }

    // 处理搜索关键词和标签筛选
    const filters: any[] = []

    if (searchKeyword.value?.trim()) {
      // 支持节点名称和IP地址搜索 - 使用OR条件组合
      filters.push({
        filters: [
          {
            column: 'name',
            operator: ' like ?',
            value: `%${searchKeyword.value.trim()}%`
          },
          {
            column: 'ip_address',
            operator: ' like ?',
            value: `%${searchKeyword.value.trim()}%`,
            or: true
          }
        ]
      })
    }

    if (selectedTags.value.length > 0) {
      // 标签筛选 - 每个标签作为一个独立的过滤条件
      selectedTags.value.forEach(tag => {
        filters.push({
          column: 'tags',
          operator: ' like ?',
          value: `%${tag}%`
        })
      })
    }

    if (filters.length > 0) {
      // 使用XXTEA加密过滤条件
      const filtersJson = JSON.stringify(filters)
      params.filters = xxtea.encryptAuto(filtersJson, 'filters')
    }

    const result = await depopsApi.node.search(params)
    availableNodes.value = result.data || []
  } catch (error) {
    console.error('加载节点列表失败:', error)
    message.error('加载节点列表失败')
  }
}

// 加载标签选项
const loadTags = async () => {
  try {
    const result = await depopsApi.node.getTags()
    tagOptions.value = (result.data || []).map((tag: string) => ({
      label: tag,
      value: tag,
    }))
  } catch (error) {
    console.error('加载标签失败:', error)
  }
}

// 处理节点选择
const handleNodeSelection = (keys: (string | number)[]) => {
  selectedNodeIds.value = keys
}

// 清空选择
const clearSelection = () => {
  selectedNodeIds.value = []
  operationResults.value = []
}

// 执行批量操作
const executeBatchOperation = async (operation: string) => {
  if (selectedNodeIds.value.length === 0) {
    message.warning('请先选择要操作的节点')
    return
  }

  try {
    operationLoading.value = true
    let results: BatchOperationResult[] = []

    switch (operation) {
      case 'restart':
        results = (await depopsApi.batch.restartNodes(selectedNodeIds.value)).data || []
        break

      case 'set_timezone':
        if (!timezoneForm.value.timezone) {
          message.warning('请选择时区')
          return
        }
        results = (await depopsApi.batch.setTimezone(selectedNodeIds.value, timezoneForm.value.timezone)).data || []
        break

      case 'manage_keys':
        if (!keyForm.value.publicKey) {
          message.warning('请输入公钥内容')
          return
        }
        results = (await depopsApi.batch.manageKeys(
          selectedNodeIds.value,
          keyForm.value.publicKey,
          keyForm.value.action
        )).data || []
        break

      case 'manage_firewall':
        if (!firewallForm.value.action) {
          message.warning('请选择防火墙操作')
          return
        }
        const ports = firewallForm.value.ports ?
          firewallForm.value.ports.split(',').map(p => parseInt(p.trim())) : undefined
        const ips = firewallForm.value.ips ?
          firewallForm.value.ips.split(',').map(ip => ip.trim()) : undefined
        results = (await depopsApi.batch.manageFirewall(
          selectedNodeIds.value,
          firewallForm.value.action,
          ports,
          ips
        )).data || []
        break

      case 'execute_command':
        if (!commandForm.value.command) {
          message.warning('请输入要执行的命令')
          return
        }
        results = (await depopsApi.batch.executeCommand(selectedNodeIds.value, commandForm.value.command)).data || []
        break

      case 'install_packages':
        if (!packageForm.value.packages) {
          message.warning('请输入软件包名称')
          return
        }
        const packages = packageForm.value.packages.split(' ').filter(p => p.trim())
        results = (await depopsApi.batch.installPackages(
          selectedNodeIds.value,
          packages,
          packageForm.value.action
        )).data || []
        break

      case 'deploy_docker_image':
        if (!dockerForm.value.imagePath) {
          message.warning('请输入目录路径')
          return
        }
        results = (await depopsApi.batch.deployDockerImagesFromDirectory(
          selectedNodeIds.value,
          dockerForm.value.imagePath
        )).data || []
        break

      case 'execute_script':
        if (!scriptForm.value.scriptPath) {
          message.warning('请输入脚本路径')
          return
        }
        results = (await depopsApi.batch.executeScript(
          selectedNodeIds.value,
          scriptForm.value.scriptPath,
          scriptForm.value.scriptType
        )).data || []
        break
    }

    operationResults.value = results

    // 显示操作统计
    const successCount = results.filter(r => r.success).length
    const failCount = results.length - successCount

    if (failCount === 0) {
      message.success(`操作完成！成功 ${successCount} 个节点`)
    } else {
      message.warning(`操作完成！成功 ${successCount} 个，失败 ${failCount} 个节点`)
    }

  } catch (error) {
    console.error('批量操作失败:', error)
  } finally {
    operationLoading.value = false
  }
}

// 监听筛选条件变化，自动搜索
watch([searchKeyword, selectedTags], () => {
  debouncedLoadNodes()
}, { 
  // 深度监听
  deep: true
})

onMounted(() => {
  loadNodes()
  loadTags()
})
</script>

<style scoped>
.page-body {
  height: calc(100vh - 120px);
  overflow: auto;
}
</style>
