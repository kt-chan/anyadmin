<template>
  <div class="system-container p-8">
    <header class="mb-8">
      <h2 class="text-2xl font-bold text-slate-800">系统管理</h2>
      <p class="text-slate-500 text-sm">用户角色、访问权限控制与全局审计日志</p>
    </header>

    <!-- User List Section -->
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden mb-8 transition hover:shadow-md">
      <div class="p-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
        <div>
          <h3 class="font-bold text-slate-800">系统用户列表</h3>
          <p class="text-[10px] text-slate-400 font-medium">管理具备后台访问权限的账户</p>
        </div>
        <el-button type="primary" size="small" class="bg-blue-600 rounded-lg font-bold" @click="dialogVisible = true">
          <i class="fas fa-user-plus mr-1"></i> 新建用户
        </el-button>
      </div>
      
      <el-table :data="users" border-none class="w-full text-sm">
        <el-table-column label="用户名">
          <template #default="scope">
            <span class="font-bold text-slate-700">{{ scope.row.username }}</span>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="150">
          <template #default="scope">
            <span :class="scope.row.role === 'admin' ? 'bg-purple-100 text-purple-700' : 'bg-slate-100 text-slate-600'" 
                  class="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider">
              {{ scope.row.role === 'admin' ? 'Administrator' : 'Operator' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default>
            <span class="text-green-500 text-xs font-bold flex items-center gap-1.5">
              <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span> Active
            </span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="200">
          <template #default="scope">
            <span class="text-slate-500 text-xs font-mono">{{ new Date(scope.row.CreatedAt).toLocaleString() }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="right">
          <template #default="scope">
            <el-button v-if="scope.row.username !== 'admin'" type="danger" size="small" link class="font-bold" @click="handleDelete(scope.row.ID)">删除</el-button>
            <span v-else class="text-slate-300 text-xs italic"><i class="fas fa-lock"></i> System</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Audit Log Section -->
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6 transition hover:shadow-md">
      <div class="flex justify-between items-center mb-6">
        <h3 class="font-bold text-slate-800">全局操作审计日志</h3>
        <el-button size="small" link>查看完整日志 <i class="fas fa-external-link-alt ml-1"></i></el-button>
      </div>
      
      <div class="space-y-4">
        <div v-for="(log, index) in auditLogs" :key="index" class="flex gap-4 items-start text-sm pb-4 border-b border-slate-50 last:border-0">
          <span class="font-mono text-slate-400 text-xs mt-1">{{ formatDate(log.CreatedAt) }}</span>
          <div class="flex-1">
            <p class="font-bold text-slate-800">
              用户 <span class="text-blue-600">{{ log.username }}</span> 
              <span class="text-slate-700 mx-1">{{ log.action }}</span>
            </p>
            <p class="text-xs text-slate-500 mt-1">Detail: {{ log.detail || 'No additional data' }}</p>
          </div>
          <span class="text-[10px] font-bold text-slate-300 uppercase tracking-widest">JWT AUTH</span>
        </div>
        
        <div v-if="auditLogs.length === 0" class="text-center py-10 text-slate-300">
          <i class="fas fa-scroll text-3xl mb-2 opacity-20"></i>
          <p class="text-xs">当前无操作日志记录</p>
        </div>
      </div>
    </div>

    <!-- Create User Dialog -->
    <el-dialog v-model="dialogVisible" title="创建系统账户" width="450px" custom-class="system-dialog">
      <el-form :model="form" label-width="80px" label-position="top">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入登录名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="分配角色">
          <el-select v-model="form.role" style="width: 100%">
            <el-option label="管理员 (完整权限)" value="admin" />
            <el-option label="操作员 (仅限查看/基本控制)" value="operator" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="dialogVisible = false" class="font-bold">取消</el-button>
          <el-button type="primary" class="bg-blue-600 font-bold" @click="handleCreate">确认创建</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '../../api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const users = ref([])
const auditLogs = ref<any[]>([])
const dialogVisible = ref(false)
const form = ref({ username: '', password: '', role: 'operator' })

const fetchData = async () => {
  try {
    const [userRes, dashboardRes] = await Promise.all([
      request.get('/users'),
      request.get('/dashboard/stats')
    ])
    users.value = userRes.data
    auditLogs.value = dashboardRes.data.logs
  } catch (error) {
    console.error('Data fetch failed', error)
  }
}

onMounted(fetchData)

const handleCreate = async () => {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请填写完整账户信息')
    return
  }
  try {
    await request.post('/users', form.value)
    ElMessage.success('系统账户创建成功')
    dialogVisible.value = false
    form.value = { username: '', password: '', role: 'operator' }
    fetchData()
  } catch (error: any) {
    ElMessage.error('账户创建失败')
  }
}

const handleDelete = (id: number) => {
  ElMessageBox.confirm('确定要永久删除此系统账户吗?', '风险警告', { 
    confirmButtonText: '确定删除',
    cancelButtonText: '保留',
    type: 'warning' 
  }).then(async () => {
    try {
      await request.delete(`/users/${id}`)
      ElMessage.success('用户已成功移除')
      fetchData()
    } catch (error) {
      ElMessage.error('移除操作失败')
    }
  })
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleTimeString()
}
</script>

<style scoped>
.system-container { font-family: 'Inter', 'Noto Sans SC', sans-serif; }
:deep(.el-table) { --el-table-border-color: transparent; }
:deep(.el-table th.el-table__cell) { background-color: #f8fafc; color: #64748b; font-size: 10px; text-transform: uppercase; font-weight: 700; }
</style>