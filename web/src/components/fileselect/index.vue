<template>
    <n-input class="file-input" @click="dirTreeProps.show = true" :placeholder="placeholder" v-model:value="value">
        <template #suffix>
            <n-icon color='#18a058'>
                <FolderIcon />
            </n-icon>
        </template>
    </n-input>

    <n-drawer v-model:show="dirTreeProps.show" :width="600">
        <n-drawer-content title="选择文件或目录">
            <n-space vertical>
                <n-tag v-if="dirTreeProps.selectedValue" type="primary" size="medium" round>
                    已选择: {{ dirTreeProps.selectedValue }}
                </n-tag>
                <n-tree :block-line="false" :checkable="false" :expand-on-click="false" :cascade="false"
                    :show-path="true" :selectable="true" :data="dirTreeProps.options" :allow-checking-not-loaded="false"
                    :on-load="handleTreeLoad" key-field="key" label-field="name" :on-update:selected-keys="dirSelected"
                    :default-expanded-keys="dirTreeProps.expandedKeys"
                    :selected-keys="dirTreeProps.selectedValue ? [dirTreeProps.selectedValue] : []">
                </n-tree>
            </n-space>

            <template #footer>
                <n-space>
                    <n-button type="primary" @click="setValue">
                        确认选择
                    </n-button>
                    <n-button @click="dirTreeProps.show = false">
                        取消
                    </n-button>
                </n-space>
            </template>
        </n-drawer-content>
    </n-drawer>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h, reactive } from "vue";
import {
    NInput,
    NIcon,
    NDrawer,
    NDrawerContent,
    NButton,
    NSpace,
    NTree,
    NTag,
    useMessage,
    TreeOption,
    TreeSelectOption,
} from "naive-ui";
import {
    Folder as FolderIcon,
    Document as DocumentIcon,
    FolderOpenOutline as FolderOpenIcon,
} from "@vicons/ionicons5";
import depopsApi from '@/api/depops'
import { FileItem } from '@/api/basic/filesystem'
import fileSystemApi from '@/api/basic/filesystem'


 

const message = useMessage()

const emit = defineEmits(['update:modelValue', 'update:value']);
const props = defineProps({
    modelValue: { type: String, default: "" },
    value: { type: String, default: "" },
    placeholder: { type: String, default: "选择文件或目录" }
});
const value = computed({
    get: () => props.modelValue || props.value,
    set: (v) => {
        emit("update:modelValue", v);
        emit("update:value", v);
    },
});
const setValue = () => {
    value.value = dirTreeProps.selectedValue;
    console.log('选择的路径:', dirTreeProps.selectedValue);
    dirTreeProps.show = false;
}
const dirTreeProps = reactive({
    show: false,
    options: new Array<TreeSelectOption>(),
    selectedValue: '',
    expandedKeys: new Array<string>(),
})
const dirSelected = (keys: Array<string | number>, option: Array<TreeOption | null>, meta: { node: TreeOption | null, action: 'select' | 'unselect' }) => {
    if (meta.action == 'select') {
        dirTreeProps.selectedValue = keys[0] as string
    }
}

// 处理树节点懒加载
const handleTreeLoad = (option: TreeOption) => {
    return new Promise<void>((resolve, reject) => {
        const nodeOption = option as TreeSelectOption
        loadChildren(nodeOption).then((children) => {
            option.children = children
            resolve()
        }).catch((error) => {
            console.error('加载子节点失败:', error)
            reject(error)
        })
    })
}

// 加载子节点
const loadChildren = async (parent: TreeSelectOption): Promise<TreeSelectOption[]> => {
    try {
        const result = await fileSystemApi.list(parent.key+'', '', 'tar')
        const fileItems: FileItem[] = result.data || []

        // 排序：目录优先，然后按名称排序
        fileItems.sort((a, b) => {
            if (a.isDir && !b.isDir) return -1
            if (!a.isDir && b.isDir) return 1
            return a.name.localeCompare(b.name)
        })

        const children: TreeSelectOption[] = fileItems.map(item => {
            const node: TreeSelectOption = {
                key: item.path,
                name: item.name,
                isDir: item.isDir,
                path: item.path,
                size: item.size,
                isLeaf: !item.isDir,
                
            }

            // 设置图标 - 使用更安全的render函数语法
            node.prefix = () => h(NIcon, {
                color: item.isDir ? '#18a058' : '#666'
            }, {
                default: () => h(item.isDir ? FolderIcon : DocumentIcon)
            })

            return node
        })

        return children
    } catch (error) {
        message.error('加载文件列表失败')
        throw error
    }
}

// 加载根目录
const loadRootTree = async () => {
    try {
        const rootNode: TreeSelectOption = {
            key: '/',
            name: '根目录',
            isDir: true,
            path: '/',
            isLeaf: false,
            depth: 0
        }

        const children = await loadChildren(rootNode)
        dirTreeProps.options = children
    } catch (error) {
    }
}

onMounted(() => {
    loadRootTree()
    dirTreeProps.selectedValue = props.modelValue || props.value
})
</script>

<style scoped>
.file-input {
    cursor: pointer !important;
}

:deep(.file-input .n-input__input-el) {
    cursor: pointer !important;
}

.file-item {
    padding: 8px;
    border-radius: 4px;
    cursor: pointer;
    transition: background-color 0.2s;
}

.file-item:hover {
    background-color: var(--n-color-hover);
}

.file-item.selected {
    background-color: var(--n-color-pressed);
}
</style>
