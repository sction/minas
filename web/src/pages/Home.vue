<template>
  <n-space class="page-body" vertical :size="12">
    <!-- 数量统计块 -->
    <n-grid cols="1 s:2 m:3 l:6" responsive="screen" :x-gap="12" :y-gap="12">
      <n-grid-item>
        <n-card>
          <n-statistic label="项目" :value="dashboardStats.projectCount">
            <template #prefix>
              <n-icon color="#18a058">
                <FolderOpenOutline />
              </n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="WebDAV共享" :value="dashboardStats.webdavCount">
            <template #prefix>
              <n-icon color="#2080f0">
                <CloudOutline />
              </n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="NFS共享" :value="dashboardStats.nfsCount">
            <template #prefix>
              <n-icon color="#722ed1">
                <LayersOutline />
              </n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="Samba共享" :value="dashboardStats.sambaCount">
            <template #prefix>
              <n-icon color="#fa8c16">
                <FolderOpenOutline />
              </n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="外部存储" :value="dashboardStats.externalStorageCount">
            <template #prefix>
              <n-icon color="#f0a020">
                <LinkOutline />
              </n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="计划任务" :value="dashboardStats.scheduledTaskCount">
            <template #prefix>
              <n-icon color="#d03050">
                <TimeOutline />
              </n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 图表区域 -->
    <n-grid cols="1 m:3" responsive="screen" :x-gap="12" :y-gap="12">
      <!-- 任务类型饼图 -->
      <n-grid-item>
        <n-card title="任务类型分布" size="small">
          <div ref="taskTypePieRef" style="width: 100%; height: 300px;"></div>
        </n-card>
      </n-grid-item>
      
      <!-- 任务执行状态饼图 -->
      <n-grid-item>
        <n-card title="任务执行状态" size="small">
          <div ref="taskStatusBarRef" style="width: 100%; height: 300px;"></div>
        </n-card>
      </n-grid-item>
      
      <!-- 各项目任务数量柱状图 -->
      <n-grid-item>
        <n-card title="各项目任务数量" size="small">
          <div ref="taskProjectBarRef" style="width: 100%; height: 300px;"></div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 任务执行趋势折线图 -->
    <n-card title="近一个月任务执行趋势" size="small">
      <div ref="taskTrendLineRef" style="width: 100%; height: 330px;"></div>
    </n-card>

    <n-hr style="margin: 4px 0" />
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, ref, nextTick } from "vue";
import {
  NSpace,
  NHr,
  NGrid,
  NGridItem,
  NCard,
  NStatistic,
  NIcon,
  useMessage
} from "naive-ui";
import {
  FolderOpenOutline,
  CloudOutline,
  LinkOutline,
  TimeOutline,
  LayersOutline
} from "@vicons/ionicons5";
import * as echarts from 'echarts';
import dashboardApi from "@/api/dashboard";
import type { DashboardStats, TaskTypeStats, TaskExecutionStats, TaskTrendData, TaskProjectStats } from "@/api/dashboard";

const message = useMessage();

// 数据
const dashboardStats = ref<DashboardStats>({
  projectCount: 0,
  webdavCount: 0,
  externalStorageCount: 0,
  scheduledTaskCount: 0,
  nfsCount: 0,
  sambaCount: 0
});

const taskTypeStats = ref<TaskTypeStats[]>([]);
const taskExecutionStats = ref<TaskExecutionStats>({
  success: 0,
  failed: 0,
  running: 0,
  total: 0
});
const taskTrendData = ref<TaskTrendData[]>([]);
const taskProjectStats = ref<TaskProjectStats[]>([]);

// 图表引用
const taskTypePieRef = ref<HTMLDivElement>();
const taskStatusBarRef = ref<HTMLDivElement>();
const taskProjectBarRef = ref<HTMLDivElement>();
const taskTrendLineRef = ref<HTMLDivElement>();

// 图表实例
let taskTypePieChart: echarts.EChartsType | null = null;
let taskStatusBarChart: echarts.EChartsType | null = null;
let taskProjectBarChart: echarts.EChartsType | null = null;
let taskTrendLineChart: echarts.EChartsType | null = null;

// 获取统计数据
async function fetchDashboardStats() {
  try {
    const response = await dashboardApi.getStats();
    if (response.code === 200) {
      dashboardStats.value = response.data as any;
    }
  } catch (error) {
    console.error('获取仪表板统计数据失败:', error);
    // 使用模拟数据
    dashboardStats.value = {
      projectCount: 15,
      webdavCount: 3,
      externalStorageCount: 5,
      scheduledTaskCount: 28,
      nfsCount: 2,
      sambaCount: 4
    };
  }
}

async function fetchTaskTypeStats() {
  try {
    const response = await dashboardApi.getTaskTypeStats();
    if (response.code === 200) {
      taskTypeStats.value = response.data as any;
    }
  } catch (error) {
    console.error('获取任务类型统计失败:', error);
    // 使用模拟数据
    taskTypeStats.value = [
      { type: 'SHELL', count: 12, label: '脚本任务' },
      { type: 'FILE_BACKUP', count: 8, label: '文件备份' },
      { type: 'FILE_CLEAN', count: 6, label: '文件清理' },
      { type: 'JOB_TASK', count: 2, label: '作业任务' }
    ];
  }
}

async function fetchTaskExecutionStats() {
  try {
    const response = await dashboardApi.getTaskExecutionStats();
    if (response.code === 200) {
      taskExecutionStats.value = response.data as any;
    }
  } catch (error) {
    console.error('获取任务执行统计失败:', error);
    // 使用模拟数据
    taskExecutionStats.value = {
      success: 156,
      failed: 12,
      running: 3,
      total: 171
    };
  }
}

async function fetchTaskTrend() {
  try {
    const response = await dashboardApi.getTaskTrend(30);
    if (response.code === 200) {
      taskTrendData.value = response.data as any;
    }
  } catch (error) {
    console.error('获取任务趋势数据失败:', error);
    // 使用模拟数据
    const mockData: TaskTrendData[] = [];
    const today = new Date();
    for (let i = 29; i >= 0; i--) {
      const date = new Date(today);
      date.setDate(date.getDate() - i);
      mockData.push({
        date: date.toISOString().split('T')[0],
        success: Math.floor(Math.random() * 10) + 5,
        failed: Math.floor(Math.random() * 3),
        total: 0
      });
    }
    mockData.forEach(item => {
      item.total = item.success + item.failed;
    });
    taskTrendData.value = mockData;
  }
}

async function fetchTaskProjectStats() {
  try {
    const response = await dashboardApi.getTaskProjectStats();
    if (response.code === 200) {
      taskProjectStats.value = response.data as any;
      // 为每个项目分配不同的颜色
      const colors = ['#2080f0', '#18a058', '#f0a020', '#d03050', '#722ed1', '#fa8c16', '#13c2c2', '#52c41a'];
      taskProjectStats.value.forEach((item, index) => {
        (item as any).color = colors[index % colors.length];
      });
    }
  } catch (error) {
    console.error('获取项目任务统计失败:', error);
    // 使用模拟数据
    taskProjectStats.value = [
      { projectId: '1', projectName: '系统维护', taskCount: 8 },
      { projectId: '2', projectName: '数据备份', taskCount: 6 },
      { projectId: '3', projectName: '日志清理', taskCount: 4 },
      { projectId: '4', projectName: '监控告警', taskCount: 3 },
      { projectId: '0', projectName: '未分类', taskCount: 7 }
    ];
    // 为模拟数据分配颜色
    const colors = ['#2080f0', '#18a058', '#f0a020', '#d03050', '#722ed1'];
    taskProjectStats.value.forEach((item, index) => {
      (item as any).color = colors[index % colors.length];
    });
  }
}

// 初始化任务类型饼图
function initTaskTypePieChart() {
  if (!taskTypePieRef.value) return;
  
  taskTypePieChart = echarts.init(taskTypePieRef.value);
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '任务类型',
        type: 'pie',
        radius: '50%',
        data: taskTypeStats.value.map(item => ({
          value: item.count,
          name: item.label
        })),
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  };
  taskTypePieChart.setOption(option);
}

// 初始化任务状态饼图
function initTaskStatusBarChart() {
  if (!taskStatusBarRef.value) return;
  
  taskStatusBarChart = echarts.init(taskStatusBarRef.value);
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '任务执行状态',
        type: 'pie',
        radius: '50%',
        data: [
          { 
            value: taskExecutionStats.value.success, 
            name: '成功',
            itemStyle: { color: '#18a058' }
          },
          { 
            value: taskExecutionStats.value.failed, 
            name: '失败',
            itemStyle: { color: '#d03050' }
          },
          { 
            value: taskExecutionStats.value.running, 
            name: '运行中',
            itemStyle: { color: '#2080f0' }
          }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  };
  taskStatusBarChart.setOption(option);
}

// 初始化各项目任务数量柱状图
function initTaskProjectBarChart() {
  if (!taskProjectBarRef.value) return;
  
  taskProjectBarChart = echarts.init(taskProjectBarRef.value);
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: taskProjectStats.value.map(item => item.projectName)
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '任务数量',
        type: 'bar',
        data: taskProjectStats.value.map(item => ({
          value: item.taskCount,
          itemStyle: { color: (item as any).color || '#2080f0' }
        }))
      }
    ]
  };
  taskProjectBarChart.setOption(option);
}

// 初始化任务趋势折线图
function initTaskTrendLineChart() {
  if (!taskTrendLineRef.value) return;
  
  taskTrendLineChart = echarts.init(taskTrendLineRef.value);
  const option = {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['成功', '失败', '总数']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    toolbox: {
      feature: {
        saveAsImage: {}
      }
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: taskTrendData.value.map(item => item.date)
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '成功',
        type: 'line',
        stack: 'Total',
        data: taskTrendData.value.map(item => item.success),
        itemStyle: { color: '#18a058' }
      },
      {
        name: '失败',
        type: 'line',
        stack: 'Total',
        data: taskTrendData.value.map(item => item.failed),
        itemStyle: { color: '#d03050' }
      },
      {
        name: '总数',
        type: 'line',
        data: taskTrendData.value.map(item => item.total),
        itemStyle: { color: '#2080f0' }
      }
    ]
  };
  taskTrendLineChart.setOption(option);
}

// 初始化所有图表
async function initCharts() {
  await nextTick();
  initTaskTypePieChart();
  initTaskStatusBarChart();
  initTaskProjectBarChart();
  initTaskTrendLineChart();
}

// 窗口大小改变时重置图表大小
function handleResize() {
  taskTypePieChart?.resize();
  taskStatusBarChart?.resize();
  taskProjectBarChart?.resize();
  taskTrendLineChart?.resize();
}

onMounted(async () => {
  // 获取所有统计数据
  await Promise.all([
    fetchDashboardStats(),
    fetchTaskTypeStats(),
    fetchTaskExecutionStats(),
    fetchTaskTrend(),
    fetchTaskProjectStats()
  ]);
  
  // 初始化图表
  await initCharts();
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize);
});
</script>

<style scoped>
.page-body {
  overflow-y: auto;
  padding: 16px;
}
</style>
