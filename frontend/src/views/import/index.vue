<template>
  <div class="import-container">
    <header class="page-header">
      <div class="header-left">
        <h2>数据批量导入</h2>
        <p>自动扫描并同步本地或对象存储中的文件至知识库</p>
      </div>
      <div class="header-right">
        <el-tag type="primary" effect="plain" class="sync-tag">
          <span class="dot blue"></span> 同步引擎就绪
        </el-tag>
      </div>
    </header>

    <el-row :gutter="20">
      <!-- 左侧: 配置面板 -->
      <el-col :span="8">
        <el-card class="config-card">
          <template #header>
            <div class="card-header">
              <el-icon><Setting /></el-icon>
              <span>数据源配置</span>
            </div>
          </template>
          <el-form :model="form" label-position="top">
            <el-form-item label="导入模式">
              <el-radio-group v-model="form.SourceType" size="small" style="width: 100%">
                <el-radio-button label="Local" style="width: 50%">文件路径</el-radio-button>
                <el-radio-button label="S3" style="width: 50%">对象存储</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="任务名称">
              <el-input v-model="form.Name" placeholder="技术部文档同步" />
            </el-form-item>
            <el-form-item label="绝对路径 / Bucket 路径">
              <el-input v-model="form.SourcePath" placeholder="/data/knowledge_base/docs/" />
            </el-form-item>
            <el-form-item>
              <div class="checksum-box">
                <span>启用 Checksum 校验</span>
                <el-switch v-model="form.checksum" />
              </div>
              <p class="hint-text">开启后将对比文件名、大小及最后更新时间。若哈希一致则跳过。</p>
            </el-form-item>
            <el-button type="primary" class="start-btn" :loading="creating" @click="handleCreate">
              开始扫描变更
            </el-button>
          </el-form>
        </el-card>

        <el-card class="stat-mini-card" style="margin-top: 20px">
          <template #header>导入统计汇总</template>
          <el-row :gutter="10">
            <el-col :span="12">
              <div class="mini-stat-box">
                <div class="label">活跃任务</div>
                <div class="value">{{ runningCount }}</div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="mini-stat-box green">
                <div class="label">总计处理</div>
                <div class="value">{{ totalProcessed }}</div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>

      <!-- 右侧: 任务流 -->
      <el-col :span="16">
        <el-card class="task-card">
          <template #header>
            <div class="card-header-flex">
              <div>
                <span class="title">当前同步任务流</span>
                <span class="subtitle">实时识别并处理文件分片</span>
              </div>
              <div class="actions">
                <el-button size="small">暂停全部</el-button>
                <el-button size="small" type="danger" plain>清空列表</el-button>
              </div>
            </div>
          </template>
          
          <el-table :data="tasks" border stripe height="500">
            <el-table-column label="任务名称/路径">
              <template #default="scope">
                <div class="task-info">
                  <el-icon :size="20" color="#409EFF"><FolderChecked /></el-icon>
                  <div class="text">
                    <p class="name">{{ scope.row.Name }}</p>
                    <p class="path">{{ scope.row.SourcePath }}</p>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="120">
              <template #default="scope">
                <el-tag :type="getStatusType(scope.row.Status)" size="small">
                  {{ translateStatus(scope.row.Status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="进度" width="200">
              <template #default="scope">
                <div class="progress-box">
                  <div class="p-labels">
                    <span>{{ scope.row.Progress }}%</span>
                    <span>{{ scope.row.Processed }} / {{ scope.row.TotalFiles }}</span>
                  </div>
                  <el-progress :percentage="scope.row.Progress" :show-text="false" :stroke-width="6" 
                               :status="scope.row.Status === 'Completed' ? 'success' : ''" />
                </div>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" align="center">
              <template #default>
                <el-button size="small" link icon="VideoPause" />
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="tasks.length === 0" description="暂无导入任务" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import request from '../../api/request'
import { ElMessage } from 'element-plus'

const tasks = ref<any[]>([])
const creating = ref(false)
const form = ref({
  Name: '',
  SourceType: 'Local',
  SourcePath: '',
  checksum: true
})

const fetchTasks = async () => {
  try {
    const res = await request.get('/import/tasks')
    tasks.value = res.data
  } catch (e) { console.error(e) }
}

onMounted(() => {
  fetchTasks()
  const timer = setInterval(fetchTasks, 3000)
  return () => clearInterval(timer)
})

const handleCreate = async () => {
  if (!form.value.Name || !form.value.SourcePath) {
    ElMessage.warning('请完整填写配置')
    return
  }
  creating.value = true
  try {
    await request.post('/import/tasks', form.value)
    ElMessage.success('同步任务已启动')
    fetchTasks()
  } catch (error: any) {
    ElMessage.error('启动失败')
  } finally {
    creating.value = false
  }
}

const runningCount = computed(() => tasks.value.filter(t => t.Status === 'Running').length)
const totalProcessed = computed(() => tasks.value.reduce((acc, t) => acc + (t.Processed || 0), 0))

const translateStatus = (s: string) => {
  const map: any = { 'Running': '正在同步', 'Completed': '同步完成', 'Failed': '失败', 'Paused': '已暂停' }
  return map[s] || s
}

const getStatusType = (s: string) => {
  if (s === 'Running') return ''
  if (s === 'Completed') return 'success'
  if (s === 'Failed') return 'danger'
  return 'info'
}
</script>

<style scoped>
.import-container { padding: 20px; background-color: #f5f7fa; min-height: 100%; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-left h2 { margin: 0; font-size: 24px; color: #303133; }
.header-left p { margin: 4px 0 0; color: #909399; font-size: 14px; }
.sync-tag { padding: 8px 15px; font-weight: bold; }
.card-header { display: flex; align-items: center; gap: 8px; font-weight: bold; }
.checksum-box { display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px; }
.hint-text { font-size: 11px; color: #909399; font-style: italic; line-height: 1.4; }
.start-btn { width: 100%; margin-top: 15px; height: 45px; font-weight: bold; }
.mini-stat-box { background: #1e222d; color: #fff; padding: 15px; border-radius: 12px; text-align: center; }
.mini-stat-box.green .value { color: #67C23A; }
.mini-stat-box .label { font-size: 10px; text-transform: uppercase; color: #909399; margin-bottom: 5px; }
.mini-stat-box .value { font-size: 20px; font-weight: bold; }
.card-header-flex { display: flex; justify-content: space-between; align-items: center; }
.card-header-flex .title { font-weight: bold; display: block; }
.card-header-flex .subtitle { font-size: 11px; color: #909399; }
.task-info { display: flex; align-items: center; gap: 12px; }
.task-info .name { font-weight: bold; margin: 0; }
.task-info .path { font-size: 11px; color: #909399; margin: 2px 0 0; }
.progress-box .p-labels { display: flex; justify-content: space-between; font-size: 11px; font-weight: bold; margin-bottom: 5px; }
.dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; margin-right: 5px; }
.dot.blue { background-color: #409EFF; box-shadow: 0 0 5px #409EFF; }
</style>
