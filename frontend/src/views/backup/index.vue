<template>
  <div class="backup-container p-8">
    <header class="mb-8">
      <h2 class="text-2xl font-bold text-slate-800">备份与恢复</h2>
      <p class="text-slate-500 text-sm">数据快照管理、系统重置与应用重刷</p>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left: Backup Actions -->
      <div class="space-y-6">
        <div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200">
          <h3 class="font-bold text-lg mb-4 text-slate-800">创建备份</h3>
          <button class="w-full bg-blue-50 text-blue-700 border border-blue-200 py-4 rounded-xl font-bold mb-3 hover:bg-blue-100 text-left px-4 flex justify-between items-center group transition"
                  @click="handleCreate" :disabled="creating">
            <span><i class="fas fa-save mr-2"></i> 全量备份 (Full)</span>
            <i class="fas fa-arrow-right opacity-0 group-hover:opacity-100 transition"></i>
          </button>
          <button class="w-full bg-slate-50 text-slate-700 border border-slate-200 py-4 rounded-xl font-bold text-left px-4 flex justify-between items-center group hover:bg-slate-100 transition">
            <span><i class="fas fa-clock mr-2"></i> 增量备份 (Incremental)</span>
            <i class="fas fa-plus opacity-0 group-hover:opacity-100 transition"></i>
          </button>
        </div>

        <!-- App Reflash -->
        <div class="bg-red-50 p-6 rounded-2xl border border-red-100">
          <h3 class="font-bold text-red-800 mb-2"><i class="fas fa-sync-alt mr-2"></i> 应用重刷 (App Reflash)</h3>
          <p class="text-xs text-red-600 mb-4 leading-relaxed">
            重新部署所有服务镜像，并基于选定备份点重置数据。用于处理底层环境损坏导致的系统级故障。
          </p>
          <button class="w-full bg-red-600 text-white py-2 rounded-lg text-sm font-bold hover:bg-red-700 shadow-lg shadow-red-200 transition">
            启动重刷流程
          </button>
        </div>
      </div>

      <!-- Right: Backup List -->
      <div class="lg:col-span-2 bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden flex flex-col">
        <div class="p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <h3 class="font-bold text-slate-800">备份历史记录</h3>
          <span class="text-xs text-slate-400 font-mono">Total Storage: {{ totalSize }}</span>
        </div>
        
        <div class="flex-1 min-h-[500px]">
          <el-table :data="backups" border-none class="w-full text-sm">
            <el-table-column label="备份ID / 生成时间">
              <template #default="scope">
                <div class="font-bold text-slate-800">{{ scope.row.Name }}</div>
                <div class="text-[10px] text-slate-400 font-mono">{{ formatDate(scope.row.CreatedAt) }}</div>
              </template>
            </el-table-column>
            <el-table-column label="类型" width="100">
              <template #default="scope">
                <span class="bg-blue-100 text-blue-700 px-2 py-0.5 rounded text-[10px] font-bold">{{ scope.row.Type }}</span>
              </template>
            </el-table-column>
            <el-table-column label="大小" width="120">
              <template #default="scope">
                <span class="text-slate-600 font-mono">{{ formatBytes(scope.row.Size) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="校验" width="100">
              <template #default="scope">
                <span v-if="scope.row.Status === 'Success'" class="text-green-500 font-bold text-xs">
                  <i class="fas fa-check-circle mr-1"></i> Pass
                </span>
                <span v-else class="text-red-500 font-bold text-xs">
                  <i class="fas fa-times-circle mr-1"></i> Fail
                </span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" align="right">
              <template #default="scope">
                <el-button type="primary" size="small" link class="font-bold" @click="handleRestore(scope.row)">恢复</el-button>
                <el-button type="danger" size="small" link class="font-bold">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <div v-if="backups.length === 0" class="flex flex-col items-center justify-center p-20 text-slate-300">
            <i class="fas fa-database text-5xl mb-4 opacity-20"></i>
            <p>暂无备份历史记录</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import request from '../../api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const backups = ref<any[]>([])
const creating = ref(false)

const fetchBackups = async () => {
  const res = await request.get('/backups')
  backups.value = res.data
}

onMounted(fetchBackups)

const handleCreate = async () => {
  creating.value = true
  try {
    await request.post('/backups')
    ElMessage.success('系统快照创建成功')
    fetchBackups()
  } catch (error: any) {
    console.error('备份请求失败详情:', error)
    const msg = error.response?.data?.error || error.message || '未知错误'
    ElMessage.error(`备份失败: ${msg}`)
  } finally {
    creating.value = false
  }
}

const handleRestore = (row: any) => {
  ElMessageBox.confirm(
    `确定要从备份 ${row.Name} 恢复吗？此操作将覆盖现有数据库并重启服务。`,
    '恢复确认',
    {
      confirmButtonText: '立即执行',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await request.post(`/backups/restore/${row.ID}`)
      ElMessage.success('恢复任务已在后台启动')
    } catch (error) {
      ElMessage.error('恢复指令发送失败')
    }
  })
}

const totalSize = computed(() => {
  const bytes = backups.value.reduce((acc, b) => acc + b.Size, 0)
  return formatBytes(bytes)
})

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}
</script>

<style scoped>
.backup-container { font-family: 'Inter', 'Noto Sans SC', sans-serif; }
</style>