<template>
  <x-page-header :subtitle="` ID:${webdav.id || ''}`">
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
    <n-form style="margin:auto; width: 480px;" :model="webdav" :rules="rules" ref="form" label-placement="top">
      <n-grid cols="1 640:1" :x-gap="24">
        <n-form-item-gi :label="t('fields.name')" path="name">
          <n-input :placeholder="t('fields.name')" v-model:value="webdav.name" />
        </n-form-item-gi>
        <n-form-item-gi :label="t('fields.account')" path="account">
          <n-input :placeholder="t('fields.account')" v-model:value="webdav.account" />
        </n-form-item-gi>
        <n-form-item-gi :label="t('fields.root_dir')" path="home">
          <n-input readonly class="dir-input" @click="dirTreeProps.show = true" :placeholder="t('fields.root_dir')"
            v-model:value="webdav.home">
            <template #suffix>
              <n-icon color='#18a058'>
                <FolderIcon />
              </n-icon>
            </template>
          </n-input>
        </n-form-item-gi>
        <n-form-item-gi :label="t('fields.remark')" path="remark">
          <n-input :placeholder="t('fields.remark')" v-model:value="webdav.remark" />
        </n-form-item-gi>
        <n-form-item-gi :label="t('fields.perms')" path="perms">
          <n-checkbox-group v-model:value="webdav.perms">
            <n-space item-style="display: flex;">
              <n-checkbox value="C" :label="t('webdav_perms.C')" />
              <n-checkbox value="U" :label="t('webdav_perms.U')" />
              <n-checkbox value="L" :label="t('webdav_perms.L')" />
              <n-checkbox value="G" :label="t('webdav_perms.G')" />
              <n-checkbox value="D" :label="t('webdav_perms.D')" />
            </n-space>
          </n-checkbox-group>
        </n-form-item-gi>
        <n-form-item-gi :label="t('fields.token')" path="token" v-if="webdav.id">
          {{ webdav.token }}
        </n-form-item-gi>
        <n-form-item-gi :label="t('fields.token')" path="token" v-else>
          {{ t('fields.auto_gen') }}
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
    <n-drawer v-model:show="dirTreeProps.show" :width="502">
      <n-drawer-content :title="t('buttons.check') + t('fields.dir')">
        <n-space vertical>
          <n-tag type="primary" size="medium" round checkable>
            {{ t('tips.selected') + t('fields.dir') }}：{{ dirTreeProps.selectedValue }}
          </n-tag>
          <n-tree :block-line="false" :checkable="false" :expand-on-click="true" :cascade="false" :show-path="true"
            :selectable="true" v-model:value="webdav.home" :data="dirTreeProps.options"
            :allow-checking-not-loaded="false" :on-load="handleDirTreeLoad"
            :on-update:expanded-keys="updatePrefixWithExpaned" :on-update:selected-keys="dirSelected"
            :default-expanded-keys="dirTreeProps.expandedKeys" :selected-keys="[dirTreeProps.selectedValue]">
             
          </n-tree></n-space>
        <template #footer>
          <n-space>
            <n-button type="primary" @click="webdav.home = dirTreeProps.selectedValue; dirTreeProps.show = false">
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
import { onMounted, reactive, ref, VNodeChild, h } from "vue";
import {
  NTag,
  NTree,
  NDrawer,
  NDrawerContent,
  NButton,
  NSpace,
  NInput,
  NInputGroup,
  NIcon,
  NForm,
  NGrid,
  NGi,
  NTreeSelect,
  NFormItemGi,
  NSwitch,
  NCheckboxGroup,
  NCheckbox,
  NRadioGroup,
  NRadio,
  NDynamicInput,
  NTooltip,
  useMessage,
  TreeSelectOption,
  TreeSelectOverrideNodeClickBehavior,
  TreeOption,
} from "naive-ui";
import {
  ArrowBackCircleOutline as BackIcon,
  SaveOutline as SaveIcon,
  CopyOutline as CopyIcon,
  Folder as FolderIcon,
  FolderOpenOutline as FolderOpenIcon,

} from "@vicons/ionicons5";
import XPageHeader from "@/components/PageHeader.vue";
import { useRoute, useRouter } from "vue-router";
import webdavApi from "@/api/nas/webdav";
import type { WebDav } from "@/api/nas/webdav";
import { useForm, emailRule, requiredRule, customRule } from "@/utils/form";
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@vueuse/core'
import { deepClone, guid } from "@/utils";

const { t } = useI18n()
const route = useRoute();
const router = useRouter();
const message = useMessage()
const listHandler = () => {
  router.push({ name: 'webdav_list' })
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
const overrideDirClick: TreeSelectOverrideNodeClickBehavior = ({ option }) => {
  if (!option.isLeaf) {
    return 'toggleExpand'
  }
  return 'default'
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
  if (option.key) {
    path = option.key.toString()
  }
  const data = await webdavApi.listDir(path);
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

const webdav = ref({ perms: ['C', 'L', 'G'], status: 1 } as any)
const rules: any = {
  name: requiredRule(),
  account: requiredRule(),
  home: requiredRule(),
};
const form = ref();

const { submit, submiting } = useForm(form, () => {
  const params = deepClone(webdav.value);
  params.perms = params.perms.join(',');
  return webdavApi.save(params)
}, () => {
  message.info(t('texts.action_success'));
  router.push({ name: 'webdav_list' })
})
const { copy, copied, isSupported } = useClipboard()

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
async function fetchData() {
  const id = route.params.id as string || ''
  if (id) {
    let r = await webdavApi.load(id);
    webdav.value = r.data as WebDav;
    webdav.value.perms = webdav.value.perms.split(',');
    dirTreeProps.selectedValue = webdav.value.home
  }
  fetchTree(webdav.value?.home)
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
