<template>
    <x-page-header :subtitle="` ID:${nfsShare.id || ''}`">
        <template #action>
            <n-button secondary size="small" @click="listHandler">
                <template #icon>
                    <n-icon>
                        <back-icon />
                    </n-icon>
                </template>
                {{ t('buttons.return') }}
            </n-button>
        </template>
    </x-page-header>
    <n-space class="page-body" vertical :size="12">
        <n-form style="margin:auto; width: 580px;" :model="nfsShare" :rules="rules" ref="form" label-placement="top">
            <n-grid cols="1 640:1" :x-gap="24">
                <n-form-item-gi label="共享名称" path="name">
                    <n-input placeholder="输入共享名称" v-model:value="nfsShare.name" />
                </n-form-item-gi>

                <n-form-item-gi label="共享路径" path="path">
                    <n-input readonly class="dir-input" @click="dirTreeProps.show = true" placeholder="选择共享目录"
                        v-model:value="nfsShare.path">
                        <template #suffix>
                            <n-icon color='#18a058'>
                                <FolderIcon />
                            </n-icon>
                        </template>
                    </n-input>
                </n-form-item-gi>

                <n-form-item-gi label="客户端IP" path="clientIp">
                    <n-input placeholder="如：192.168.1.0/24 或 * (允许所有)" v-model:value="nfsShare.clientIp" />
                    <template #feedback>
                        支持IP地址、网段或通配符，如：192.168.1.100、192.168.1.0/24、* 等
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="读写权限" path="permission">
                    <n-radio-group v-model:value="nfsShare.permission">
                        <n-space>
                            <n-radio value="ro">只读 (ro)</n-radio>
                            <n-radio value="rw">读写 (rw)</n-radio>
                        </n-space>
                    </n-radio-group>
                </n-form-item-gi>

                <n-form-item-gi label="同步模式" path="sync">
                    <n-radio-group v-model:value="nfsShare.sync">
                        <n-space>
                            <n-radio value="">无</n-radio>
                            <n-radio value="sync">同步</n-radio>
                            <n-radio value="async">异步</n-radio>
                        </n-space>
                    </n-radio-group>
                    <template #feedback>
                        sync: 保证数据一致性，性能较低；async: 性能更高，但可能在崩溃时丢失数据
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="Root权限映射" path="rootSquash">
                    <n-radio-group v-model:value="nfsShare.rootSquash">
                        <n-space>
                            <n-radio value="">无</n-radio>
                            <n-radio value="root_squash">root_squash</n-radio>
                            <n-radio value="no_root_squash">no_root_squash</n-radio>
                            <n-radio value="all_squash">all_squash</n-radio>
                        </n-space>
                    </n-radio-group>
                    <template #feedback>
                        root_squash: 将客户端root映射为匿名用户，推荐用于安全环境；no_root_squash: 保留客户端root权限，仅在可信环境下使用；all_squash: 将所有用户映射为匿名用户，适用于公共共享
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="子树检查" path="subtreeCheck">
                    <n-radio-group v-model:value="nfsShare.subtreeCheck">
                        <n-space>
                            <n-radio value="">无</n-radio>
                            <n-radio value="subtree_check">subtree_check</n-radio>
                            <n-radio value="no_subtree_check">no_subtree_check</n-radio>
                        </n-space>
                    </n-radio-group>
                    <template #feedback>
                        subtree_check: 启用检查，安全性更高但性能稍低；no_subtree_check: 禁用检查，性能更好，适合导出整个文件系统
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="其他选项" path="options">
                    <n-input placeholder="如：insecure" v-model:value="nfsShare.options" />
                    <template #feedback>
                        多个选项用逗号分隔，常用选项：insecure, no_wdelay 等<br>
                        insecure: 允许>1024端口连接；secure：仅允许客户端使用&lt;1024端口（默认）<br>
                        no_wdelay: 立即执行写操作（需配合sync）；wdelay: 合并写操作（默认）<br>
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="备注" path="remark">
                    <n-input placeholder="输入备注信息" v-model:value="nfsShare.remark" />
                </n-form-item-gi>

                <n-gi :span="2">
                    <n-button :disabled="submiting" :loading="submiting" @click.prevent="submit" type="primary">
                        <template #icon>
                            <n-icon>
                                <save-icon />
                            </n-icon>
                        </template>
                        {{ t('buttons.save') }}
                    </n-button>
                </n-gi>
            </n-grid>
        </n-form>

        <!-- 目录选择器 -->
        <n-drawer v-model:show="dirTreeProps.show" :width="502">
            <n-drawer-content title="选择共享目录">
                <n-space vertical>
                    <n-tag type="primary" size="medium" round>
                        已选择目录：{{ dirTreeProps.selectedValue }}
                    </n-tag>
                    <n-tree :block-line="false" :checkable="false" :expand-on-click="true" :cascade="false"
                        :show-path="true" :selectable="true" v-model:value="nfsShare.path" :data="dirTreeProps.options"
                        :allow-checking-not-loaded="false" :on-load="handleDirTreeLoad"
                        :on-update:expanded-keys="updatePrefixWithExpaned" :on-update:selected-keys="dirSelected"
                        :default-expanded-keys="dirTreeProps.expandedKeys"
                        :selected-keys="[dirTreeProps.selectedValue]">
                    </n-tree>
                </n-space>
                <template #footer>
                    <n-space>
                        <n-button type="primary"
                            @click="nfsShare.path = dirTreeProps.selectedValue; dirTreeProps.show = false">
                            {{ t('buttons.confirm') }}
                        </n-button>
                        <n-button @click="dirTreeProps.show = false">
                            {{ t('buttons.cancel') }}
                        </n-button>
                    </n-space>
                </template>
            </n-drawer-content>
        </n-drawer>
    </n-space>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, h } from "vue";
import {
    NTag,
    NTree,
    NDrawer,
    NDrawerContent,
    NButton,
    NSpace,
    NInput,
    NIcon,
    NForm,
    NGrid,
    NGi,
    NFormItemGi,
    NSwitch,
    NRadioGroup,
    NRadio,
    NSelect,
    useMessage,
    TreeSelectOption,
    TreeOption,
} from "naive-ui";
import {
    ArrowBackCircleOutline as BackIcon,
    SaveOutline as SaveIcon,
    Folder as FolderIcon,
    FolderOpenOutline as FolderOpenIcon,
} from "@vicons/ionicons5";
import XPageHeader from "@/components/PageHeader.vue";
import { useRoute, useRouter } from "vue-router";
import nfsApi from "@/api/nas/nfs";
import type { NfsShare } from "@/api/nas/nfs";
import { useForm, requiredRule } from "@/utils/form";
import { useI18n } from 'vue-i18n'
import { deepClone } from "@/utils";

const { t } = useI18n()
const route = useRoute();
const router = useRouter();
const message = useMessage()

const listHandler = () => {
    router.push({ name: 'nfs_list' })
}

const dirTreeProps = reactive({
    show: false,
    options: new Array<TreeSelectOption>(),
    selectedValue: '',
    expandedKeys: new Array<string>(),
})

const nfsShare = ref<NfsShare>({
    id: 0,  // 修改为number类型的默认值
    name: '',
    path: '',
    clientIp: '*',
    permission: 'ro',
    sync: 'sync',
    rootSquash: '',        // 修改为空字符串，表示不设置
    subtreeCheck: 'subtree_check',      // 修改为默认启用子树检查
    options: '',
    remark: '',
    is_disable: 0,  // 使用is_disable字段，0表示启用
    created_at: '',
    updated_at: '',
    created_by: 0,
    updated_by: 0,
})

const rules = {
    name: requiredRule(),
    path: requiredRule(),
    permission: requiredRule()
}

const form = ref()

const { submit, submiting } = useForm(form, () => {
    const params = deepClone(nfsShare.value);
    return nfsApi.save(params)
}, () => {
    message.success('保存成功');
    router.push({ name: 'nfs_list' })
})

// 目录树相关函数
const updatePrefixWithExpaned = (
    _keys: Array<string | number>,
    _option: Array<TreeOption | null>,
    meta: {
        node: TreeOption | null
        action: 'expand' | 'collapse' | 'filter'
    }
) => {
    if (!meta.node) return
    switch (meta.action) {
        case 'expand':
            meta.node.prefix = () =>
                h(NIcon, { color: '#18a058' }, {
                    default: () => h(FolderOpenIcon)
                })
            break
        case 'collapse':
            meta.node.prefix = () =>
                h(NIcon, { color: '#18a058' }, {
                    default: () => h(FolderIcon)
                })
            break
    }
}

const dirSelected = (keys: Array<string | number>, option: Array<TreeOption | null>, meta: { node: TreeOption | null, action: 'select' | 'unselect' }) => {
    if (meta.action == 'select') {
        dirTreeProps.selectedValue = keys[0] as string
    }
}

const handleDirTreeLoad = (option: TreeSelectOption) => {
    return new Promise<void>((resolve, reject) => {
        getChildrenDir(option).then((data) => {
            option.children = data;
            resolve()
        }).catch(() => {
            reject()
        })
    })
}

const getChildrenDir = async (option: TreeSelectOption) => {
    let path = ""
    if (option.key) {
        path = option.key.toString()
    }
    const data = await nfsApi.listDir(path);
    const children = new Array<TreeSelectOption>();
    if (data.code == 200) {
        for (var i = 0; i < data.data.length; i++) {
            const item = data.data[i]
            item.isLeaf = false
            item.depth = (option as { depth: number }).depth + 1
            item.prefix = () =>
                h(NIcon, { color: '#18a058' }, { default: () => h(FolderIcon) })
            children.push(item)
        }
    }
    return children;
}

// 递归加载目录树
async function autoLazyLoad(items: TreeSelectOption[] = [], paths: string[] = [], index: number = 2) {
    if (paths.length > 1 && index < paths.length) {
        const path = paths.slice(0, index).join('/');
        const parent = items.find(item => item.key == path)
        if (parent) {
            await handleDirTreeLoad(parent)
            if (parent.children && index < paths.length + 1) {
                dirTreeProps.expandedKeys.push(path)
                await autoLazyLoad(parent.children, paths, index + 1)
            }
        }
    }
}

// 加载根目录
async function fetchTree(path?: string) {
    const children = (await getChildrenDir({ label: '', key: '', depth: 0, isLeaf: false })) || []
    dirTreeProps.options = children
    if (path) {
        await autoLazyLoad(children, path.split('/'), 2)
    }
}

async function fetchData() {
    const id = route.params.id as string || ''
    if (id) {
        const result = await nfsApi.load(parseInt(id));
        nfsShare.value = result.data as any;
        dirTreeProps.selectedValue = nfsShare.value.path
    }
    fetchTree(nfsShare.value?.path)
}

onMounted(fetchData);
</script>

<style scoped>
.dir-input {
    cursor: pointer !important;
}

:deep(.dir-input .n-input__input-el) {
    cursor: pointer !important;
}
</style>