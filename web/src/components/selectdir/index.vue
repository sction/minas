<template>
  <n-input class="dir-input" @click="dirTreeProps.show = true" :placeholder="placeholder" v-model:value="value">
    <template #suffix>
      <n-icon color='#18a058'>
        <FolderIcon />
      </n-icon>
    </template>
  </n-input>
  <n-drawer v-model:show="dirTreeProps.show" :width="502">
    <n-drawer-content :title="t('buttons.check') + t('fields.dir')">
      <n-space vertical>
        <n-tag type="primary" size="medium" round checkable>
          {{ t('tips.selected') + t('fields.dir') }}：{{ dirTreeProps.selectedValue }}
        </n-tag>
        <n-tree :block-line="false" :checkable="false" :expand-on-click="true" :cascade="false" :show-path="true"
          :selectable="true" v-model:value="dirTreeProps.selectedValue" :data="dirTreeProps.options"
          :allow-checking-not-loaded="false" :on-load="handleDirTreeLoad"
          :on-update:expanded-keys="updatePrefixWithExpaned" key-field="key" label-field="Name"
          :on-update:selected-keys="dirSelected" :default-expanded-keys="dirTreeProps.expandedKeys"
          :selected-keys="[dirTreeProps.selectedValue]">
        </n-tree></n-space>
      <template #footer>
        <n-space>
          <n-button type="primary" @click="setValue">
            {{ t('buttons.confirm') }}
          </n-button>
          <n-button @click="dirTreeProps.show = false">
            {{ t('buttons.cancel') }}
          </n-button>
        </n-space>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { onMounted, reactive, h, computed, ref, watch } from "vue";
import {
  NTag,
  NTree,
  NDrawer,
  NDrawerContent,
  NButton,
  NSpace,
  NInput,
  NIcon,
  TreeSelectOption,
  TreeOption,
} from "naive-ui";
import {
  Folder as FolderIcon,
  FolderOpenOutline as FolderOpenIcon,

} from "@vicons/ionicons5";
import externalNasApi from "@/api/nas/external";
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const emit = defineEmits(['update:modelValue']);
const props = defineProps({
  modelValue: { type: String, default: "" },
  nas: { type: String, default: "" },
  placeholder: { type: String, default: "" },
});
const value = computed({
  get: () => props.modelValue,
  set: (v) => emit("update:modelValue", v),
});
const setValue = () => {
  value.value = dirTreeProps.selectedValue;
  dirTreeProps.show = false;
}
const dirTreeProps = reactive({
  show: false,
  options: new Array<TreeSelectOption>(),
  selectedValue: '',
  expandedKeys: new Array<string>(),
})
const updatePrefixWithExpaned = (
  _keys: Array<string | number>,
  _option: Array<TreeOption | null>,
  meta: {
    node: TreeOption | null
    action: 'expand' | 'collapse' | 'filter'
  }
) => {
  if (!meta.node)
    return
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

// TreeSelectOption[] | undefined
const getChildrenDir = async (option: TreeSelectOption) => {
  let path = ""
  console.log("option:", option)
  if (option.key) {
    path = option.key.toString()
  }
  const data = await externalNasApi.listDir(props.nas, path);
  const children = new Array<TreeSelectOption>();
  if (data.code == 200) {
    for (var i = 0; i < data.data.length; i++) {
      const item = data.data[i]
      item.isLeaf = false
      item.key = path + "/" + item.Path
      item.depth = (option as { depth: number }).depth + 1
      item.prefix = () =>
        h(NIcon, { color: '#18a058' }, { default: () => h(FolderIcon) })

      children.push(item)
    }
  }
  return children;
}



// 递归加载目录树 选中的上级目录
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
// 加载跟目录
async function fetchTree(home?: string) {
  const children = (await getChildrenDir({ label: '', key: '', depth: 0, isLeaf: false })) || []
  dirTreeProps.options = children
  if (home) {
    //如果有选中的目录则自动加载子目录
    await autoLazyLoad(children, home.split('/'), 2)
  }
}
watch(
  () => props.nas,
  (v) => {
    fetchTree(props.modelValue)
  }
);
onMounted(() => {
  dirTreeProps.selectedValue = props.modelValue
  fetchTree(props.modelValue)
})
</script>
<style scoped>
.dir-input {
  cursor: pointer !important;
}

:deep(.dir-input .n-input__input-el) {
  cursor: pointer !important;
}
</style>
