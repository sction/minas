<template>
  <div class="wrap">
    <!-- 右边图 -->
    <main class="panel-center">
      <div class="tool-bar">
        <div class="bar">
          <div id="" style="width: 100%;height: 100%;display: flex;">
            <div class="bar-title">
              {{ props.title }}
            </div>
            <div class="bar-item btn-list">
              <n-button size="small" class="topButton" @click="back">
                <template #icon><n-icon>
                    <BackIcon />
                  </n-icon></template>
                返回
              </n-button>
              <n-button size="small" class="topButton" @click="copy">
                <template #icon><n-icon>
                    <CopyOutlined />
                  </n-icon></template>
                复制代码
              </n-button>
              <n-button size="small" class="topButton" @click="edit">
                <template #icon><n-icon>
                    <EditOutlined />
                  </n-icon></template>
                编辑代码
              </n-button>
              <n-button size="small" class="topButton" @click="center">
                <template #icon><n-icon>
                    <FullscreenOutlined />
                  </n-icon></template>
                适应屏幕
              </n-button>
              <n-button size="small" class="topButton" @click="clear">
                <template #icon><n-icon>
                    <ClearOutlined />
                  </n-icon></template>
                清空
              </n-button>
              <n-button size="small" class="topButton" @click="save">
                <template #icon><n-icon>
                    <SaveOutlined />
                  </n-icon></template>
                保存
              </n-button>
              <n-button size="small" class="topButton" @click="exec">
                <template #icon><n-icon>
                    <SendOutlined />
                  </n-icon></template>
                执行
              </n-button>
            </div>
          </div>
        </div>
      </div>
      <!-- 左侧组件 -->
      <aside class="panel-left">
        <div ref="leftToolsRef"></div>
        <n-collapse style="margin-left: 20px;" arrow-placement="right" default-expanded-names="1" accordion>
          <n-collapse-item v-for="(group, index) in groups" :key="group.id" :name="group.id" :title="group.name">
            <template v-for="(item, subindex) in group.items">
              <div v-if="!item.hidden" :key="item.id" :name="item.id" @click="() => { }"
                @mousedown="startDrag(item.id, $event)">
                {{ item.name }}
              </div>
            </template>
          </n-collapse-item>
        </n-collapse>
      </aside>
      <!-- 图设计区 -->
      <div class="x6-graph-box" style="display: flex;" ref="graphBoxRef">
        <div ref="containerRef" class="x6-graph-container" style="flex: 1;"></div>
      </div>
      <!-- 缩略图 -->
      <div ref="miniMapRef" class="mini-map"></div>
    </main>
    <!-- 右键菜单 -->
    <contextMenu :getGraph="getGraph" :funcs="funcs" @attrs="openAttrs" />
    <!-- 弹窗格式 -->
    <attribute-form ref="dialogAttributes" @change="onChange">
    </attribute-form>
    <codeeditor ref="codeEditorRef" @upgrade="updateCode"></codeeditor>
  </div>
</template>

<script setup lang="ts">
import { Cell, Edge } from "@antv/x6";
import useClipboard from "vue-clipboard3";
const { toClipboard } = useClipboard();
import { ArrowBackCircleOutline as BackIcon, SaveOutline as SaveIcon, FolderOutline as FolderIcon } from '@vicons/ionicons5';
import codeeditor from "./components/code.vue";
import { MiniMap } from '@antv/x6-plugin-minimap';
import {
  defineAsyncComponent,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  shallowRef,
  watch,
  h
} from 'vue';
import { useMessage, NIcon, NButton, NCollapse, NCollapseItem } from 'naive-ui';
import { DagGraph } from "./graph/index";
import {
  CopyOutlined,
  EditOutlined,
  FullscreenOutlined,
  ClearOutlined,
  SaveOutlined,
  SendOutlined,
  FolderOutlined,
  FolderOpenOutlined,
  AppstoreOutlined,
  FolderFilled,
  FolderOpenFilled,
} from '@vicons/antd';

defineOptions({
  name: "SDagDesigner"
})

// 移除Element Plus的CSS变量设置
// 引入组件
const contextMenu = defineAsyncComponent(() => import('./components/contextmenu/index.vue'));
const attributeForm = defineAsyncComponent(() => import('./components/attributeform/index.vue'));
// 定义父组件传过来的值
const props = defineProps({
  initNodes: { type: Array<any>, default: () => new Array<any>() },
  content: { type: Object },
  editable: { type: Boolean, default: false },
  title: { type: String },
  groups: { type: Array<any>, default: () => new Array<any>() },
  edgeFields: { type: Array<any>, default: () => new Array<any>() },
  openGroup: { type: String, default: '' },
  result: { type: Object, default: () => new Object() }
});
// 定义子组件向父组件传值/事件
const emit = defineEmits(['save', 'exec', 'back', 'change']);
// 页面元素引用
const dialogAttributes = shallowRef();
const containerRef = ref();
const leftToolsRef = ref();
const graphBoxRef = ref();
const miniMapRef = ref();
const codeEditorRef = ref();
//格式化组件数据
const procs: any = {
  "dag-edge": {
    id: 'edge',
    name: '', icon: '',
    fields: props.edgeFields
  }
};
procs.edge = procs['dag-edge'];
for (let i in props.groups) {
  const group = props.groups[i] as any;
  const items: any[] = group.items;
  for (let j in items) {
    const proc = items[j];
    procs[proc.id] = proc;
  }
}

// 获取图标组件
const getFolderIcon = (icon: string) => {
  if (icon && icon.includes('folder')) return FolderOutlined;
  return FolderOutlined;
};

const getItemIcon = (icon: string) => {
  // 这里可以根据icon字符串返回相应的组件
  // 默认返回一个通用图标
  return AppstoreOutlined;
};
export interface CurrentPageData {
  dagGraph: DagGraph | null;
  currentCell: Cell | Edge | null | undefined;
  code: any;
}
//页面数据
const state = reactive<CurrentPageData>({
  dagGraph: null,
  currentCell: null,
  code: ''
});
//获取图对象
const getGraph = () => {
  if (state.dagGraph) {
    return state.dagGraph.getGraph();
  }
  return null;
}
const funcs = () => {
  return { newId: state.dagGraph?.newId };
}
//设置居中
const center = () => {
  if (state.dagGraph) {
    state.dagGraph.getGraph().zoomTo(1);
    state.dagGraph.getGraph().centerContent();
  }
}
//清空图
const clear = () => {
  if (state.dagGraph) {
    state.dagGraph?.initGraphStart();
  }
}
const message = useMessage();
const back = () => {
  emit('back');
}
const copy = () => {
  state.dagGraph?.graph.resetSelection();
  if (state.currentCell !== null) {
    state.currentCell = null;
  }
  const content = getContent();
  console.log("content:", content)
  toClipboard(JSON.stringify(content))
  message.success('已复制到剪切板');
}
const edit = () => {
  state.dagGraph?.graph.resetSelection();
  if (state.currentCell !== null) {
    state.currentCell = null;
  }
  const content = getContent();
  state.code = JSON.stringify(content);
  codeEditorRef.value.openDialog(state.code);
}
//保存
const save = () => {
  //先取消选中
  state.dagGraph?.graph.resetSelection();
  if (state.currentCell !== null) {
    state.currentCell = null;
  }
  emit('save', getContent())
}
//运行
const exec = () => {
  emit('exec', getContent())
}

//开始拖拽时 拖起的节点
const startDrag = (shape: string, event: MouseEvent) => {
  if (state.dagGraph) {
    const graph = state.dagGraph.getGraph();
    if (graph) {
      const n = graph?.createNode({
        shape: shape
      });
      state.dagGraph.getDnd()?.start(n, event);
    }
  }
}

const updateCode = (code: string) => {
  console.log("updateCode")
  setContent(JSON.parse(code));
}
// 页面加载时
onMounted(() => {
  console.log("onMounted:", containerRef.value)
  // 构建图形
  state.dagGraph = new DagGraph(containerRef.value, props.initNodes);
  if (state.dagGraph) {

    // stencilTool(state.dagGraph.getGraph(), stencilRef.value, props.groups)
    // 初始化事件
    initEvent();
    //设置图数据
    setContent(props.content);
    //调整图大小
    resizeFn();
    //监听窗口大小调整
    window.addEventListener('resize', resizeFn);
  }
  //构建缩略图
  getGraph()?.use(
    new MiniMap({
      container: miniMapRef.value,
      width: 200,
      height: 160,
      padding: 10, graphOptions: {
        createCellView(cell) {
          // 可以返回三种类型数据
          // 1. null: 不渲染
          // 2. undefined: 使用 X6 默认渲染方式
          // 3. CellView: 自定义渲染
          if (cell.isEdge()) {
            return undefined;
          }
          if (cell.isNode()) {
            return undefined;
          }
        },
      },
    }),
  )

  // 监听图形内容变化事件 必须在最后，否则开始初始化图形是就会触发变化
  state.dagGraph?.graph?.on("cell:added", onChange);
  state.dagGraph?.graph?.on("cell:removed", onChange);
  state.dagGraph?.graph?.on("cell:change:*", onChange);
  state.dagGraph?.graph?.on("edge:connected", onChange);
});
const onChange = () => {
  emit('change', getContent());
}
// 页面销毁时
onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeFn);
});
//调整图大小
const resizeFn = () => {
  //自动计算图容器大小
  const width = graphBoxRef.value.offsetWidth - (props.editable ? 270 : 0);
  const height = graphBoxRef.value.offsetHeight - (props.editable ? 65 : 0);
  state.dagGraph?.graph.resize(width, height);
};
//监控数据变化
watch(
  () => props.content,
  (value) => {
    setContent(value);
  }
);
//监控数据变化
watch(
  () => props.result,
  (value) => {
    console.log("Result updated:", value);
    if (value && state.dagGraph) {
      // 更新各节点状态样式
      updateNodesStatus(value);
    }
  }
);
//设置图数据
const setContent = (value?: any) => {
  console.log("setContent:", value)
  if (value && value.cells) {
    state.dagGraph?.initGraphShape(value);
  } else {
    state.dagGraph?.initGraphStart();
  }
}
//获取图数据
const getContent = () => {
  return state.dagGraph?.graph.toJSON();
}
// 打开属性配置表单窗口
const openAttrs = () => {
  const cell = state.currentCell;
  if (cell) {
    dialogAttributes.value.openDialog(cell, procs);
  }
  //防止双击造成文字全选
  window.getSelection()?.removeAllRanges();
}

// 选中节点
const selectNode = (cellNode: any) => {
  console.log("selectNode cell:", cellNode);
  state.currentCell = cellNode;
}
// 初始化事件
const initEvent = async () => {
  const graph = state.dagGraph?.graph;
  //保存
  graph?.bindKey(['meta+s', 'ctrl+s'], () => {
    save();
    return false
  })
  //点击空白取消选择
  graph?.on("blank:click", async () => {
    console.log("blank:click")
    if (state.currentCell !== null) {
      state.currentCell = null;
    }
  });
  //点击节点选中
  graph?.on("cell:click", (data: any) => {
    selectNode(data.cell);
  });
  //右键菜单
  graph?.on("cell:contextmenu", (data: any) => {
    selectNode(data.cell);
  });
  //双击打开属性配置
  graph?.on("cell:dblclick", () => {
    openAttrs()
    return false;
  });


}

// 更新节点状态样式
const updateNodesStatus = (result: any) => {
  if (!result || !state.dagGraph) return;

  // 获取执行上下文中的节点状态信息
  const nodeStatus = result.nodeStatus || {};
  const nodeErrors = result.nodeErrors || {};
  const nodeResults = result.nodeResults || {};
  const graph = state.dagGraph.getGraph();

  // 遍历所有节点，更新样式
  graph.getNodes().forEach(node => {
    const nodeId = node.id.toString();
    const status = nodeStatus[nodeId];

    if (status) {
      // 更新节点数据中的状态
      let nodeData = node.getData() || {}; 

      // 设置节点状态
      nodeData.status = status;

      // 如果有错误信息，添加到节点数据中
      if (nodeErrors[nodeId]) {
        nodeData.error = nodeErrors[nodeId];
      }

      // 如果有执行结果，添加到节点数据中
      if (nodeId in nodeResults) {
        nodeData.result = nodeResults[nodeId];
      }
      //更新数据
      node.trigger("change:data", {
        cell: node,
        current: nodeData,
        options: {
          silent: false,
          overwrite: false,
          deep: true,
        }
      });
    }
  });

};
</script>

<style lang="scss" scoped>
@use "./index.scss";

// 添加Naive UI组件样式覆盖
.panel-left {
  .n-menu {
    height: 100%;
    background-color: var(--n-color-secondary);

    .n-menu-item-content {
      cursor: grab;

      &:hover {
        background-color: rgba(0, 0, 0, 0.1);
      }
    }
  }
}

.bar-item {
  .topButton {
    margin-right: 6px;
  }
}
</style>