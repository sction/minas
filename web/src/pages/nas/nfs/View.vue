<template>
  <x-page-header :subtitle="model.nfsShare.name">
    <template #action>
      <n-button secondary size="small" @click="listHandler">
        <template #icon>
          <n-icon>
            <back-icon />
          </n-icon>
        </template>
        {{ t('buttons.return') }}
      </n-button>
      <n-button secondary size="small" @click="editHandler">{{ t('buttons.edit') }}</n-button>
    </template>
  </x-page-header>
  <n-space class="page-body" vertical :size="16">
    <x-description cols="1 640:1" label-position="left" label-align="right" :label-width="120">
      <x-description-item label="ID">{{ model.nfsShare.id }}</x-description-item>
      <x-description-item label="共享名称">{{ model.nfsShare.name }}</x-description-item>
      <x-description-item label="共享路径">{{ model.nfsShare.path }}</x-description-item>
      <x-description-item label="客户端IP">{{ model.nfsShare.clientIp || '*' }}</x-description-item>
      <x-description-item label="读写权限">
        <n-tag round size="small" :type="model.nfsShare.permission === 'rw' ? 'success' : 'warning'">
          {{ model.nfsShare.permission === 'rw' ? '读写 (rw)' : '只读 (ro)' }}
        </n-tag>
      </x-description-item>
      <x-description-item label="同步模式">
        <n-tag round size="small" type="info">
          {{ model.nfsShare.sync === 'sync' ? '同步 (sync)' : '异步 (async)' }}
        </n-tag>
      </x-description-item>
      <x-description-item label="Root权限映射">
        {{ getRootSquashLabel(model.nfsShare.rootSquash) }}
      </x-description-item>
      <x-description-item label="子树检查">
        <n-tag round size="small" :type="model.nfsShare.subtreeCheck === 'subtree_check' ? 'success' : 'warning'">
          {{ getSubtreeCheckLabel(model.nfsShare.subtreeCheck) }}
        </n-tag>
      </x-description-item>
      <x-description-item label="其他选项" v-if="model.nfsShare.options">
        {{ model.nfsShare.options }}
      </x-description-item>
      <x-description-item label="备注" v-if="model.nfsShare.remark">
        {{ model.nfsShare.remark }}
      </x-description-item>
      <x-description-item label="状态">
        <n-tag round size="small" :type="model.nfsShare.is_disable ? 'error' : 'success'">
          {{ model.nfsShare.is_disable ? '禁用' : '启用' }}
        </n-tag>
      </x-description-item>
      <x-description-item label="导出配置行">
        <n-code :code="getExportLine()" language="text" />
        <n-button strong secondary size="small" circle type="primary" @click="copyText(getExportLine())">
          <template #icon>
            <n-icon>
              <copy-icon />
            </n-icon>
          </template>
        </n-button>
      </x-description-item>
      <x-description-item label="创建时间">
        {{ model.nfsShare.created_at }}
      </x-description-item>
      <x-description-item label="更新时间">
        {{ model.nfsShare.updated_at }}
      </x-description-item>
    </x-description>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, reactive, watch } from "vue";
import {
  NButton,
  NTag,
  NSpace,
  NIcon,
  NCode,
  useMessage,
} from "naive-ui";
import { useRoute, useRouter } from "vue-router";
import { ArrowBackCircleOutline as BackIcon, CopyOutline as CopyIcon } from "@vicons/ionicons5";
import XPageHeader from "@/components/PageHeader.vue";
import { XDescription, XDescriptionItem } from "@/components/description";
import nfsApi from "@/api/nas/nfs";
import type { NfsShare } from "@/api/nas/nfs";
import { copyText } from "@/utils";
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()
const route = useRoute();
const router = useRouter();

const model = reactive({
  nfsShare: {} as NfsShare,
});

const listHandler = () => {
  router.push({ name: 'nfs_list' })
}

const editHandler = () => {
  router.push({ name: 'nfs_edit', params: { id: model.nfsShare.id } })
}

// 获取Root权限映射的标签
const getRootSquashLabel = (value: string) => {
  const labels: Record<string, string> = {
    '': '使用默认配置',
    'root_squash': 'root_squash (映射root为匿名用户)',
    'no_root_squash': 'no_root_squash (不映射root用户)',
    'all_squash': 'all_squash (映射所有用户为匿名用户)',
  }
  return labels[value] || value
}

// 获取子树检查的标签
const getSubtreeCheckLabel = (value: string) => {
  const labels: Record<string, string> = {
    '': '使用默认配置',
    'subtree_check': '启用子树检查',
    'no_subtree_check': '禁用子树检查',
  }
  return labels[value] || value
}

// 生成导出配置行
const getExportLine = () => {
  if (!model.nfsShare.path) return ''
  
  const clientIp = model.nfsShare.clientIp || '*'
  const options = []
  
  // 添加基本权限
  if (model.nfsShare.permission) {
    options.push(model.nfsShare.permission)
  } else {
    options.push('ro')
  }
  
  // 添加同步模式
  if (model.nfsShare.sync) {
    options.push(model.nfsShare.sync)
  } else {
    options.push('sync')
  }
  
  // 只有在明确设置时才添加root权限映射
  if (model.nfsShare.rootSquash) {
    options.push(model.nfsShare.rootSquash)
  }
  
  // 只有在明确设置时才添加子树检查
  if (model.nfsShare.subtreeCheck) {
    options.push(model.nfsShare.subtreeCheck)
  }
  
  // 添加其他选项
  if (model.nfsShare.options) {
    const customOptions = model.nfsShare.options.split(',')
    for (const opt of customOptions) {
      const trimmedOpt = opt.trim()
      if (trimmedOpt) {
        options.push(trimmedOpt)
      }
    }
  }
  
  return `${model.nfsShare.path} ${clientIp}(${options.join(',')})`
}

async function fetchData() {
  try {
    const result = await nfsApi.load(parseInt(route.params.id as string));
    model.nfsShare = result.data as any;
  } catch (error) {
    console.error('加载NFS共享详情失败:', error);
  }
}

watch(() => route.params.id, fetchData)

onMounted(fetchData);
</script>