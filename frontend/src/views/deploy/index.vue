<template>
  <div class="deploy-container">
    <header class="page-header">
      <h2>部署配置向导</h2>
      <p>支持全新环境部署或对接现有中间件服务</p>
    </header>

    <el-card class="wizard-card">
      <div class="wizard-steps-container">
        <el-steps :active="activeStep" finish-status="success" align-center>
          <el-step title="模式选择" />
          <el-step title="硬件配置" />
          <el-step title="组件参数" />
          <el-step title="生成部署" />
        </el-steps>
      </div>

      <div class="step-content">
        <!-- 步骤 1: 模式选择 -->
        <div v-if="activeStep === 0" class="mode-selection">
          <h3>选择您的部署模式</h3>
          <el-row :gutter="20">
            <el-col :span="12">
              <div class="mode-box" :class="{ active: form.deploy_mode === 'new' }" @click="form.deploy_mode = 'new'">
                <div class="icon-circle blue"><el-icon :size="24"><Box /></el-icon></div>
                <h4>全新部署</h4>
                <p>自动拉取并配置所有核心组件（推理、向量库、解析）。适用于空环境初始化。</p>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="mode-box" :class="{ active: form.deploy_mode === 'existing' }" @click="form.deploy_mode = 'existing'">
                <div class="icon-circle green"><el-icon :size="24"><Link /></el-icon></div>
                <h4>对接现有服务</h4>
                <p>手动输入现有服务的 IP 和端口，仅部署管理面板。适用于已有基础设施。</p>
              </div>
            </el-col>
          </el-row>
        </div>

        <!-- 步骤 2: 硬件配置 -->
        <div v-if="activeStep === 1" class="hardware-selection">
          <h3>目标硬件环境</h3>
          <div class="hardware-btns">
            <div class="hw-btn" :class="{ active: form.hardware === 'Ascend' }" @click="form.hardware = 'Ascend'">
              <i class="fas fa-microchip"></i>
              <span>华为昇腾 (Ascend NPU)</span>
              <el-icon v-if="form.hardware === 'Ascend'"><Check /></el-icon>
            </div>
            <div class="hw-btn" :class="{ active: form.hardware === 'NVIDIA' }" @click="form.hardware = 'NVIDIA'">
              <i class="fab fa-nvidia"></i>
              <span>NVIDIA GPU</span>
              <el-icon v-if="form.hardware === 'NVIDIA'"><Check /></el-icon>
            </div>
          </div>
        </div>

        <!-- 步骤 3: 组件参数 -->
        <div v-if="activeStep === 2">
          <h3>设置组件参数</h3>
          <el-form :model="form" label-position="top">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="服务名称">
                  <el-input v-model="form.name" placeholder="例如: qwen-7b" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="推理引擎">
                  <el-select v-model="form.engine" style="width: 100%">
                    <el-option v-if="form.hardware === 'NVIDIA'" label="vLLM" value="vLLM" />
                    <el-option v-if="form.hardware === 'Ascend'" label="MindIE" value="MindIE" />
                    <el-option label="Ollama" value="Ollama" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>

            <!-- 对接现有模式专用 -->
            <template v-if="form.deploy_mode === 'existing'">
              <el-row :gutter="20">
                <el-col :span="16">
                  <el-form-item label="服务 IP 地址">
                    <el-input v-model="form.ip" placeholder="192.168.1.100" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="服务端口">
                    <el-input v-model="form.port" placeholder="11434" />
                  </el-form-item>
                </el-col>
              </el-row>
            </template>

            <!-- 全新部署模式专用 -->
            <template v-else>
              <el-form-item label="模型路径">
                <el-input v-model="form.modelPath" placeholder="/data/models/qwen" />
              </el-form-item>
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-form-item label="映射端口 (Host)">
                    <el-input v-model="form.port" placeholder="8000" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="并发限制">
                    <el-input-number v-model="form.maxConcurrency" :min="1" :max="512" style="width: 100%" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="显存配额">
                    <el-input-number v-model="form.gpuMemory" :step="0.1" :min="0.1" :max="1" style="width: 100%" />
                  </el-form-item>
                </el-col>
              </el-row>
            </template>
          </el-form>
        </div>

        <!-- 步骤 4: 预览部署 -->
        <div v-if="activeStep === 3">
          <h3>确认配置并执行</h3>
          <div class="preview-box">
            <div class="preview-item"><span>部署模式:</span> <strong>{{ form.deploy_mode === 'new' ? '全新部署' : '对接现有' }}</strong></div>
            <div class="preview-item"><span>服务名称:</span> <strong>{{ form.name }}</strong></div>
            <div v-if="form.deploy_mode === 'existing'" class="preview-item">
              <span>连接地址:</span> <strong>{{ form.ip }}:{{ form.port }}</strong>
            </div>
            <div v-else>
              <div class="preview-item"><span>模型路径:</span> <strong>{{ form.modelPath }}</strong></div>
              <div class="preview-item"><span>容器映射端口:</span> <strong>{{ form.port }}</strong></div>
            </div>
          </div>
          <div class="action-footer">
            <el-button type="primary" size="large" class="deploy-btn" :loading="loading" @click="handleDeploy">
              {{ form.deploy_mode === 'new' ? '开始执行部署流程' : '保存连接配置' }}
            </el-button>
          </div>
        </div>
      </div>

      <div class="wizard-footer">
        <el-button v-if="activeStep > 0" @click="activeStep--">上一步</el-button>
        <el-button v-if="activeStep < 3" type="primary" @click="activeStep++">下一步</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import request from '../../api/request'

const activeStep = ref(0)
const loading = ref(false)
const form = reactive({
  deploy_mode: 'new',
  hardware: 'NVIDIA',
  engine: 'vLLM',
  name: '',
  modelPath: '',
  ip: '',
  port: '8000',
  maxConcurrency: 64,
  tokenLimit: 4096,
  gpuMemory: 0.9
})

const handleDeploy = async () => {
  if (!form.name) return ElMessage.error('名称不能为空')
  loading.value = true
  try {
    await request.post('/deploy/generate', form)
    ElNotification({ title: '成功', message: '操作已完成', type: 'success' })
  } catch (error: any) {
    const errorMsg = error.response?.data?.error || '执行失败'
    ElMessageBox.alert(errorMsg, '错误提示')
  } finally { loading.value = false }
}
</script>

<style scoped>
.deploy-container { padding: 20px; background-color: #f5f7fa; min-height: 100%; }
.page-header h2 { margin: 0; font-size: 24px; color: #303133; }
.page-header p { margin: 4px 0 0; color: #909399; font-size: 14px; }
.wizard-card { border-radius: 16px; border: none; }
.wizard-steps-container { background-color: #f8fafc; padding: 20px; border-bottom: 1px solid #f0f2f5; margin: -20px -20px 20px -20px; }
.step-content { min-height: 400px; padding: 20px; }
.mode-box { border: 2px solid #f0f2f5; padding: 30px; border-radius: 16px; text-align: center; cursor: pointer; transition: all 0.3s; }
.mode-box.active { border-color: #409EFF; background-color: #ecf5ff; }
.icon-circle { width: 50px; height: 50px; border-radius: 12px; display: flex; align-items: center; justify-content: center; margin: 0 auto 15px; }
.icon-circle.blue { background: #ecf5ff; color: #409EFF; }
.icon-circle.green { background: #f0f9eb; color: #67C23A; }
.hardware-btns { display: flex; gap: 20px; margin-top: 20px; }
.hw-btn { border: 1px solid #dcdfe6; padding: 15px 30px; border-radius: 12px; cursor: pointer; display: flex; align-items: center; gap: 10px; font-weight: bold; }
.hw-btn.active { border-color: #409EFF; color: #409EFF; background: #ecf5ff; }
.preview-box { background: #f8fafc; padding: 20px; border-radius: 12px; border: 1px solid #f0f2f5; margin-bottom: 30px; }
.preview-item { display: flex; justify-content: space-between; margin-bottom: 10px; font-size: 14px; }
.deploy-btn { width: 100%; height: 55px; border-radius: 12px; }
.wizard-footer { display: flex; justify-content: space-between; border-top: 1px solid #f0f2f5; padding-top: 20px; }
</style>