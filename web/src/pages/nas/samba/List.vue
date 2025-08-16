<template>
    <x-page-header>
    </x-page-header>
    <n-space class="page-body" vertical :size="12">
        <!-- 服务状态信息 -->
        <n-card size="small">
            <template #header>
                <n-space align="center">
                    <n-icon size="20">
                        <server-icon />
                    </n-icon>
                    Samba 服务状态
                </n-space>
            </template>
            <template #header-extra>
                <n-button size="small" @click="fetchServiceStatus" secondary>
                    <template #icon>
                        <n-icon>
                            <refresh-icon />
                        </n-icon>
                    </template>
                    刷新服务状态
                </n-button>
            </template>
            <n-space :size="16" align="center">
                <n-tag round :type="serviceStatus.installed ? 'success' : 'error'">
                    {{ serviceStatus.installed ? '已安装' : '未安装' }}
                </n-tag>
                <n-tag round :type="serviceStatus.running ? 'success' : 'warning'">
                    {{ serviceStatus.running ? '运行中' : '已停止' }}
                </n-tag>
                <n-space>
                    <n-button size="small" type="success" @click="startService" :disabled="serviceStatus.running">
                        启动服务
                    </n-button>
                    <n-button size="small" type="warning" @click="stopService" :disabled="!serviceStatus.running">
                        停止服务
                    </n-button>
                    <n-button size="small" type="primary" @click="restartService" :disabled="!serviceStatus.running">
                        重启服务
                    </n-button>
                    <n-space align="center">
                        <span>开机自启:</span>
                        <n-switch v-model:value="serviceStatus.enabled" @update:value="handleAutoStartToggle"
                            :disabled="switchLoading" />
                    </n-space>
                </n-space>
            </n-space>
        </n-card>

        <!-- 同步状态信息 -->
        <n-card size="small">
            <template #header>
                <n-space align="center">
                    <n-icon size="20">
                        <sync-icon />
                    </n-icon>
                    配置同步状态
                </n-space>
            </template>
            <template #header-extra>
                <n-space>
                    <n-button size="small" @click="viewSmbConfig" secondary>
                        <template #icon>
                            <n-icon>
                                <eye-icon />
                            </n-icon>
                        </template>
                        查看smb.conf配置
                    </n-button>
                    <n-button size="small" @click="backupSmbConfig" secondary>
                        <template #icon>
                            <n-icon>
                                <save-icon />
                            </n-icon>
                        </template>
                        备份smb.conf配置
                    </n-button>
                    <n-button size="small" @click="testConfig" secondary>
                        <template #icon>
                            <n-icon>
                                <checkmark-icon />
                            </n-icon>
                        </template>
                        测试配置
                    </n-button>
                    <n-button size="small" @click="fetchSyncStatus" secondary :loading="syncLoading">
                        <template #icon>
                            <n-icon>
                                <refresh-icon />
                            </n-icon>
                        </template>
                        刷新状态
                    </n-button>
                </n-space>
            </template>
            <n-spin :show="syncLoading">
                <n-grid cols="2 640:4" :x-gap="12" :y-gap="8">
                    <n-gi>
                        <n-space vertical align="center">
                            <n-tag round type="info">服务器配置</n-tag>
                            <span style="font-size: 18px; font-weight: bold;">{{ syncStatus.server_shares_count || 0 }}</span>
                        </n-space>
                    </n-gi>
                    <n-gi>
                        <n-space vertical align="center">
                            <n-tag round type="primary">数据库配置</n-tag>
                            <span style="font-size: 18px; font-weight: bold;">{{ syncStatus.db_shares_count || 0 }}</span>
                        </n-space>
                    </n-gi>
                    <n-gi>
                        <n-space vertical align="center">
                            <n-tag round type="warning">需要拉取</n-tag>
                            <span style="font-size: 18px; font-weight: bold;">{{ syncStatus.server_only_count || 0 }}</span>
                        </n-space>
                    </n-gi>
                    <n-gi>
                        <n-space vertical align="center">
                            <n-tag round type="error">需要推送</n-tag>
                            <span style="font-size: 18px; font-weight: bold;">{{ syncStatus.db_only_count || 0 }}</span>
                        </n-space>
                    </n-gi>
                </n-grid>
                <n-space justify="center" class="mt15">
                    <n-button type="warning" size="small" @click="syncFromServer"
                        :disabled="!syncStatus.need_sync_from_server" :loading="syncFromServerLoading">
                        <template #icon>
                            <n-icon>
                                <download-icon />
                            </n-icon>
                        </template>
                        从服务器拉取配置
                    </n-button>
                    <n-button type="primary" size="small" @click="syncToServer"
                        :disabled="!syncStatus.need_sync_to_server" :loading="syncToServerLoading">
                        <template #icon>
                            <n-icon>
                                <upload-icon />
                            </n-icon>
                        </template>
                        推送配置到服务器
                    </n-button>
                    <n-button type="primary" size="small" @click="reloadConfig">
                        <template #icon>
                            <n-icon>
                                <refresh-icon />
                            </n-icon>
                        </template>
                        重新加载配置
                    </n-button>
                </n-space>
                <n-alert v-if="!syncStatus.need_sync_from_server && !syncStatus.need_sync_to_server" type="success"
                    class="mt10">
                    配置已同步，无需操作
                </n-alert>
            </n-spin>
        </n-card>
        
        <n-card size="small">
            <template #header>
                <n-space align="center">
                    <n-icon size="20">
                        <server-icon />
                    </n-icon>
                    服务列表
                </n-space>
            </template>
            <template #header-extra>
                <n-button secondary size="small" @click="newHandler">
                    <template #icon>
                        <n-icon>
                            <add-icon />
                        </n-icon>
                    </template>
                    {{ t('buttons.new') }}
                </n-button>
                <n-button secondary class="ml10" size="small" @click="fetchData()">
                    <template #icon>
                        <n-icon>
                            <refresh-icon />
                        </n-icon>
                    </template>
                    {{ t('buttons.refresh') }}
                </n-button>
                <n-button secondary class="ml10" size="small" @click="openUserManageModal">
                    <template #icon>
                        <n-icon>
                            <person-icon />
                        </n-icon>
                    </template>
                    用户管理
                </n-button>
            </template>

        <!-- 搜索区域 -->
        <n-space :size="12">
            <n-input size="small" v-model:value="args.name" placeholder="共享名称" clearable />
            <n-input size="small" v-model:value="args.path" placeholder="共享路径" clearable />
            <n-button size="small" type="primary" @click="() => fetchData()">{{ t('buttons.search') }}</n-button>
        </n-space>

        <!-- 数据表格 -->
        <n-data-table remote :row-key="(row: any) => row.id" size="small" :columns="columns" :data="state.data"
            :pagination="pagination" :loading="state.loading" @update:page="fetchData"
            @update-page-size="changePageSize" scroll-x="max-content" />
        </n-card>
    </n-space>

    <!-- Samba用户管理模态框 -->
    <n-modal v-model:show="showUserManageModal" preset="card" title="用户管理" style="width: 900px;">
        <template #header-extra>
            <n-space>
                <n-button size="small" @click="showUserAddDialog = true" type="primary">
                    <template #icon>
                        <n-icon>
                            <add-icon />
                        </n-icon>
                    </template>
                    添加用户
                </n-button>
                <n-button size="small" @click="loadUserList" secondary>
                    <template #icon>
                        <n-icon>
                            <refresh-icon />
                        </n-icon>
                    </template>
                    刷新列表
                </n-button>
            </n-space>
        </template>
        
        <n-data-table 
            size="small" 
            :columns="userTableColumns" 
            :data="userList" 
            :loading="userListLoading"
            :pagination="{ pageSize: 10 }" 
        />
        
        <!-- 添加用户对话框 -->
        <n-modal v-model:show="showUserAddDialog" preset="dialog" title="添加用户">
            <n-form ref="userFormRef" :model="userFormData" :rules="userFormRules">
                <n-form-item label="用户名" path="username">
                    <n-input v-model:value="userFormData.username" placeholder="请输入用户名" />
                </n-form-item>
                <n-form-item label="密码" path="password">
                    <n-input v-model:value="userFormData.password" type="password" placeholder="请输入密码" />
                </n-form-item>
            </n-form>
            <template #action>
                <n-space>
                    <n-button @click="showUserAddDialog = false">取消</n-button>
                    <n-button type="primary" @click="handleAddUser" :loading="userActionLoading">添加</n-button>
                </n-space>
            </template>
        </n-modal>
        
        <!-- 修改密码对话框 -->
        <n-modal v-model:show="showPasswordDialog" preset="dialog" title="修改用户密码">
            <n-form ref="passwordFormRef" :model="passwordFormData" :rules="passwordFormRules">
                <n-form-item label="用户名">
                    <n-input :value="passwordFormData.username" readonly />
                </n-form-item>
                <n-form-item label="新密码" path="password">
                    <n-input v-model:value="passwordFormData.password" type="password" placeholder="请输入新密码" />
                </n-form-item>
            </n-form>
            <template #action>
                <n-space>
                    <n-button @click="showPasswordDialog = false">取消</n-button>
                    <n-button type="primary" @click="handleChangePassword" :loading="userActionLoading">修改</n-button>
                </n-space>
            </template>
        </n-modal>
    </n-modal>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch, h } from "vue";
import {
    NButton,
    NSpace,
    NInput,
    NDataTable,
    NSwitch,
    NTag,
    NCard,
    NGrid,
    NGi,
    NSpin,
    NAlert,
    NIcon,
    NModal,
    NForm,
    NFormItem,
    useMessage,
    useDialog,
} from "naive-ui";

import {
    AddOutline as AddIcon,
    RefreshOutline as RefreshIcon,
    DownloadOutline as DownloadIcon,
    CloudUploadOutline as UploadIcon,
    ServerOutline as ServerIcon,
    SyncOutline as SyncIcon,
    EyeOutline as EyeIcon,
    SaveOutline as SaveIcon,
    CheckmarkOutline as CheckmarkIcon,
    PersonOutline as PersonIcon,
} from "@vicons/ionicons5";
import { useRoute, useRouter } from "vue-router";
import XPageHeader from "@/components/PageHeader.vue";
import { useDataTable } from "@/utils/data-table";
import { renderButtons, renderLink, renderTag, renderTime } from "@/utils/render";
import sambaApi from "@/api/nas/samba";
import type { SambaShare, SambaServiceStatus } from "@/api/nas/samba";
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const route = useRoute();
const router = useRouter();
const message = useMessage();
const dialog = useDialog();

const newHandler = () => {
    router.push({ name: 'samba_new' })
}

const args = reactive({
    name: "",
    path: "",
});

const serviceStatus = ref<SambaServiceStatus>({
    installed: false,
    running: false,
    enabled: false,
});

const switchLoading = ref(false);
const syncLoading = ref(false);
const syncFromServerLoading = ref(false);
const syncToServerLoading = ref(false);
const showUserManageModal = ref(false);

// 用户管理相关
const userList = ref<any[]>([]);
const userListLoading = ref(false);
const showUserAddDialog = ref(false);
const showPasswordDialog = ref(false);
const userActionLoading = ref(false);
const userFormRef = ref();
const passwordFormRef = ref();

const userFormData = reactive({
    username: '',
    password: ''
});

const passwordFormData = reactive({
    username: '',
    password: ''
});

const userFormRules = {
    username: [
        { required: true, message: '请输入用户名', trigger: 'blur' },
        { min: 3, max: 20, message: '用户名长度应为3-20个字符', trigger: 'blur' }
    ],
    password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码至少6个字符', trigger: 'blur' }
    ]
};

const passwordFormRules = {
    password: [
        { required: true, message: '请输入新密码', trigger: 'blur' },
        { min: 6, message: '密码至少6个字符', trigger: 'blur' }
    ]
};

const userTableColumns = [
    {
        title: '用户名',
        key: 'username',
    },
    {
        title: '操作',
        key: 'actions',
        render(row: any) {
            return renderButtons([
                { type: 'warning', text: '修改密码', action: () => openPasswordDialog(row.username) },
                { type: 'error', text: '删除', action: () => handleDeleteUser(row.username), prompt: '确定要删除这个Samba用户吗？' },
            ])
        },
    },
];



const columns = [
    {
        title: 'ID',
        key: "id",
        render: (row: SambaShare) => renderLink({ name: 'samba_detail', params: { id: row.id.toString() } }, row.id.toString()),
    },
    {
        title: '共享名称',
        key: "name",
    },
    {
        title: '共享路径',
        key: "path",
    },
    {
        title: '描述',
        key: "comment",
        render: (row: SambaShare) => {
            const comment = row.comment || '-';
            if (comment.length > 20) {
                return h('span', { title: comment }, comment.substring(0, 17) + '...');
            }
            return comment;
        },
    },
    {
        title: '可浏览',
        key: "browseable",
        render: (row: SambaShare) => renderTag(
            row.browseable ? '是' : '否',
            row.browseable ? "success" : "warning"
        ),
    },
    {
        title: '只读',
        key: "readonly",
        render: (row: SambaShare) => renderTag(
            row.readOnly ? '是' : '否',
            row.readOnly ? "warning" : "success"
        ),
    },
    {
        title: '来宾访问',
        key: "guestOk",
        render: (row: SambaShare) => renderTag(
            row.guestOk ? '允许' : '禁止',
            row.guestOk ? "success" : "error"
        ),
    },
    {
        title: '状态',
        key: "is_disable",
        render: (row: SambaShare) => renderTag(
            row.is_disable === 0 ? '启用' : '禁用',
            row.is_disable === 0 ? "success" : "error"
        ),
    },
    {
        title: '更新时间',
        key: "updated_at",
        render: (row: SambaShare) => renderTime(row.updated_at),
    },
    {
        title: '操作',
        key: "actions",
        render(row: SambaShare, index: number) {
            return renderButtons([
                row.is_disable === 1 ?
                    { type: 'success', text: '启用', action: () => enable(row) } :
                    { type: 'warning', text: '禁用', action: () => disable(row), prompt: '确定要禁用这个Samba共享吗？' },
                { type: 'primary', text: '查看', action: () => router.push({ name: 'samba_detail', params: { id: row.id } }) },
                { type: 'warning', text: '编辑', action: () => router.push({ name: 'samba_edit', params: { id: row.id } }) },
                { type: 'error', text: '删除', action: () => remove(row, index), prompt: '确定要删除这个Samba共享吗？' },
            ])
        },
    },
];

const { state, pagination, fetchData, changePageSize } = useDataTable(sambaApi.search, () => {
    return { ...args, filters: route.query.filters }
})

// 获取服务状态
async function fetchServiceStatus() {
    try {
        const result = await sambaApi.getServiceStatus();
        if (result.code === 200) {
            serviceStatus.value = result.data as any;
        }
    } catch (error) {
        console.error('获取Samba服务状态失败:', error);
    }
}

// 同步状态
const syncStatus = ref({
    server_shares_count: 0,
    db_shares_count: 0,
    server_only_count: 0,
    db_only_count: 0,
    need_sync_from_server: false,
    need_sync_to_server: false,
});

// 获取同步状态
async function fetchSyncStatus() {
    syncLoading.value = true;
    try {
        const result = await sambaApi.getSyncStatus();
        if (result.code === 200 && result.data) {
            const data = result.data as any;
            syncStatus.value = {
                server_shares_count: data.server_shares_count || 0,
                db_shares_count: data.db_shares_count || 0,
                server_only_count: data.server_only_count || 0,
                db_only_count: data.db_only_count || 0,
                need_sync_from_server: data.need_sync_from_server || false,
                need_sync_to_server: data.need_sync_to_server || false,
            };
        }
    } catch (error) {
        console.error('获取同步状态失败:', error);
    } finally {
        syncLoading.value = false;
    }
}

// 用户管理方法
function openUserManageModal() {
    showUserManageModal.value = true;
    loadUserList();
}

async function loadUserList() {
    userListLoading.value = true;
    try {
        const result = await sambaApi.listSambaUsers();
        if (result.code === 200) {
            if (Array.isArray(result.data)) {
                userList.value = result.data;
            } else if (result.data && 'data' in result.data) {
                userList.value = result.data.data || [];
            } else {
                userList.value = [];
            }
        }
    } catch (error) {
        console.error('获取Samba用户列表失败:', error);
        message.error('获取用户列表失败');
    } finally {
        userListLoading.value = false;
    }
}

function openPasswordDialog(username: string) {
    passwordFormData.username = username;
    passwordFormData.password = '';
    showPasswordDialog.value = true;
}

async function handleAddUser() {
    try {
        await userFormRef.value?.validate();
        userActionLoading.value = true;
        
        const result = await sambaApi.addSambaUser(userFormData.username, userFormData.password);
        if (result.code === 200) {
            message.success('用户添加成功');
            showUserAddDialog.value = false;
            userFormData.username = '';
            userFormData.password = '';
            await loadUserList();
        }
    } catch (error) {
        console.error('添加用户失败:', error);
        message.error('添加用户失败');
    } finally {
        userActionLoading.value = false;
    }
}

async function handleChangePassword() {
    try {
        await passwordFormRef.value?.validate();
        userActionLoading.value = true;
        
        const result = await sambaApi.changeSambaUserPassword(passwordFormData.username, passwordFormData.password);
        if (result.code === 200) {
            message.success('密码修改成功');
            showPasswordDialog.value = false;
            passwordFormData.password = '';
        }
    } catch (error) {
        console.error('修改密码失败:', error);
        message.error('修改密码失败');
    } finally {
        userActionLoading.value = false;
    }
}

async function handleDeleteUser(username: string) {
    try {
        await sambaApi.deleteSambaUser(username);
        message.success('用户删除成功');
        await loadUserList();
    } catch (error) {
        console.error('删除用户失败:', error);
        message.error('删除用户失败');
    }
}

// 服务管理
async function startService() {
    dialog.warning({
        title: '确认启动服务',
        content: '确定要启动Samba服务吗？',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            try {
                await sambaApi.startService();
                await fetchServiceStatus();
                message.success('Samba服务启动成功');
            } catch (error) {
                console.error('启动服务失败:', error);
                message.error('启动服务失败');
            }
        }
    });
}

async function stopService() {
    dialog.warning({
        title: '确认停止服务',
        content: '确定要停止Samba服务吗？停止后所有Samba共享将不可用。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            try {
                await sambaApi.stopService();
                await fetchServiceStatus();
                message.success('Samba服务停止成功');
            } catch (error) {
                console.error('停止服务失败:', error);
                message.error('停止服务失败');
            }
        }
    });
}

async function restartService() {
    dialog.warning({
        title: '确认重启服务',
        content: '确定要重启Samba服务吗？重启过程中Samba共享将暂时不可用。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            try {
                await sambaApi.restartService();
                await fetchServiceStatus();
                message.success('Samba服务重启成功');
            } catch (error) {
                console.error('重启服务失败:', error);
                message.error('重启服务失败');
            }
        }
    });
}

// 处理开机自启动开关切换
async function handleAutoStartToggle(value: boolean) {
    const actionText = value ? '启用' : '禁用';
    const description = value
        ? '启用后，系统重启时将自动启动Samba服务。'
        : '禁用后，系统重启时将不会自动启动Samba服务。';

    dialog.warning({
        title: `确认${actionText}开机自启`,
        content: `确定要${actionText}Samba服务开机自启吗？${description}`,
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            switchLoading.value = true;
            try {
                if (value) {
                    await sambaApi.enableAutoStart();
                } else {
                    await sambaApi.disableAutoStart();
                }
                await fetchServiceStatus();
                message.success(value ? 'Samba开机自启已启用' : 'Samba开机自启已禁用');
            } catch (error) {
                console.error('切换自启状态失败:', error);
                serviceStatus.value.enabled = !value;
                message.error('切换自启状态失败');
            } finally {
                switchLoading.value = false;
            }
        },
        onNegativeClick: () => {
            serviceStatus.value.enabled = !value;
        }
    });
}

// 配置管理
async function reloadConfig() {
    await sambaApi.reloadConfig();
}

// 启用/禁用Samba共享
async function enable(share: SambaShare) {
    try {
        await sambaApi.enable(share.id);
        await fetchData();
    } catch (error) {
        console.error('启用失败:', error);
        message.error('启用失败');
    }
}

async function disable(share: SambaShare) {
    try {
        await sambaApi.disable(share.id);
        await fetchData();
    } catch (error) {
        console.error('禁用失败:', error);
        message.error('禁用失败');
    }
}

async function remove(share: SambaShare, index: number) {
    try {
        await sambaApi.delete(share.id);
        state.data.splice(index, 1);
        message.success('删除成功');
    } catch (error) {
        console.error('删除失败:', error);
    }
}

// 从服务器同步配置
async function syncFromServer() {
    dialog.warning({
        title: '确认同步',
        content: '确定要从服务器拉取Samba配置到数据库吗？这将添加服务器上存在但数据库中不存在的配置。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            syncFromServerLoading.value = true;
            try {
                const result = await sambaApi.syncFromServer();
                if (result.code === 200) {
                    message.success(result.msg || '同步成功');
                    await fetchData();
                    await fetchSyncStatus();
                }
            } catch (error) {
                console.error('同步失败:', error);
                message.error('同步失败');
            } finally {
                syncFromServerLoading.value = false;
            }
        }
    });
}

// 同步配置到服务器
async function syncToServer() {
    dialog.warning({
        title: '确认同步',
        content: '确定要将数据库配置推送到服务器吗？这将覆盖服务器上的Samba配置文件。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            syncToServerLoading.value = true;
            try {
                const result = await sambaApi.syncToServer();
                if (result.code === 200) {
                    message.success(result.msg || '同步成功');
                    await fetchSyncStatus();
                }
            } catch (error) {
                console.error('同步失败:', error);
                message.error('同步失败');
            } finally {
                syncToServerLoading.value = false;
            }
        }
    });
}

// 查看smb.conf配置
async function viewSmbConfig() {
    try {
        const result = await sambaApi.getSmbConfig();
        if (result.code === 200) {
            const content = (result.data || '配置文件为空') as string;
            dialog.info({
                title: 'smb.conf配置文件内容',
                content: () => h('pre', { 
                    style: { 
                        'white-space': 'pre-wrap', 
                        'font-family': 'monospace',
                        'font-size': '12px',
                        'line-height': '1.4',
                        'max-height': '600px',
                        'overflow-y': 'auto',
                        'background': '#f5f5f5',
                        'padding': '12px',
                        'border-radius': '4px'
                    } 
                }, content),
                positiveText: '关闭',
                style: {
                    width: '800px'
                }
            });
        }
    } catch (error) {
        console.error('获取smb.conf配置失败:', error);
        message.error('获取smb.conf配置失败');
    }
}

// 备份smb.conf配置
async function backupSmbConfig() {
    try {
        await sambaApi.backupSmbConfig();
        message.success('备份成功');
    } catch (error) {
        console.error('备份smb.conf配置失败:', error);
        message.error('备份smb.conf配置失败');
    }
}

// 测试配置
async function testConfig() {
    try {
        const result = await sambaApi.testConfig();
        if (result.code === 200) {
            message.success('配置测试通过');
        }
    } catch (error) {
        console.error('配置测试失败:', error);
        message.error('配置测试失败');
    }
}

watch(() => route.query.filter, (newValue: any, oldValue: any) => {
    fetchData()
})

onMounted(() => {
    fetchServiceStatus();
    fetchSyncStatus();
    fetchData();
});
</script>