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
      <n-grid-item :span="6">
        <n-card :bordered="false" size="small">
          <n-spin :show="treeLoading">
            <n-tree :data="treeData" block-line :default-expanded-keys="expandedKeys" :selected-keys="selectedKeys"
              :on-update:selected-keys="handleSelectNode" selectable />
          </n-spin>
        </n-card>
      </n-grid-item>
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
import { reactive, ref, watch, onMounted } from "vue";
import {
  NSpace,
  NInput,
  NButton,
  NIcon,
  NDataTable,
  NGrid,
  NGridItem,
  NCard,
  NTree,
  NSpin,
  TreeOption,
  useMessage,
} from "naive-ui";
import {
  AddOutline as AddIcon,
  RefreshOutline as RefreshIcon,
} from "@vicons/ionicons5";
import { useRoute, useRouter } from "vue-router";
import XPageHeader from "@/components/PageHeader.vue";
import projectDirApi from "@/api/basic/projectdir";
import type { ProjectDirItem } from "@/api/basic/projectdir";
import { useDataTable } from "@/utils/data-table";
import { renderButtons, renderLink, renderTag, renderTime } from "@/utils/render";
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const message = useMessage();

// 新建目录处理函数
const newHandler = () => {
  // 如果已选择了项目目录，则将目录ID传递给新建页面
  if (selectedKeys.value.length > 0) {
    router.push({
      name: 'projectdir_new',
      query: { parent_id: selectedKeys.value[0] }
    });
  } else {
    router.push({ name: 'projectdir_new' });
  }
};

// 查询参数
const args = reactive({
  name: "",
  parent_id: undefined as number | undefined
});

// 表格列定义
const columns = [
  {
    title: t('fields.id'),
    key: "id",
    render: (row: ProjectDirItem) => renderLink({ name: 'projectdir_detail', params: { id: row.id } }, row.id + ""),
  },
  {
    title: t('fields.name'),
    key: "name",
  },
  {
    title: t('fields.parent'),
    key: "parent_id",
  },
  {
    title: t('fields.status'),
    key: "is_disable",
    render: (row: ProjectDirItem) => renderTag(
      row.is_disable ? t('enums.blocked') : t('enums.normal'),
      row.is_disable ? "warning" : "success"
    ),
  },
  {
    title: t('fields.remark'),
    key: "remark",
  },
  {
    title: t('fields.actions'),
    key: "actions",
    render(row: ProjectDirItem, index: number) {
      return renderButtons([
        row.is_disable ?
          { type: 'success', text: t('buttons.enable'), action: () => enable(row) } :
          { type: 'warning', text: t('buttons.block'), action: () => disable(row), prompt: t('prompts.block') },
        { type: 'warning', text: t('buttons.edit'), action: () => router.push({ name: 'projectdir_edit', params: { id: row.id } }) },
        { type: 'error', text: t('buttons.delete'), action: () => remove(row, index), prompt: t('prompts.delete') },
      ])
    },
  },
];

// 使用数据表格钩子
const { state, pagination, fetchData, changePageSize } = useDataTable(projectDirApi.search, () => {
  return { ...args, filters: route.query.filters }
});

// 目录树相关状态
const treeData = ref<TreeOption[]>([]);
const treeLoading = ref(false);
const expandedKeys = ref<string[]>([]);
const selectedKeys = ref<string[]>([]);

// 目录树加载
const loadProjectDirTree = async () => {
  treeLoading.value = true;
  try {
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
    args.parent_id = parseInt(keys[0]);
  } else {
    args.parent_id = undefined;
  }
  fetchData(1); // 重新加载第一页的数据
};

// 启用目录
async function enable(dir: ProjectDirItem) {
  await projectDirApi.enable(dir.id);
  fetchData();
  loadProjectDirTree(); // 刷新目录树
}

// 禁用目录
async function disable(dir: ProjectDirItem) {
  await projectDirApi.disable(dir.id);
  fetchData();
  loadProjectDirTree(); // 刷新目录树
}

// 删除目录
async function remove(dir: ProjectDirItem, index: number) {
  await projectDirApi.delete(dir.id);
  state.data.splice(index, 1);
  loadProjectDirTree(); // 刷新目录树
}

// 监听路由查询参数变化
watch(() => route.query.filter, (newValue: any, oldValue: any) => {
  fetchData();
});

// 组件挂载时加载项目目录树
onMounted(() => {
  loadProjectDirTree();
});
</script>

<style scoped>
.page-body {
  height: calc(100vh - 120px);
  overflow: auto;
}
</style>