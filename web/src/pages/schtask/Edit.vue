<template>
  <x-page-header :subtitle="` ID:${schTask.id || ''}`">
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
    <n-form style="margin:auto; width: 680px;" :model="schTask" :rules="rules" ref="form" label-placement="top">
      <n-grid :cols="24" :x-gap="24">
        <n-form-item-gi :span="24" :label="t('fields.name')" path="name">
          <n-input :placeholder="t('fields.name')" v-model:value="schTask.name" />
        </n-form-item-gi>
        
        <!-- 添加项目目录选择 -->
        <n-form-item-gi :span="24" :label="t('fields.projectdir')" path="project_dir_id">
          <n-tree-select
            v-model:value="schTask.project_dir_id"
            :options="treeData"
            :default-expanded-keys="['0']"
            :expanded-keys="expandedKeys"
            placeholder="选择项目目录"
            :node-props="treeNodeProps"
            :render-label="renderTreeLabel"
          />
        </n-form-item-gi>
        
        <n-form-item-gi :span="12" :label="t('fields.type')" path="type">
          <n-select v-model:value="schTask.type" :options="typeOptions" @update:value="onSelectType" />
        </n-form-item-gi>
        <n-form-item-gi :span="12" :label="t('fields.log_keep_num')" path="log_keep_num">
          <n-input-number :placeholder="t('fields.log_keep_num')" v-model:value="schTask.log_keep_num" />
        </n-form-item-gi>
        <n-form-item-gi :span="24" :label="t('fields.cron')" path="cron">
          <n-input @click="cronPopover = true" :placeholder="t('fields.cron')" v-model:value="schTask.cron" />
        </n-form-item-gi>

        <template v-if="schTask.type == 'SHELL'">
          <n-form-item-gi :span="24" :label="t('fields.script')" path="script">
            <n-input type="textarea" :placeholder="t('fields.script')" :autosize="{ minRows: 3, }"
              v-model:value="schTask.script.shell" />
          </n-form-item-gi>
        </template>
        <template v-if="schTask.type == 'FILE_CLEAN'">
          <n-form-item-gi :span="12" :label="t('fields.storage')">
            <n-select :placeholder="t('tips.select_storage')" v-model:value="schTask.script.storage" :options="nasOptions" clearable>
            </n-select>
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.run_mode')">
            <n-switch v-model:value="schTask.script.is_test" :checked-value="0" :unchecked-value="1">
              <template #checked>
                {{ t('fields.standard') }}
              </template>
              <template #unchecked>
                {{ t('fields.test') }}
              </template>
            </n-switch>
          </n-form-item-gi>
          <n-form-item-gi :span="24" :label="t('fields.work_dir')">
            <!-- <n-input :placeholder="t('fields.work_dir')" v-model:value="schTask.script.work_dir" /> -->
            <select-dir :placeholder="t('tips.select_work_dir')" :nas="schTask.script.storage" v-model="schTask.script.work_dir" />
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.includes')">
            <n-input type="textarea" :placeholder="t('tips.include_rule')" :autosize="{ minRows: 2, }"
              v-model:value="schTask.script.includes" />
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.excludes')">
            <n-input type="textarea" :placeholder="t('tips.exclude_rule')" :autosize="{ minRows: 2, }"
              v-model:value="schTask.script.excludes" />
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.keep_num')">
            <n-input-number :placeholder="t('tips.min_keep_files')" v-model:value="schTask.script.keep_num" />
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.keep_days')">
            <n-input-number :placeholder="t('tips.min_keep_days')" v-model:value="schTask.script.offset_day" />
          </n-form-item-gi>
        </template>
        <template v-if="schTask.type == 'FILE_BACKUP'">
          <n-form-item-gi :span="18" :label="t('fields.backup_method')">
              <n-radio-group v-model:value="schTask.script.type" name="radiobuttongroup1">
                <n-radio-button v-for="option in backupTypeOptions" :key="option.value" :value="option.value"
                  :label="option.label" />
              </n-radio-group>
          </n-form-item-gi>
          <n-form-item-gi :span="6" :label="t('fields.empty_dir')">
            <n-switch v-model:value="schTask.script.is_create_dir" :checked-value="1" :unchecked-value="0">
              <template #checked>
                {{ t('fields.sync_create') }}
              </template>
              <template #unchecked>
                {{ t('fields.no_create') }}
              </template>
            </n-switch>
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.source_storage')">
            <n-select :placeholder="t('tips.select_source_storage')" v-model:value="schTask.script.source_nas_id" :options="nasOptions"
              clearable>
            </n-select>
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.source_path')">
            <!-- <n-input :placeholder="t('tips.input_source_path')" v-model:value="schTask.script.source" /> -->
            <select-dir :placeholder="t('tips.input_source_path')" :nas="schTask.script.source_nas_id" v-model="schTask.script.source" />
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.target_storage')">
            <n-select :placeholder="t('tips.select_target_storage')" v-model:value="schTask.script.target_nas_id" :options="nasOptions"
              clearable>
            </n-select>
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.target_path')">
            <!-- <n-input :placeholder="t('tips.input_target_path')" v-model:value="schTask.script.target" /> -->
            <select-dir :placeholder="t('tips.input_target_path')" :nas="schTask.script.target_nas_id" v-model="schTask.script.target" />
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.includes')">
            <n-input type="textarea" :placeholder="t('tips.include_rule_simple')" :autosize="{ minRows: 2, }"
              v-model:value="schTask.script.includes" />
          </n-form-item-gi>
          <n-form-item-gi :span="12" :label="t('fields.excludes')">
            <n-input type="textarea" :placeholder="t('tips.exclude_rule_simple')" :autosize="{ minRows: 2, }"
              v-model:value="schTask.script.excludes" />
          </n-form-item-gi>
        </template>
        <n-form-item-gi :span="24" :label="t('fields.remark')" path="remark">
          <n-input type="textarea" :placeholder="t('fields.remark')" :autosize="{ minRows: 2, }"
            v-model:value="schTask.remark" />
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
    <n-drawer v-model:show="cronPopover" :width="802">
      <n-drawer-content>
        <template #header>
          {{ t('fields.set_cron') }}
        </template>
        <cron v-model="schTask.cron" />
        <template #footer>
        </template>
      </n-drawer-content>
    </n-drawer>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, VNodeChild, h } from "vue";
import cron from "@/components/cron/index.vue";
import selectDir from "@/components/selectdir/index.vue";
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
  TreeOption,
  TreeSelectOverrideNodeClickBehavior,
  NSelect,
  NInputNumber,
  NRadioButton,
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
import schTaskApi from "@/api/sch/task";
import externalNasApi from "@/api/nas/external";
import projectDirApi from "@/api/basic/projectdir";
import type { SchTask } from "@/api/sch/task";
import type { ProjectDirItem } from "@/api/basic/projectdir";
import { typeOptions, backupTypeOptions } from "@/api/sch/task";
import { useForm, emailRule, requiredRule, customRule } from "@/utils/form";
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@vueuse/core'
import { deepClone, guid } from "@/utils";
import { getPageData } from "@/utils/render";

const { t } = useI18n()
const route = useRoute();
const router = useRouter();
const message = useMessage()
const cronPopover = ref(false)
const cronValue = ref("0 1 2 3 * ? *")
const crontabFill = (v: string) => {
  schTask.value.cron = v;
}

const id = route.params.id as string || ''
// 检查URL查询参数是否包含项目目录ID
const projectDirId = getPageData(route.path, 'project_dir_id') as string

const listHandler = () => {
  router.push({ name: 'schtask_list'});
}
const schTask = ref({ cron: "", log_keep_num: 30, type: 'SHELL', script: { type: 1, shell: '' } as any } as any)
const rules: any = {
  name: requiredRule(),
  type: requiredRule(),
  home: requiredRule(),
};
const nasOptions = ref(new Array<{ label: string; value: string; }>());
const form = ref({ script: {} as any } as any);

// 项目目录树相关状态
const treeData = ref<TreeOption[]>([]);
const treeLoading = ref(false);
const expandedKeys = ref<any[]>([projectDirId]);

// 项目目录树加载
const loadProjectDirTree = async () => {
  treeLoading.value = true;
  try {
    // 获取task类型的项目目录树
    const res = await projectDirApi.getTree(); // 1表示模块类型
    if (res.data && Array.isArray(res.data)) {
      treeData.value = convertToTreeOptions(res.data);
      if (treeData.value.length > 0) {
        expandedKeys.value = [treeData.value[0].key];
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

// 树节点渲染
const renderTreeLabel = (info: { option: TreeOption }) => {
  return h(
    "div",
    {
      style: "display: flex; align-items: center;"
    },
    [
      h(
        NIcon,
        {
          style: "margin-right: 4px"
        },
        {
          default: () => h(FolderIcon)
        }
      ),
      info.option.label as string
    ]
  );
};

// 树节点属性
const treeNodeProps = ({ option }: { option: TreeOption }) => {
  return {
    style: "cursor: pointer;"
  };
};

const { submit, submiting } = useForm(form, () => {
  const params = deepClone(schTask.value);
  params.script = JSON.stringify(schTask.value.script)
  return schTaskApi.save(params)
}, () => {
  message.info(t('texts.action_success'));
  router.push({ name: 'schtask_list' })
})
const { copy, copied, isSupported } = useClipboard()
const onSelectType = () => {
  console.log("onSelectType:", schTask.value)
}

async function fetchData() {
  // 加载项目目录树
  await loadProjectDirTree();
  
  
  if (id) {
    const r = await schTaskApi.load(id);
    schTask.value = r.data as SchTask;
    schTask.value.script = JSON.parse(schTask.value.script)
  } else if (projectDirId) {
    schTask.value.project_dir_id =  projectDirId
  }
  
  const nasList = (await externalNasApi.search({
    filters: "",
    page: 1,
    size: 0,
  }));
  nasOptions.value = []
  nasOptions.value.push({ "label": t('fields.local_storage'), "value": "" })
  if (nasList.data) {
    for (let i in nasList.data) {
      const item = nasList.data[i];
      nasOptions.value.push({ "label": item.name, "value": item.rc_name })
    }
  }
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
