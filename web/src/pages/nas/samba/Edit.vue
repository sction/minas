<template>
    <x-page-header :subtitle="` ID:${sambaShare.id || ''}`">
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
        <n-form style="margin:auto; width: 580px;" :model="sambaShare" :rules="rules" ref="form" label-placement="top">
            <n-grid cols="1 640:1" :x-gap="24">
                <n-form-item-gi label="共享名称" path="name">
                    <n-input placeholder="输入共享名称（英文字母、数字、下划线）" v-model:value="sambaShare.name" />
                </n-form-item-gi>

                <n-form-item-gi label="共享路径" path="path">
                    <n-input readonly class="dir-input" @click="dirTreeProps.show = true" placeholder="选择共享目录"
                        v-model:value="sambaShare.path">
                        <template #suffix>
                            <n-icon color='#18a058'>
                                <FolderIcon />
                            </n-icon>
                        </template>
                    </n-input>
                </n-form-item-gi>

                <n-form-item-gi label="共享描述" path="comment">
                    <n-input placeholder="输入共享描述" v-model:value="sambaShare.comment" />
                </n-form-item-gi>

                <n-form-item-gi label="是否可浏览" path="browseable">
                    <n-switch v-model:value="sambaShare.browseable">
                        <template #checked>
                            可浏览
                        </template>
                        <template #unchecked>
                            不可浏览
                        </template>
                    </n-switch>
                    <template #feedback>
                        启用后，此共享将在网络邻居中显示
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="只读权限" path="readOnly">
                    <n-switch v-model:value="sambaShare.readOnly">
                        <template #checked>
                            只读
                        </template>
                        <template #unchecked>
                            读写
                        </template>
                    </n-switch>
                    <template #feedback>
                        启用后，用户只能读取文件，无法修改或删除
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="来宾访问" path="guestOk">
                    <n-switch v-model:value="sambaShare.guestOk">
                        <template #checked>
                            允许
                        </template>
                        <template #unchecked>
                            禁止
                        </template>
                    </n-switch>
                    <template #feedback>
                        启用后，无需用户名密码即可访问
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="有效用户" path="validUsers">
                    <n-select 
                        multiple 
                        clearable
                        placeholder="选择可以访问此共享的用户"
                        :options="sambaUserOptions"
                        :value="validUsersArray"
                        @update:value="updateValidUsers"
                        :loading="sambaUserLoading"
                        filterable
                    />
                    <template #feedback>
                        指定可以访问此共享的用户。留空表示所有用户可访问
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="写权限用户" path="writeList">
                    <n-select 
                        multiple 
                        clearable
                        placeholder="选择具有写权限的用户"
                        :options="sambaUserOptions"
                        :value="writeListUsers"
                        @update:value="updateWriteList"
                        :loading="sambaUserLoading"
                        filterable
                    />
                    <template #feedback>
                        在只读模式下，指定具有写权限的用户
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="读权限用户" path="readList">
                    <n-select 
                        multiple 
                        clearable
                        placeholder="选择只有读权限的用户"
                        :options="sambaUserOptions"
                        :value="readListUsers"
                        @update:value="updateReadList"
                        :loading="sambaUserLoading"
                        filterable
                    />
                    <template #feedback>
                        在读写模式下，指定只有读权限的用户
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="创建文件权限" path="createMask">
                    <n-input placeholder="如：0744" v-model:value="sambaShare.createMask" />
                    <template #feedback>
                        新创建文件的权限掩码，默认：0744
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="创建目录权限" path="directoryMask">
                    <n-input placeholder="如：0755" v-model:value="sambaShare.directoryMask" />
                    <template #feedback>
                        新创建目录的权限掩码，默认：0755
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="强制用户" path="forceUser">
                    <n-input placeholder="用户名（将自动填充为目录所有者）" v-model:value="sambaShare.forceUser" />
                    <template #feedback>
                        强制所有操作以指定用户身份执行。已自动设置为共享目录的所有者
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="强制组" path="forceGroup">
                    <n-input placeholder="组名" v-model:value="sambaShare.forceGroup" />
                    <template #feedback>
                        强制所有操作以指定组身份执行
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="可用空间限制" path="availableSpace">
                    <n-input placeholder="如：100G" v-model:value="sambaShare.availableSpace" />
                    <template #feedback>
                        限制此共享的最大可用空间，支持单位：B、K、M、G、T
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="其他选项" path="options">
                    <n-input type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" 
                        placeholder="其他Samba配置选项，每行一个" v-model:value="sambaShare.options" />
                    <template #feedback>
                        额外的Samba配置选项，每行格式：参数名 = 参数值
                    </template>
                </n-form-item-gi>

                <n-form-item-gi label="备注" path="remark">
                    <n-input placeholder="输入备注信息" v-model:value="sambaShare.remark" />
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
                        :show-path="true" :selectable="true" v-model:value="sambaShare.path" :data="dirTreeProps.options"
                        :allow-checking-not-loaded="false" :on-load="handleDirTreeLoad"
                        :on-update:expanded-keys="updatePrefixWithExpaned" :on-update:selected-keys="dirSelected"
                        :default-expanded-keys="dirTreeProps.expandedKeys"
                        :selected-keys="[dirTreeProps.selectedValue]">
                    </n-tree>
                </n-space>
                <template #footer>
                    <n-space>
                        <n-button type="primary"
                            @click="confirmDirSelection">
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
import { onMounted, reactive, ref, h, watch } from "vue";
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
import sambaApi from "@/api/nas/samba";
import type { SambaShare } from "@/api/nas/samba";
import { useForm, requiredRule } from "@/utils/form";
import { useI18n } from 'vue-i18n'
import { deepClone } from "@/utils";

const { t } = useI18n()
const route = useRoute();
const router = useRouter();
const message = useMessage()

const listHandler = () => {
    router.push({ name: 'samba_shares' })
}

const dirTreeProps = reactive({
    show: false,
    options: new Array<TreeSelectOption>(),
    selectedValue: '',
    expandedKeys: new Array<string>(),
})

// 用户相关
const sambaUserLoading = ref(false)
const sambaUserOptions = ref<Array<{ label: string; value: string }>>([])
const validUsersArray = ref<string[]>([])
const writeListUsers = ref<string[]>([])
const readListUsers = ref<string[]>([])

// 更新有效用户列表
const updateValidUsers = (value: string[]) => {
    validUsersArray.value = value
    sambaShare.value.validUsers = value.join(',')
}

// 更新写权限用户列表
const updateWriteList = (value: string[]) => {
    writeListUsers.value = value
    sambaShare.value.writeList = value.join(',')
}

// 更新读权限用户列表
const updateReadList = (value: string[]) => {
    readListUsers.value = value
    sambaShare.value.readList = value.join(',')
}

const sambaShare = ref<SambaShare>({
    id: 0,
    name: '',
    path: '',
    comment: '',
    browseable: true,
    readOnly: false,
    guestOk: false,
    validUsers: '',
    writeList: '',
    readList: '',
    createMask: '0744',
    directoryMask: '0755',
    forceUser: '',
    forceGroup: '',
    availableSpace: '',
    options: '',
    remark: '',
    is_disable: 0,
    created_at: '',
    updated_at: '',
    created_by: 0,
    updated_by: 0,
})

const rules = {
    name: [
        { required: true, message: '请输入共享名称', trigger: 'blur' },
        { 
            pattern: /^[a-zA-Z0-9_]+$/, 
            message: '共享名称只能包含字母、数字和下划线', 
            trigger: 'blur' 
        }
    ],
    path: requiredRule(),
}

const form = ref()

const { submit, submiting } = useForm(form, () => {
    const params = deepClone(sambaShare.value);
    return sambaApi.save(params)
}, () => {
    message.success('保存成功');
    router.push({ name: 'samba_shares' })
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
    const data = await sambaApi.listDir(path);
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

// 获取Samba用户列表
async function fetchSambaUsers() {
    sambaUserLoading.value = true
    try {
        const result = await sambaApi.listSambaUsers()
        if (result.code === 200) {
            // 检查数据格式
            let users: any[] = []
            if (Array.isArray(result.data)) {
                users = result.data
            } else if (result.data && 'data' in result.data) {
                users = result.data.data || []
            }
            
            sambaUserOptions.value = users.map((user: any) => ({
                label: user.username,
                value: user.username
            }))
        }
    } catch (error) {
        console.error('获取Samba用户列表失败:', error)
    } finally {
        sambaUserLoading.value = false
    }
}

// 获取目录所有者并自动填充强制用户字段
async function fetchDirOwner(path: string) {
    if (!path) return
    
    try {
        const result = await sambaApi.getDirOwner(path)
        if (result.code === 200 && result.data) {
            // 只有在强制用户字段为空时才自动填充
            if (!sambaShare.value.forceUser) {
                sambaShare.value.forceUser = result.data.owner
            }
            // 可选：同时设置强制组
            if (!sambaShare.value.forceGroup) {
                sambaShare.value.forceGroup = result.data.group
            }
        }
    } catch (error) {
        console.error('获取目录所有者失败:', error)
    }
}

// 监听路径变化，自动获取目录所有者
watch(() => sambaShare.value.path, (newPath) => {
    if (newPath) {
        fetchDirOwner(newPath)
    }
}, { immediate: false })

// 确认目录选择
const confirmDirSelection = () => {
    sambaShare.value.path = dirTreeProps.selectedValue
    dirTreeProps.show = false
    // 选择目录后自动获取所有者
    if (dirTreeProps.selectedValue) {
        fetchDirOwner(dirTreeProps.selectedValue)
    }
}

async function fetchData() {
    const id = route.params.id as string || ''
    
    // 同时获取Samba用户列表
    await fetchSambaUsers()
    
    if (id) {
        const result = await sambaApi.load(parseInt(id));
        sambaShare.value = result.data as any;
        dirTreeProps.selectedValue = sambaShare.value.path
        
        // 将逗号分隔的字符串转换为数组
        validUsersArray.value = sambaShare.value.validUsers ? sambaShare.value.validUsers.split(',').filter(u => u.trim()) : []
        writeListUsers.value = sambaShare.value.writeList ? sambaShare.value.writeList.split(',').filter(u => u.trim()) : []
        readListUsers.value = sambaShare.value.readList ? sambaShare.value.readList.split(',').filter(u => u.trim()) : []
        
        // 如果是编辑模式且没有设置强制用户，则自动获取目录所有者
        if (sambaShare.value.path && !sambaShare.value.forceUser) {
            await fetchDirOwner(sambaShare.value.path)
        }
    }
    fetchTree(sambaShare.value?.path)
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