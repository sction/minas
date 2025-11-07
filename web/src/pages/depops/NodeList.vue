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
      <n-button secondary size="small" @click="fetchData(1)">
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
    <n-space style="margin-bottom: 12px;" :size="12">
      <n-input size="small" v-model:value="searchKeyword" placeholder="搜索节点名称、IP地址或主机名" clearable />
      <n-select size="small" v-model:value="selectedStatus" placeholder="节点状态" clearable style="width: 120px;"
        :options="statusOptions" />
      <n-select size="small" v-model:value="selectedTags" placeholder="标签" clearable multiple style="width: 200px;"
        :options="tagOptions" />
      <n-button size="small" type="primary" @click="() => fetchData(1)">{{ t('buttons.search') }}</n-button>
    </n-space>

    <n-data-table remote :row-key="(row: any) => row.id" size="small" :columns="columns" :data="state.data"
      :pagination="pagination" :loading="state.loading" @update:page="fetchData" @update-page-size="changePageSize"
      :row-props="rowProps" scroll-x="max-content" />
  </n-space>

  <!-- 新建/编辑节点模态框 -->
  <n-modal v-model:show="showEditModal" preset="dialog" title="节点配置" style="width: 700px;">
    <n-form ref="formRef" :model="editForm" :rules="editRules" label-placement="left" label-width="120px">
      <n-grid :cols="2" :x-gap="12">
        <n-grid-item>
          <n-form-item label="节点名称" path="name">
            <n-input v-model:value="editForm.name" placeholder="请输入节点名称" />
          </n-form-item>
        </n-grid-item>
        <n-grid-item>
          <n-form-item label="IP地址" path="ip_address">
            <n-input v-model:value="editForm.ip_address" placeholder="请输入IP地址" />
          </n-form-item>
        </n-grid-item>
        <n-grid-item>
          <n-form-item label="SSH端口" path="ssh_port">
            <n-input-number v-model:value="editForm.ssh_port" placeholder="SSH端口" :min="1" :max="65535"
              style="width: 100%;" />
          </n-form-item>
        </n-grid-item>
        <n-grid-item>
          <n-form-item label="用户名" path="username">
            <n-input v-model:value="editForm.username" placeholder="SSH用户名" />
          </n-form-item>
        </n-grid-item>
      </n-grid>

      <n-form-item label="认证方式">
        <n-radio-group v-model:value="authMethod">
          <n-radio value="password">密码认证</n-radio>
          <n-radio value="key">私钥认证</n-radio>
        </n-radio-group>
      </n-form-item>

      <n-form-item v-if="authMethod === 'password'" label="密码" path="password">
        <n-input v-model:value="editForm.password" type="password" placeholder="SSH密码" />
      </n-form-item>

      <div v-if="authMethod === 'key'">
        <n-form-item label="私钥" path="private_key">
          <n-input v-model:value="editForm.private_key" type="textarea" placeholder="SSH私钥内容" :rows="4" />
        </n-form-item>
        <n-form-item label="私钥密码">
          <n-input v-model:value="editForm.private_key_pass" type="password" placeholder="私钥密码（可选）" />
        </n-form-item>
      </div>

      <n-form-item label="标签">
        <n-dynamic-tags round :bordered="false" type="info" v-model:value="nodeTags" @create="handleCreateTag" />
      </n-form-item>

      <n-form-item label="备注">
        <n-input v-model:value="editForm.remark" type="textarea" placeholder="节点备注信息" :rows="2" />
      </n-form-item>
    </n-form>

    <template #action>
      <n-space>
        <n-button @click="showEditModal = false">取消</n-button>
        <n-button type="info" @click="testFormConnection" :loading="testing">测试连接</n-button>
        <n-button type="primary" @click="saveNode" :loading="saving">保存</n-button>
      </n-space>
    </template>
  </n-modal>

</template>

<script setup lang="ts">
import { reactive, ref, watch, onMounted } from "vue";
import {
  NSpace,
  NInput,
  NButton,
  NIcon,
  NDataTable,
  NModal,
  NForm,
  NFormItem,
  NGrid,
  NGridItem,
  NInputNumber,
  NRadio,
  NRadioGroup,
  NSelect,
  NDynamicTags,
  useMessage,
  DataTableColumns,
  FormRules,
} from "naive-ui";
import {
  AddOutline as AddIcon,
  RefreshOutline as RefreshIcon,
} from "@vicons/ionicons5";
import { useRoute, useRouter } from "vue-router";
import XPageHeader from "@/components/PageHeader.vue";
import depopsApi from "@/api/depops";
import type { NodeItem } from "@/api/depops";
import { useDataTable } from "@/utils/data-table";
import { renderButtons, renderLink, renderTag, renderTime } from "@/utils/render";
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const message = useMessage();

// 搜索条件
const searchKeyword = ref("");
const selectedStatus = ref<number | null>(null);
const selectedTags = ref<string[]>([]);

// 状态选项
const statusOptions = [
  { label: "未知", value: 0 },
  { label: "在线", value: 1 },
  { label: "离线", value: 2 },
];

// 标签选项
const tagOptions = ref<{ label: string; value: string }[]>([]);

// 表格列定义
const columns: DataTableColumns<NodeItem> = [
  {
    title: t('fields.id'),
    key: "id",
    width: 80,
    render: (row: NodeItem) => renderLink({ name: 'depops_node_detail', params: { id: row.id } }, row.id + ""),
  },
  {
    title: "节点名称",
    key: "name",
    width: 120,
  },
  {
    title: "IP地址",
    key: "ip_address",
    width: 120,
  },
  {
    title: "SSH端口",
    key: "ssh_port",
    width: 80,
  },
  {
    title: "操作系统",
    key: "os_info",
    width: 150,
    render: (row: NodeItem) => row.os_type ? `${row.os_type} ${row.os_version}` : "未知",
  },
  {
    title: "主机名",
    key: "hostname",
    width: 120,
  },
  {
    title: "状态",
    key: "status",
    width: 80,
    render: (row: NodeItem) => {
      const statusMap = { 0: { text: "未知", type: "default" }, 1: { text: "在线", type: "success" }, 2: { text: "离线", type: "error" } };
      const status = statusMap[row.status as keyof typeof statusMap];
      return renderTag(status.text, status.type as any);
    },
  },
  {
    title: "最后连接",
    key: "last_connected",
    width: 150,
    render: (row: NodeItem) => row.last_connected ? renderTime(row.last_connected) : "从未连接",
  },
  {
    title: t('fields.status'),
    key: "is_disable",
    width: 80,
    render: (row: NodeItem) => renderTag(
      row.is_disable ? t('enums.blocked') : t('enums.normal'),
      row.is_disable ? "warning" : "success"
    ),
  },
  {
    title: t('fields.actions'),
    key: "actions",
    width: 240,
    fixed: 'right',
    render(row: NodeItem, index: number) {
      return renderButtons([
        { type: 'info', text: '连接测试', action: () => testNodeConnection(row.id) },
        { type: 'warning', text: t('buttons.edit'), action: () => editHandler(row) },
        row.is_disable ?
          { type: 'success', text: t('buttons.enable'), action: () => enable(row) } :
          { type: 'warning', text: t('buttons.block'), action: () => disable(row), prompt: t('prompts.block') },
        { type: 'error', text: t('buttons.delete'), action: () => remove(row, index), prompt: t('prompts.delete') },
      ])
    },
  },
];

// 使用数据表格钩子
const { state, pagination, fetchData, changePageSize } = useDataTable(depopsApi.node.search, () => {
  const params: any = {};
  if (searchKeyword.value) {
    params.search = searchKeyword.value;
  }
  if (selectedStatus.value !== null) {
    params.status = selectedStatus.value;
  }
  if (selectedTags.value.length > 0) {
    params.tags = selectedTags.value.join(',');
  }
  return { ...params, filters: route.query.filters };
});

// 行属性
const rowProps = (row: NodeItem) => {
  return {
    style: 'cursor: pointer;'
  }
};

// 新建处理函数
const newHandler = () => {
  editForm.value = {
    name: "",
    ip_address: "",
    ssh_port: 22,
    username: "root",
    password: "",
    private_key: "",
    private_key_pass: "",
    tags: "",
    remark: "",
  };
  authMethod.value = "password";
  nodeTags.value = [];
  showEditModal.value = true;
};

// 编辑模态框
const showEditModal = ref(false);
const editForm = ref<Partial<NodeItem>>({});
const authMethod = ref<"password" | "key">("password");
const nodeTags = ref<string[]>([]);
const saving = ref(false);
const testing = ref(false);

// 表单规则
const editRules: FormRules = {
  name: { required: true, message: '请输入节点名称', trigger: 'blur' },
  ip_address: { required: true, message: '请输入IP地址', trigger: 'blur' },
  ssh_port: { required: true, type: 'number', message: '请输入SSH端口', trigger: 'blur' },
  username: { required: true, message: '请输入用户名', trigger: 'blur' },
  password: [{
    validator: (rule: any, value: string) => {
      if (authMethod.value === 'password' && !value) {
        return new Error('请输入密码');
      }
      return true;
    },
    trigger: 'blur'
  }],
  private_key: [{
    validator: (rule: any, value: string) => {
      if (authMethod.value === 'key' && !value) {
        return new Error('请输入私钥');
      }
      return true;
    },
    trigger: 'blur'
  }],
};

// 编辑处理函数
const editHandler = (row: NodeItem) => {
  editForm.value = { ...row };
  authMethod.value = row.private_key ? "key" : "password";
  if (row.tags) {
    try {
      const tags = JSON.parse(row.tags) as string[];
      nodeTags.value = tags;
    } catch {
      nodeTags.value = [];
    }
  } else {
    nodeTags.value = [];
  }
  showEditModal.value = true;
};

// 创建标签
const handleCreateTag = (label: string) => {
  return label;
};

// 保存节点
const saveNode = async () => {
  try {
    saving.value = true;

    // 检查是否是新建节点或连接信息有变化
    const isNewNode = !editForm.value.id;

    // 构建标签JSON
    const tags = nodeTags.value;
    editForm.value.tags = JSON.stringify(tags);

    // 根据认证方式清空对应字段
    if (authMethod.value === 'password') {
      editForm.value.private_key = "";
      editForm.value.private_key_pass = "";
    } else {
      editForm.value.password = "";
    }

    await depopsApi.node.save(editForm.value);

    showEditModal.value = false;
    fetchData(1);
  } catch (error) {
    console.error('保存节点失败:', error);
    message.error('保存失败');
  } finally {
    saving.value = false;
  }
};


// 测试表单数据连接（无需保存）
const testFormConnection = async () => {
  if (testing.value) return;

  try {
    testing.value = true;

    // 验证必填字段
    if (!editForm.value.name || !editForm.value.ip_address || !editForm.value.username) {
      message.warning('请填写节点名称、IP地址和用户名');
      return;
    }

    // 验证认证信息
    if (authMethod.value === 'password' && !editForm.value.password) {
      message.warning('请输入SSH密码');
      return;
    }
    if (authMethod.value === 'key' && !editForm.value.private_key) {
      message.warning('请输入SSH私钥');
      return;
    }

    // 构建测试数据
    const testData = { ...editForm.value };

    // 根据认证方式清空对应字段
    if (authMethod.value === 'password') {
      testData.private_key = "";
      testData.private_key_pass = "";
    } else {
      testData.password = "";
    }

    await depopsApi.node.testConnectionData(testData);
  } catch (error) {
    console.error('连接测试失败:', error);
  } finally {
    testing.value = false;
  }
};

// 测试节点连接
const testNodeConnection = async (id: number) => {
  try {
    await depopsApi.node.testConnection(id);
  } catch (error) {
    console.error('连接测试失败:', error);
  }
  fetchData(); // 刷新列表以更新状态
};


// 启用节点
async function enable(node: NodeItem) {
  await depopsApi.node.enable(node.id);
  message.success('启用成功');
  fetchData(1);
}

// 禁用节点
async function disable(node: NodeItem) {
  await depopsApi.node.disable(node.id);
  message.success('禁用成功');
  fetchData(1);
}

// 删除节点
async function remove(node: NodeItem, index: number) {
  await depopsApi.node.delete(node.id);
  message.success('删除成功');
  fetchData(1);
}

// 加载标签选项
const loadTags = async () => {
  try {
    const result = await depopsApi.node.getTags();
    tagOptions.value = (result.data || []).map((tag: string) => ({
      label: tag,
      value: tag,
    }));
  } catch (error) {
    console.error('加载标签失败:', error);
  }
};

// 监听搜索条件变化
watch([searchKeyword, selectedStatus, selectedTags], () => {
  fetchData(1);
}, { deep: true });

// 组件挂载时加载数据
onMounted(() => {
  loadTags();
});
</script>

<style scoped>
.page-body {
  height: calc(100vh - 120px);
  overflow: auto;
}
</style>
