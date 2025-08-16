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
                    服务状态
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
                    <n-button size="small" @click="viewExportsConfig" secondary>
                        <template #icon>
                            <n-icon>
                                <eye-icon />
                            </n-icon>
                        </template>
                        查看exports配置
                    </n-button>
                    <n-button size="small" @click="backupExportsConfig" secondary>
                        <template #icon>
                            <n-icon>
                                <save-icon />
                            </n-icon>
                        </template>
                        备份exports配置
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
                            <span style="font-size: 18px; font-weight: bold;">{{ syncStatus.server_shares_count || 0
                            }}</span>
                        </n-space>
                    </n-gi>
                    <n-gi>
                        <n-space vertical align="center">
                            <n-tag round type="primary">数据库配置</n-tag>
                            <span style="font-size: 18px; font-weight: bold;">{{ syncStatus.db_shares_count || 0
                            }}</span>
                        </n-space>
                    </n-gi>
                    <n-gi>
                        <n-space vertical align="center">
                            <n-tag round type="warning">需要拉取</n-tag>
                            <span style="font-size: 18px; font-weight: bold;">{{ syncStatus.server_only_count || 0
                            }}</span>
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
                    <n-button type="primary" size="small" @click="reloadExports">
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
                <n-button secondary class="ml10" size="small" @click="newHandler">
                    <template #icon>
                        <n-icon>
                            <add-icon />
                        </n-icon>
                    </template>
                    {{ t('buttons.new') }}
                </n-button>
                <n-button class="ml10" secondary size="small" @click="fetchData()">
                    <template #icon>
                        <n-icon>
                            <refresh-icon />
                        </n-icon>
                    </template>
                    {{ t('buttons.refresh') }}
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
} from "@vicons/ionicons5";
import { useRoute, useRouter } from "vue-router";
import XPageHeader from "@/components/PageHeader.vue";
import { useDataTable } from "@/utils/data-table";
import { renderButtons, renderLink, renderTag, renderTime } from "@/utils/render";
import nfsApi from "@/api/nas/nfs";
import type { NfsShare, NfsServiceStatus } from "@/api/nas/nfs";
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const route = useRoute();
const router = useRouter();
const message = useMessage();
const dialog = useDialog();

const newHandler = () => {
    router.push({ name: 'nfs_new' })
}

const args = reactive({
    name: "",
    path: "",
});

const serviceStatus = ref<NfsServiceStatus>({
    installed: false,
    running: false,
    enabled: false,
});

const switchLoading = ref(false);
const syncLoading = ref(false);
const syncFromServerLoading = ref(false);
const syncToServerLoading = ref(false);

const columns = [
    {
        title: 'ID',
        key: "id",
        render: (row: NfsShare) => renderLink({ name: 'nfs_detail', params: { id: row.id.toString() } }, row.id.toString()),
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
        title: '客户端IP',
        key: "clientIp",
        render: (row: NfsShare) => {
            const clientIp = row.clientIp || '*';
            // 如果IP太长，截断显示
            if (clientIp.length > 15) {
                return h('span', { title: clientIp }, clientIp.substring(0, 12) + '...');
            }
            return clientIp;
        },
    },
    {
        title: '权限',
        key: "permission",
        render: (row: NfsShare) => renderTag(
            row.permission === 'rw' ? '读写' : '只读',
            row.permission === 'rw' ? "success" : "warning"
        ),
    },
    {
        title: '同步模式',
        key: "sync",
        render: (row: NfsShare) => {
            if (!row.sync || row.sync === '') {
                return renderTag('无', "default");
            }
            return renderTag(
                row.sync === 'sync' ? '同步' : '异步',
                row.sync === 'sync' ? "info" : "warning"
            );
        },
    },
    {
        title: '状态',
        key: "is_disable",
        render: (row: NfsShare) => renderTag(
            row.is_disable === 0 ? '启用' : '禁用',
            row.is_disable === 0 ? "success" : "error"
        ),
    },
    {
        title: '更新时间',
        key: "updated_at",
        render: (row: NfsShare) => renderTime(row.updated_at),
    },
    {
        title: '操作',
        key: "actions",
        render(row: NfsShare, index: number) {
            return renderButtons([
                row.is_disable === 1 ?
                    { type: 'success', text: '启用', action: () => enable(row) } :
                    { type: 'warning', text: '禁用', action: () => disable(row), prompt: '确定要禁用这个NFS共享吗？' },
                { type: 'primary', text: '查看', action: () => router.push({ name: 'nfs_detail', params: { id: row.id } }) },
                { type: 'warning', text: '编辑', action: () => router.push({ name: 'nfs_edit', params: { id: row.id } }) },
                { type: 'error', text: '删除', action: () => remove(row, index), prompt: '确定要删除这个NFS共享吗？' },
            ])
        },
    },
];

const { state, pagination, fetchData, changePageSize } = useDataTable(nfsApi.search, () => {
    return { ...args, filters: route.query.filters }
})

// 获取服务状态
async function fetchServiceStatus() {
    try {
        const result = await nfsApi.getServiceStatus();
        if (result.code === 200) {
            serviceStatus.value = result.data as any;
        }
    } catch (error) {
        console.error('获取NFS服务状态失败:', error);
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
        const result = await nfsApi.getSyncStatus();
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

// 服务管理 - 添加二次确认
async function startService() {
    dialog.warning({
        title: '确认启动服务',
        content: '确定要启动NFS服务吗？',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            try {
                await nfsApi.startService();
                await fetchServiceStatus();
                message.success('NFS服务启动成功');
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
        content: '确定要停止NFS服务吗？停止后所有NFS共享将不可用。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            try {
                await nfsApi.stopService();
                await fetchServiceStatus();
                message.success('NFS服务停止成功');
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
        content: '确定要重启NFS服务吗？重启过程中NFS共享将暂时不可用。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            try {
                await nfsApi.restartService();
                await fetchServiceStatus();
                message.success('NFS服务重启成功');
            } catch (error) {
                console.error('重启服务失败:', error);
                message.error('重启服务失败');
            }
        }
    });
}

// 处理开机自启动开关切换 - 添加二次确认
async function handleAutoStartToggle(value: boolean) {
    const actionText = value ? '启用' : '禁用';
    const description = value
        ? '启用后，系统重启时将自动启动NFS服务。'
        : '禁用后，系统重启时将不会自动启动NFS服务。';

    dialog.warning({
        title: `确认${actionText}开机自启`,
        content: `确定要${actionText}NFS服务开机自启吗？${description}`,
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            switchLoading.value = true;
            try {
                if (value) {
                    await nfsApi.enableAutoStart();
                } else {
                    await nfsApi.disableAutoStart();
                }
                await fetchServiceStatus();
                message.success(value ? 'NFS开机自启已启用' : 'NFS开机自启已禁用');
            } catch (error) {
                console.error('切换自启状态失败:', error);
                // 如果操作失败，恢复开关状态
                serviceStatus.value.enabled = !value;
                message.error('切换自启状态失败');
            } finally {
                switchLoading.value = false;
            }
        },
        onNegativeClick: () => {
            // 用户取消操作，恢复开关状态
            serviceStatus.value.enabled = !value;
        }
    });
}

// 导出配置
async function reloadExports() {
    await nfsApi.reloadExports();
}

// 启用/禁用NFS共享
async function enable(share: NfsShare) {
    try {
        await nfsApi.enable(share.id);
        await fetchData(); // 刷新列表
    } catch (error) {
        console.error('启用失败:', error);
        message.error('启用失败');
    }
}

async function disable(share: NfsShare) {
    try {
        await nfsApi.disable(share.id);
        await fetchData(); // 刷新列表
    } catch (error) {
        console.error('禁用失败:', error);
        message.error('禁用失败');
    }
}

async function remove(share: NfsShare, index: number) {
    try {
        await nfsApi.delete(share.id);
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
        content: '确定要从服务器拉取NFS配置到数据库吗？这将添加服务器上存在但数据库中不存在的配置。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            syncFromServerLoading.value = true;
            try {
                const result = await nfsApi.syncFromServer();
                if (result.code === 200) {
                    message.success(result.msg || '同步成功');
                    await fetchData(); // 刷新列表
                    await fetchSyncStatus(); // 刷新同步状态
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
        content: '确定要将数据库配置推送到服务器吗？这将覆盖服务器上的NFS配置文件。',
        positiveText: '确定',
        negativeText: '取消',
        onPositiveClick: async () => {
            syncToServerLoading.value = true;
            try {
                const result = await nfsApi.syncToServer();
                if (result.code === 200) {
                    message.success(result.msg || '同步成功');
                    await fetchSyncStatus(); // 刷新同步状态
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

// 查看exports配置
async function viewExportsConfig() {
    try {
        const result = await nfsApi.getExportsConfig();
        if (result.code === 200) {
            const content = (result.data || '配置文件为空') as string;
            dialog.info({
                title: 'exports配置文件内容',
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
        console.error('获取exports配置失败:', error);
        message.error('获取exports配置失败');
    }
}

// 备份exports配置
async function backupExportsConfig() {
    try {
        await nfsApi.backupExportsConfig();
    } catch (error) {
        console.error('备份exports配置失败:', error);
        message.error('备份exports配置失败');
    }
}

watch(() => route.query.filter, (newValue: any, oldValue: any) => {
    fetchData()
})

onMounted(() => {
    fetchServiceStatus();
    fetchSyncStatus();
});
</script>