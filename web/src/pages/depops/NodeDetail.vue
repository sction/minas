<template>
  <x-page-header>
    <template #action>
      <n-button secondary size="small" @click="router.go(-1)">
        <template #icon>
          <n-icon>
            <arrow-back-icon />
          </n-icon>
        </template>
        返回
      </n-button>
      <n-button secondary size="small" @click="refreshData">
        <template #icon>
          <n-icon>
            <refresh-icon />
          </n-icon>
        </template>
        刷新
      </n-button>
      <n-button secondary size="small" @click="testConnection" :loading="testing">
        <template #icon>
          <n-icon>
            <wifi-icon />
          </n-icon>
        </template>
        测试连接
      </n-button>
    </template>
  </x-page-header>

  <n-space class="page-body" vertical :size="12">
    <n-spin :show="loading">
      <!-- 基本信息 -->
      <n-card title="基本信息" :bordered="false" size="small">
        <n-descriptions :column="3" bordered>
          <n-descriptions-item label="节点名称">{{ nodeData.name }}</n-descriptions-item>
          <n-descriptions-item label="IP地址">{{ nodeData.ip_address }}</n-descriptions-item>
          <n-descriptions-item label="SSH端口">{{ nodeData.ssh_port }}</n-descriptions-item>
          <n-descriptions-item label="用户名">{{ nodeData.username }}</n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-tag :type="getStatusType(nodeData.status)">{{ getStatusText(nodeData.status) }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="最后连接">
            {{ nodeData.last_connected ? formatTime(nodeData.last_connected) : '从未连接' }}
          </n-descriptions-item>
        </n-descriptions>
      </n-card>

      <!-- 系统信息 -->
      <n-card title="系统信息" :bordered="false" size="small">
        <n-descriptions :column="3" bordered>
          <n-descriptions-item label="操作系统">{{ nodeData.os_type || '未知' }}</n-descriptions-item>
          <n-descriptions-item label="系统版本">{{ nodeData.os_version || '未知' }}</n-descriptions-item>
          <n-descriptions-item label="主机名">{{ nodeData.hostname || '未知' }}</n-descriptions-item>
          <n-descriptions-item label="时区">{{ nodeData.timezone || '未知' }}</n-descriptions-item>
          <n-descriptions-item label="标签">
            <n-space size="small">
              <n-tag v-for="tag in getTags()" :bordered="false" :key="tag" type="info" round >
                {{ tag }}
              </n-tag>
              <span v-if="getTags().length === 0">无标签</span>
            </n-space>
          </n-descriptions-item>
          <n-descriptions-item label="备注">{{ nodeData.remark || '无' }}</n-descriptions-item>
        </n-descriptions>
      </n-card>

      <!-- 命令执行 -->
      <n-card title="命令执行" :bordered="false" size="small">
        <n-space vertical :size="12">
          <n-input-group>
            <n-input 
              v-model:value="command" 
              placeholder="输入要执行的命令" 
              @keyup.enter="executeCommand"
              :disabled="executing"
            />
            <n-button type="primary" @click="executeCommand" :loading="executing">执行</n-button>
          </n-input-group>
          
          <n-card v-if="commandHistory.length > 0" title="执行历史" size="small" :bordered="false">
            <n-scrollbar style="max-height: 400px;">
              <n-space vertical :size="8">
                <div v-for="(item, index) in commandHistory" :key="index" class="command-item">
                  <div class="command-header">
                    <n-text strong>$ {{ item.command }}</n-text>
                    <n-text depth="3" style="font-size: 12px;">{{ formatTime(item.timestamp) }}</n-text>
                  </div>
                  <n-code 
                    v-if="item.output" 
                    :code="item.output" 
                    language="bash"
                    style="margin-top: 8px;"
                  />
                  <n-alert 
                    v-if="item.error" 
                    type="error" 
                    :title="item.error"
                    style="margin-top: 8px;"
                  />
                </div>
              </n-space>
            </n-scrollbar>
          </n-card>
        </n-space>
      </n-card>

      <!-- 系统资源监控（预留） -->
      <n-card title="系统监控" :bordered="false" size="small">
        <n-alert type="info">
          系统资源监控功能开发中...
        </n-alert>
      </n-card>
    </n-spin>
  </n-space>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NSpace,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NButton,
  NIcon,
  NSpin,
  NInput,
  NInputGroup,
  NCode,
  NAlert,
  NText,
  NScrollbar,
  useMessage,
} from 'naive-ui'
import {
  ArrowBackOutline as ArrowBackIcon,
  RefreshOutline as RefreshIcon,
  WifiOutline as WifiIcon,
  InformationCircleOutline as InformationIcon,
} from '@vicons/ionicons5'
import XPageHeader from '@/components/PageHeader.vue'
import depopsApi from '@/api/depops'
import type { NodeItem } from '@/api/depops'

const router = useRouter()
const route = useRoute()
const message = useMessage()

const nodeId = parseInt(route.params.id as string)
const loading = ref(false)
const testing = ref(false)
const loadingSystemInfo = ref(false)
const executing = ref(false)

const nodeData = ref<NodeItem>({} as NodeItem)
const command = ref('')

interface CommandHistoryItem {
  command: string
  output?: string
  error?: string
  timestamp: string
}

const commandHistory = ref<CommandHistoryItem[]>([])

// 获取节点详情
const loadNodeData = async () => {
  try {
    loading.value = true
    const result = await depopsApi.node.find(nodeId)
    nodeData.value = result.data || {} as NodeItem
  } catch (error) {
    console.error('加载节点详情失败:', error)
  } finally {
    loading.value = false
  }
}

// 刷新数据
const refreshData = () => {
  loadNodeData()
}

// 测试连接
const testConnection = async () => {
  try {
    testing.value = true
    await depopsApi.node.testConnection(nodeId)
    message.success('连接测试成功')
    // 刷新节点数据以更新状态
    await loadNodeData()
  } catch (error) {
    console.error('连接测试失败:', error)
  } finally {
    testing.value = false
  }
}



// 执行命令
const executeCommand = async () => {
  if (!command.value.trim()) {
    message.warning('请输入命令')
    return
  }

  try {
    executing.value = true
    const result = await depopsApi.node.executeCommand(nodeId, command.value)
    
    const historyItem: CommandHistoryItem = {
      command: command.value,
      timestamp: new Date().toISOString(),
    }

    if (result.data) {
      if (result.data.error) {
        historyItem.error = result.data.error
        historyItem.output = result.data.output
      } else {
        historyItem.output = result.data.output
      }
    }

    commandHistory.value.unshift(historyItem)
    
    // 限制历史记录数量
    if (commandHistory.value.length > 20) {
      commandHistory.value = commandHistory.value.slice(0, 20)
    }

    command.value = ''
  } catch (error) {
    console.error('命令执行失败:', error)
  } finally {
    executing.value = false
  }
}

// 获取状态类型
import type { TagProps } from 'naive-ui'
const getStatusType = (status?: number): TagProps['type'] => {
  if (status === 1) return 'success'
  if (status === 2) return 'error'
  return 'default'
}

// 获取状态文本
const getStatusText = (status: number) => {
  const textMap = { 0: '未知', 1: '在线', 2: '离线' }
  return textMap[status as keyof typeof textMap] || '未知'
}

// 获取标签列表
const getTags = (): string[] => {
  if (!nodeData.value.tags) return []
  try {
    return JSON.parse(nodeData.value.tags) as string[]
  } catch {
    return []
  }
}

// 格式化时间
const formatTime = (time: string) => {
  return new Date(time).toLocaleString('zh-CN')
}

onMounted(() => {
  loadNodeData()
})
</script>

<style scoped>
.page-body {
  height: calc(100vh - 120px);
  overflow: auto;
}

.command-item {
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background-color: var(--card-color);
}

.command-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
