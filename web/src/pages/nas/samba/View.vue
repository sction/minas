<template>
  <x-page-header :subtitle="model.sambaShare.name">
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
      <x-description-item label="ID">{{ model.sambaShare.id }}</x-description-item>
      <x-description-item label="共享名称">{{ model.sambaShare.name }}</x-description-item>
      <x-description-item label="共享路径">{{ model.sambaShare.path }}</x-description-item>
      <x-description-item label="共享描述" v-if="model.sambaShare.comment">
        {{ model.sambaShare.comment }}
      </x-description-item>
      <x-description-item label="是否可浏览">
        <n-tag round size="small" :type="model.sambaShare.browseable ? 'success' : 'warning'">
          {{ model.sambaShare.browseable ? '可浏览' : '不可浏览' }}
        </n-tag>
      </x-description-item>
      <x-description-item label="读写权限">
        <n-tag round size="small" :type="model.sambaShare.readonly ? 'warning' : 'success'">
          {{ model.sambaShare.readonly ? '只读' : '读写' }}
        </n-tag>
      </x-description-item>
      <x-description-item label="来宾访问">
        <n-tag round size="small" :type="model.sambaShare.guestOk ? 'success' : 'error'">
          {{ model.sambaShare.guestOk ? '允许' : '禁止' }}
        </n-tag>
      </x-description-item>
      <x-description-item label="有效用户" v-if="model.sambaShare.validUsers">
        {{ model.sambaShare.validUsers }}
      </x-description-item>
      <x-description-item label="写权限用户" v-if="model.sambaShare.writeList">
        {{ model.sambaShare.writeList }}
      </x-description-item>
      <x-description-item label="读权限用户" v-if="model.sambaShare.readList">
        {{ model.sambaShare.readList }}
      </x-description-item>
      <x-description-item label="创建文件权限" v-if="model.sambaShare.createMask">
        {{ model.sambaShare.createMask }}
      </x-description-item>
      <x-description-item label="创建目录权限" v-if="model.sambaShare.directoryMask">
        {{ model.sambaShare.directoryMask }}
      </x-description-item>
      <x-description-item label="强制用户" v-if="model.sambaShare.forceUser">
        {{ model.sambaShare.forceUser }}
      </x-description-item>
      <x-description-item label="强制组" v-if="model.sambaShare.forceGroup">
        {{ model.sambaShare.forceGroup }}
      </x-description-item>
      <x-description-item label="可用空间限制" v-if="model.sambaShare.availableSpace">
        {{ model.sambaShare.availableSpace }}
      </x-description-item>
      <x-description-item label="其他选项" v-if="model.sambaShare.options">
        <n-code :code="model.sambaShare.options" language="text" />
      </x-description-item>
      <x-description-item label="备注" v-if="model.sambaShare.remark">
        {{ model.sambaShare.remark }}
      </x-description-item>
      <x-description-item label="状态">
        <n-tag round size="small" :type="model.sambaShare.is_disable ? 'error' : 'success'">
          {{ model.sambaShare.is_disable ? '禁用' : '启用' }}
        </n-tag>
      </x-description-item>
      <x-description-item label="SMB配置段">
        <n-code :code="getSmbConfigSection()" language="text" />
        <n-button strong secondary size="small" circle type="primary" @click="copyText(getSmbConfigSection())">
          <template #icon>
            <n-icon>
              <copy-icon />
            </n-icon>
          </template>
        </n-button>
      </x-description-item>
      <x-description-item label="创建时间">
        {{ model.sambaShare.created_at }}
      </x-description-item>
      <x-description-item label="更新时间">
        {{ model.sambaShare.updated_at }}
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
import sambaApi from "@/api/nas/samba";
import type { SambaShare } from "@/api/nas/samba";
import { copyText } from "@/utils";
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()
const route = useRoute();
const router = useRouter();

const model = reactive({
  sambaShare: {} as SambaShare,
});

const listHandler = () => {
  router.push({ name: 'samba_shares' })
}

const editHandler = () => {
  router.push({ name: 'samba_share_edit', params: { id: model.sambaShare.id } })
}

// 生成SMB配置段
const getSmbConfigSection = () => {
  if (!model.sambaShare.name) return ''
  
  let config = `[${model.sambaShare.name}]\n`
  config += `path = ${model.sambaShare.path}\n`
  
  if (model.sambaShare.comment) {
    config += `comment = ${model.sambaShare.comment}\n`
  }
  
  config += `browseable = ${model.sambaShare.browseable ? 'yes' : 'no'}\n`
  config += `read only = ${model.sambaShare.readonly ? 'yes' : 'no'}\n`
  config += `guest ok = ${model.sambaShare.guestOk ? 'yes' : 'no'}\n`
  
  if (model.sambaShare.validUsers) {
    config += `valid users = ${model.sambaShare.validUsers}\n`
  }
  
  if (model.sambaShare.writeList) {
    config += `write list = ${model.sambaShare.writeList}\n`
  }
  
  if (model.sambaShare.readList) {
    config += `read list = ${model.sambaShare.readList}\n`
  }
  
  if (model.sambaShare.createMask) {
    config += `create mask = ${model.sambaShare.createMask}\n`
  }
  
  if (model.sambaShare.directoryMask) {
    config += `directory mask = ${model.sambaShare.directoryMask}\n`
  }
  
  if (model.sambaShare.forceUser) {
    config += `force user = ${model.sambaShare.forceUser}\n`
  }
  
  if (model.sambaShare.forceGroup) {
    config += `force group = ${model.sambaShare.forceGroup}\n`
  }
  
  if (model.sambaShare.availableSpace) {
    config += `dfree command = /usr/bin/df\n`
  }
  
  // 添加其他选项
  if (model.sambaShare.options) {
    const lines = model.sambaShare.options.split('\n')
    for (const line of lines) {
      const trimmedLine = line.trim()
      if (trimmedLine) {
        config += `${trimmedLine}\n`
      }
    }
  }
  
  return config
}

async function fetchData() {
  try {
    const result = await sambaApi.load(parseInt(route.params.id as string));
    model.sambaShare = result.data as any;
  } catch (error) {
    console.error('加载Samba共享详情失败:', error);
  }
}

watch(() => route.params.id, fetchData)

onMounted(fetchData);
</script>