<template>
  <div class="dashboard-container">
    <!-- 1. 顶部资源统计 (参考 demoadmin.html 布局) -->
    <el-row :gutter="20">
      <el-col :span="6" v-for="item in statsCards" :key="item.title">
        <el-card shadow="hover" class="stat-card">
          <div class="card-header">
            <div class="header-left">
              <el-icon :size="20" :color="item.color"><component :is="item.icon" /></el-icon>
              <span class="card-status" :style="{color: item.badgeColor || '#67C23A'}">{{ item.status }}</span>
            </div>
            <el-button v-if="item.hasDetail" type="primary" link size="small" @click="showGpuDetail = true">详情</el-button>
          </div>
          <div class="card-label">{{ item.title }}</div>
          <div class="card-value">{{ item.value }}</div>
          <div class="card-sub-value" v-if="item.subValue">{{ item.subValue }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 2. 下方列表与审计 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card class="list-card">
          <template #header>
            <div class="card-header-flex">
              <span class="font-bold">服务运行清单</span>
              <el-button type="primary" link @click="$router.push('/services')">管理所有 ></el-button>
            </div>
          </template>
          <el-table :data="store.services" border stripe>
            <el-table-column prop="name" label="服务名称">
              <template #default="scope">
                <div class="service-name-cell">
                  <span>{{ scope.row.name }}</span>
                  <el-tooltip v-if="scope.row.pid" placement="top">
                    <template #content>
                      <div>PID: {{ scope.row.pid }}</div>
                      <div>CPU: {{ scope.row.cpu?.toFixed(2) }}%</div>
                      <div>MEM: {{ formatBytes(scope.row.memory) }}</div>
                    </template>
                    <el-icon color="#409EFF" style="margin-left:4px"><InfoFilled /></el-icon>
                  </el-tooltip>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="角色/类型" width="120">
              <template #default="scope">{{ translateType(scope.row.type) }}</template>
            </el-table-column>
            <el-table-column label="健康状况" width="120">
              <template #default="scope">
                <span :class="scope.row.status === 'Running' ? 'text-green' : 'text-gray'">
                  <i class="fas" :class="scope.row.status === 'Running' ? 'fa-check-circle' : 'fa-circle'"></i>
                  {{ scope.row.status === 'Running' ? ' 运行中' : ' 已停止' }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" align="center">
              <template #default="scope">
                <div class="table-ops">
                  <button class="op-btn" @click="handleAction(scope.row.name, 'restart')"><i class="fas fa-redo"></i></button>
                  <button class="op-btn red" @click="handleAction(scope.row.name, 'stop')"><i class="fas fa-power-off"></i></button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card class="audit-card">
          <template #header>
            <span class="font-bold">最近操作审计</span>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="(log, index) in store.logs"
              :key="index"
              :timestamp="formatLogDate(log.CreatedAt)"
              :type="log.Level === 'Info' ? 'primary' : 'warning'"
            >
              <div class="audit-log-content">
                <span class="log-user">{{ log.username }}</span> : {{ log.action }}
              </div>
              <div style="font-size: 12px; color: #666; margin-top: 4px; background: #f0f2f5; padding: 4px 8px; border-radius: 4px;" v-if="log.detail">
                {{ log.detail }}
              </div>
            </el-timeline-item>
          </el-timeline>
          <el-empty v-if="!store.logs.length" description="暂无记录" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>

    <!-- GPU 详情抽屉 -->
    <el-drawer v-model="showGpuDetail" title="加速硬件明细" size="400px">
      <div v-for="dev in store.system.gpuDevices" :key="'gpu-'+dev.index" class="gpu-detail-item">
        <div class="gpu-header">
          <span>#{{ dev.index }} {{ dev.model }}</span>
          <span class="usage-text">{{ dev.usage.toFixed(1) }}%</span>
        </div>
        <el-progress :percentage="dev.usage" status="warning" :stroke-width="8" />
        <div class="gpu-footer">
          显存: {{ formatBytes(dev.memUsed) }} / {{ formatBytes(dev.memTotal) }} 
          ({{ calcPercentage(dev.memUsed, dev.memTotal) }}%)
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useDashboardStore } from '../../store/dashboard'
import request from '../../api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const store = useDashboardStore()
const showGpuDetail = ref(false)

onMounted(() => {
  store.fetchStats()
  const timer = setInterval(() => store.fetchStats(), 15000)
  return () => clearInterval(timer)
})

const healthyCount = computed(() => store.services.filter(s => s.health === 'Healthy').length)
const hasAccelerator = computed(() => (store.system as any).gpuMemTotal > 0 || (store.system as any).npuMemTotal > 0)

// 提取当前使用的加速器数据（优先 GPU）
const acceleratorUsage = computed(() => Math.max(store.system.gpuUsage, store.system.npuUsage))
const acceleratorMemUsed = computed(() => Math.max(store.system.gpuMemUsed, store.system.npuMemUsed))
const acceleratorMemTotal = computed(() => Math.max(store.system.gpuMemTotal, store.system.npuMemTotal))

const statsCards = computed(() => [
  { 
    title: '运行中服务', 
    value: `${healthyCount.value} / ${store.services.length}`, 
    subValue: '集群组件在线状态', 
    icon: 'Server', color: '#2563eb', status: '100% Online', badgeColor: '#10b981'
  },
  { 
    title: '算力负载', 
    value: `${acceleratorUsage.value.toFixed(1)}%`, 
    subValue: hasAccelerator.value ? (store.system.npuMemTotal > 0 ? 'NPU (Ascend)' : 'NVIDIA GPU') : 'Host CPU', 
    icon: 'Cpu', color: '#f59e0b', status: hasAccelerator.value ? 'Accelerator' : 'System', badgeColor: '#94a3b8'
  },
  { 
    title: hasAccelerator.value ? '显存占用' : '内存占用', 
    value: hasAccelerator.value ? `${calcPercentage(acceleratorMemUsed.value, acceleratorMemTotal.value)}%` : `${calcPercentage(store.system.memoryUsed, store.system.memoryTotal)}%`, 
    subValue: hasAccelerator.value 
      ? `${formatBytes(acceleratorMemUsed.value)} / ${formatBytes(acceleratorMemTotal.value)}` 
      : `${formatBytes(store.system.memoryUsed)} / ${formatBytes(store.system.memoryTotal)}`,
    icon: 'Management', color: '#6366f1', status: hasAccelerator.value ? 'VRAM' : 'RAM', badgeColor: '#94a3b8',
    hasDetail: hasAccelerator.value
  },
  { 
    title: '解析任务队列', 
    value: '12', 
    subValue: '待处理文档序列', 
    icon: 'Finished', color: '#14b8a6', status: 'Normal', badgeColor: '#14b8a6'
  }
])

const translateType = (t: string) => { const m: any = { 'Inference': '推理引擎', 'VectorDB': '向量库', 'Collector': '解析引擎', 'Core': '核心 API' }; return m[t] || t; }

const handleAction = (name: string, action: string) => {
  const actionName = action === 'restart' ? '重启' : '停止'
  ElMessageBox.confirm(`确定执行${actionName}服务 ${name} 吗?`, '提示', { type: 'warning' }).then(async () => {
    try { 
      await request.post('/container/control', { name, action }); 
      ElMessage.success('指令已发送'); 
      setTimeout(() => store.fetchStats(), 2000); 
    } catch (e) { ElMessage.error('失败'); }
  })
}

const formatBytes = (b: number) => { 
  if (!b) return '0 B'; 
  const k = 1024, s = ['B', 'KB', 'MB', 'GB', 'TB'], i = Math.floor(Math.log(b) / Math.log(k)); 
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + s[i]; 
}
const calcPercentage = (u: number, t: number) => t ? Math.round((u / t) * 100) : 0
const formatLogDate = (d: string) => d ? new Date(d).toLocaleTimeString() : ''
</script>

<style scoped>
.dashboard-container { padding: 24px; background-color: #f8fafc; min-height: 100%; }
.stat-card { height: 160px; margin-bottom: 20px; border-radius: 16px; border: 1px solid #f1f5f9; transition: all 0.3s; }
.stat-card:hover { transform: translateY(-2px); box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1); }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.header-left { display: flex; align-items: center; gap: 8px; }
.card-status { font-size: 10px; font-weight: 700; text-transform: uppercase; font-family: monospace; }
.card-label { font-size: 13px; color: #64748b; font-weight: 500; margin-bottom: 4px; }
.card-value { font-size: 24px; font-weight: 700; color: #1e293b; }
.card-sub-value { font-size: 11px; color: #94a3b8; margin-top: 6px; font-family: monospace; }
.card-header-flex { display: flex; justify-content: space-between; align-items: center; }
.service-name-cell { display: flex; align-items: center; gap: 4px; }
.text-green { color: #10b981; font-weight: 600; }
.text-gray { color: #94a3b8; }
.table-ops { display: flex; gap: 12px; }
.op-btn { color: #94a3b8; border: none; background: none; cursor: pointer; transition: color 0.2s; }
.op-btn:hover { color: #2563eb; }
.op-btn.red:hover { color: #ef4444; }
.audit-log-content { font-size: 13px; color: #334155; }
.log-user { font-weight: 700; color: #2563eb; }
.gpu-detail-item { padding: 16px; border: 1px solid #f1f5f9; border-radius: 12px; margin-bottom: 16px; background: #fff; }
.gpu-header { display: flex; justify-content: space-between; margin-bottom: 8px; font-weight: 700; font-size: 14px; }
.usage-text { color: #f59e0b; }
.gpu-footer { margin-top: 10px; font-size: 11px; color: #64748b; font-family: monospace; text-align: right; }
</style>
