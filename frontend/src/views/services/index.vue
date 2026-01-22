<template>
  <div class="services-container">
    <el-container>
      <el-aside width="300px" style="padding-right: 20px">
        <el-card class="list-card">
          <template #header>模型服务清单</template>
          <div v-for="cfg in configs" :key="cfg.ID" 
               class="service-item" :class="{ active: selectedConfig?.ID === cfg.ID }"
               @click="handleSelect(cfg)">
            <div class="item-header">
              <span class="item-name">{{ cfg.name }}</span>
              <el-tag size="small" :type="getServiceStatus(cfg.name)">{{ getServiceStatusText(cfg.name) }}</el-tag>
            </div>
            <div class="item-meta">{{ cfg.engine }} | {{ cfg.modelPath || 'External' }}</div>
          </div>
          <el-empty v-if="!configs.length" :image-size="60" description="暂无模型" />
        </el-card>
      </el-aside>

      <el-main style="padding: 0">
        <el-card v-if="selectedConfig">
          <template #header>
            <div class="detail-header">
              <span>模型参数配置: {{ selectedConfig.name }}</span>
              <el-button-group>
                <el-button type="primary" size="small" icon="VideoPlay" @click="controlContainer('start')">启动</el-button>
                <el-button type="danger" size="small" icon="VideoPause" @click="controlContainer('stop')">停止</el-button>
              </el-button-group>
            </div>
          </template>

          <el-form :model="selectedConfig" label-width="140px">
            <el-divider content-position="left">推理核心参数</el-divider>
            <el-form-item label="最大并发限制">
              <el-slider v-model="selectedConfig.maxConcurrency" :min="1" :max="256" show-input />
            </el-form-item>
            <el-form-item label="Token 上下文长度">
              <el-select v-model="selectedConfig.tokenLimit" style="width: 100%">
                <el-option label="4,096 Tokens" :value="4096" />
                <el-option label="8,192 Tokens" :value="8192" />
                <el-option label="16,384 Tokens" :value="16384" />
                <el-option label="32,768 Tokens" :value="32768" />
              </el-select>
            </el-form-item>

            <el-divider content-position="left">资源与性能</el-divider>
            <el-form-item label="显存配额 (%)">
              <el-slider v-model="selectedConfig.gpuMemory" :step="0.1" :min="0.1" :max="1" />
            </el-form-item>

            <div class="form-actions">
              <el-button type="primary" @click="handleSave" :loading="saving">保存配置并应用重启</el-button>
            </div>
          </el-form>
        </el-card>
        <el-card v-else class="empty-card">
          <el-empty description="请选择模型进行管理" />
        </el-card>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '../../api/request'
import { ElMessage } from 'element-plus'
import { useDashboardStore } from '../../store/dashboard'

const store = useDashboardStore()
const configs = ref<any[]>([])
const selectedConfig = ref<any>(null)
const saving = ref(false)

const fetchConfigs = async () => {
  const res = await request.get('/configs/inference')
  configs.value = res.data
}

onMounted(() => {
  fetchConfigs()
  store.fetchStats()
})

const handleSelect = (cfg: any) => selectedConfig.value = JSON.parse(JSON.stringify(cfg))
const getServiceStatus = (n: string) => store.services.find(s => s.name === n)?.status === 'Running' ? 'success' : 'info'
const getServiceStatusText = (n: string) => store.services.find(s => s.name === n)?.status === 'Running' ? '运行中' : '已停止'

const handleSave = async () => {
  saving.value = true
  try {
    await request.post('/configs/inference', selectedConfig.value)
    ElMessage.success('配置已更新')
    await controlContainer('restart')
    fetchConfigs()
  } catch (e) { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

const controlContainer = async (action: string) => {
  if (!selectedConfig.value) return
  try {
    await request.post('/container/control', { name: selectedConfig.value.name, action })
    ElMessage.success('操作指令已发送')
    setTimeout(() => store.fetchStats(), 2000)
  } catch (error) { ElMessage.error('操作失败') }
}
</script>

<style scoped>
.services-container { padding: 20px; height: 100%; }
.list-card { height: calc(100vh - 120px); overflow-y: auto; }
.service-item { padding: 15px; border-bottom: 1px solid #f0f2f5; cursor: pointer; transition: all 0.2s; }
.service-item:hover { background-color: #f5f7fa; }
.service-item.active { background-color: #ecf5ff; border-left: 4px solid #409EFF; }
.item-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.item-name { font-weight: bold; color: #303133; }
.item-meta { font-size: 11px; color: #909399; }
.detail-header { display: flex; justify-content: space-between; align-items: center; }
.form-actions { margin-top: 40px; text-align: center; border-top: 1px solid #f0f2f5; padding-top: 20px; }
</style>