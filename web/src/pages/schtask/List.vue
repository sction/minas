<template>
  <x-page-header>
    <template #action>
      <n-button secondary size="small" @click="newHandler">
        <template #icon>
          <n-icon>
            <add-icon />
          </n-icon>
        </template>
        {{ t('buttons.new') }}
      </n-button>
      <n-button secondary size="small" @click="fetchData()">
        <template #icon>
          <n-icon>
            <refresh-icon />
          </n-icon>
        </template>
        {{ t('buttons.refresh') }}
      </n-button>
    </template>
  </x-page-header>
  <n-space class="page-body" vertical :size="12">
    <n-grid :cols="24" :x-gap="12">
      <!-- 左侧项目目录树 -->
      <n-grid-item :span="6">
          <n-spin :show="treeLoading">
            <n-tree
              block-line
              :data="treeData"
              :default-expanded-keys="expandedKeys"
              :selected-keys="selectedKeys"
              :on-update:selected-keys="handleSelectNode"
            />
          </n-spin>
      </n-grid-item>
      <!-- 右侧任务列表 -->
      <n-grid-item :span="18">
        <n-space style="margin-bottom: 12px;" :size="12">
          <n-input size="small" v-model:value="args.name" :placeholder="t('fields.name')" clearable />
          <n-button size="small" type="primary" @click="() => fetchData()">{{ t('buttons.search') }}</n-button>
        </n-space>
        <n-data-table remote :row-key="(row: any) => row.id" size="small" :columns="columns" :data="state.data"
          :pagination="pagination" :loading="state.loading" @update:page="fetchData" @update-page-size="changePageSize"
          scroll-x="max-content" />
      </n-grid-item>
    </n-grid>
  </n-space>
</template>

<script setup lang="ts">
import { reactive, watch, ref, onMounted } from "vue";
import {
  NSpace,
  NInput,
  NButton,
  NIcon,
  NDataTable,
  useMessage,
  NGrid,
  NGridItem,
  NCard,
  NTree,
  NSpin,
  TreeOption
} from "naive-ui";
import {
  AddOutline as AddIcon,
  RefreshOutline as RefreshIcon,
} from "@vicons/ionicons5";
import { useRoute, useRouter } from "vue-router";
import XPageHeader from "@/components/PageHeader.vue";
import schtaskApi from "@/api/sch/task";
import projectDirApi from "@/api/basic/projectdir";
import type { ProjectDirItem } from "@/api/basic/projectdir";
import type { SchTask } from "@/api/sch/task";
import { runStatusMapping, typeMapping } from "@/api/sch/task";
import { useDataTable } from "@/utils/data-table";
import { getPageData, renderButtons, renderLink, renderTag, renderTime, savePageData } from "@/utils/render";
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const route = useRoute();
const router = useRouter();


const args = reactive({
  name: "",
  project_dir_id: getPageData(route.path, 'project_dir_id') as string
});

// 项目目录树相关状态
const treeData = ref<TreeOption[]>([]);
const treeLoading = ref(false);
const expandedKeys = ref<string[]>([]); // 默认展开第一级目录
const selectedKeys = ref<string[]>([args.project_dir_id]);

// 项目目录树加载
const loadProjectDirTree = async () => {
  treeLoading.value = true;
  try {
    // 获取task类型的项目目录树
    const res = await projectDirApi.getTree(); 
    if (res.data && Array.isArray(res.data)) {
      treeData.value = convertToTreeOptions(res.data);
      
      // 如果有数据，默认展开第一级
      if (treeData.value.length > 0) {
        expandedKeys.value = [treeData.value[0].key as string];
      }
    }
  } catch (error) {
    console.error("加载项目目录树失败", error);
  } finally {
    treeLoading.value = false;
  }
};

// 转换项目目录数据为树形选项
const convertToTreeOptions = (dirs: ProjectDirItem[]): TreeOption[] => {
  return dirs.map(dir => ({
    key: dir.id.toString(),
    label: dir.name,
    children: dir.children && dir.children.length > 0 
      ? convertToTreeOptions(dir.children) 
      : undefined
  }));
};

// 处理选择节点事件
const handleSelectNode = (keys: string[]) => {
  selectedKeys.value = keys;
  // 根据选中的目录节点过滤任务列表
  if (keys.length > 0) {
    args.project_dir_id = keys[0];
  } else {
    args.project_dir_id = '';
  }
  savePageData(route.path, 'project_dir_id', args.project_dir_id);
  fetchData(1);
};

// 新建任务处理
const newHandler = () => {
  // 如果已选择了项目目录，则将目录ID传递给新建页面
  if (selectedKeys.value.length > 0) {
    router.push({ 
      name: 'schtask_new',
      query: { project_dir_id: selectedKeys.value[0] }
    });
  } else {
    router.push({ name: 'schtask_new' });
  }
};


const columns = [
  {
    title: t('fields.id'),
    key: "id",
    render: (row: SchTask) => renderLink({ name: 'schtask_detail', params: { id: row.id } }, row.id),
  },
  {
    title: t('fields.name'),
    key: "name",
  },
  {
    title: t('fields.type'),
    key: "type",
    render: (row: SchTask) => typeMapping[row.type + ""],
  },
  {
    title: t('fields.cron'),
    key: "cron",
  },
  {
    title: t('fields.status'),
    key: "is_disable",
    render: (row: SchTask) => renderTag(
      row.is_disable ? t('enums.blocked') : t('enums.normal'),
      row.is_disable ? "warning" : "success"
    ),
  },
  {
    title: t('fields.updated_at'),
    key: "updatedAt",
    render: (row: SchTask) => renderTime(row.updated_at),
  },
  {
    title: t('fields.actions'),
    key: "actions",
    render(row: SchTask, index: number) {
      return renderButtons([
        { type: 'success', text: t('buttons.run_now'), action: () => exec(row, index), prompt: t('prompts.run_now') },
        { type: 'info', text: t('buttons.run_logs'), action: () => router.push({ name: 'schlog_list', params: { id: row.id } }) },
        row.is_disable ?
          { type: 'success', text: t('buttons.enable'), action: () => enable(row), prompt: t('prompts.enable'), } :
          { type: 'warning', text: t('buttons.block'), action: () => disable(row), prompt: t('prompts.block'), },
        { type: 'warning', text: t('buttons.edit'), action: () => router.push({ name: 'schtask_edit', params: { id: row.id } }) },
        { type: 'error', text: t('buttons.delete'), action: () => remove(row, index), prompt: t('prompts.delete') },
      ])
    },
  },
];
const { state, pagination, fetchData, changePageSize } = useDataTable(schtaskApi.search, () => {
  return { ...args, filters: route.query.filters }
})
const message = useMessage()
async function enable(u: SchTask) {
  await schtaskApi.enable(u.id);
  fetchData()
}
async function disable(u: SchTask) {
  await schtaskApi.disable(u.id);
  fetchData()
}
async function remove(u: SchTask, index: number) {
  const data = await schtaskApi.delete(u.id);
  if (data.code === 200) {
    state.data.splice(index, 1)
    message.info(t('texts.action_success'));
  } else {
    message.info(t('texts.action_error'));
  }
}
async function exec(u: SchTask, index: number) {
  const data = await schtaskApi.exec(u.id);
  if (data.code === 200) {
    message.info(t('texts.action_success'));
  } else {
    message.info(t('texts.action_error'));
  }
}

watch(() => route.query.filter, (newValue: any, oldValue: any) => {
  fetchData()
})

// 组件挂载时加载项目目录树
onMounted(() => {
  loadProjectDirTree();
});

</script>