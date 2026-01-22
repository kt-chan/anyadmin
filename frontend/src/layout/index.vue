<template>
  <el-container class="layout-container">
    <el-aside width="240px" class="aside">
      <div class="logo-box">
        <i class="fas fa-brain"></i>
        <span>anyzearch 管理</span>
      </div>
      <el-menu
        router
        :default-active="$route.path"
        background-color="#1e222d"
        text-color="#b1b3b8"
        active-text-color="#ffffff"
        class="menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>监控仪表板</span>
        </el-menu-item>
        <el-sub-menu index="components">
          <template #title>
            <el-icon><Operation /></el-icon>
            <span>组件管理</span>
          </template>
          <el-menu-item index="/deploy">部署配置向导</el-menu-item>
          <el-menu-item index="/services">服务与模型管理</el-menu-item>
          <el-menu-item index="/import">数据批量导入</el-menu-item>
        </el-sub-menu>
        <el-menu-item index="/backup">
          <el-icon><DataLine /></el-icon>
          <span>备份与恢复</span>
        </el-menu-item>
        <el-menu-item index="/system">
          <el-icon><Setting /></el-icon>
          <span>系统管理 (用户)</span>
        </el-menu-item>
      </el-menu>
      
      <div class="aside-footer">
        <el-button link type="danger" @click="handleLogout">
          <el-icon><SwitchButton /></el-icon> 退出登录
        </el-button>
      </div>
    </el-aside>

    <el-container class="main-container">
      <el-header class="header">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item>{{ translateName($route.name) }}</el-breadcrumb-item>
        </el-breadcrumb>
        <div class="user-info">
          <span class="user-name">管理员</span>
          <el-avatar :size="32" src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" />
        </div>
      </el-header>
      
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'

const router = useRouter()
const translateName = (n: any) => {
  const m: any = { 'Dashboard': '仪表板', 'Deploy': '部署向导', 'Services': '服务管理', 'Import': '数据导入', 'Backup': '备份恢复', 'System': '系统管理' }
  return m[n] || n
}

const handleLogout = () => {
  ElMessageBox.confirm('确定退出吗？', '提示', { type: 'warning' }).then(() => {
    localStorage.removeItem('token')
    router.push('/login')
  })
}
</script>

<style scoped>
.layout-container { height: 100vh; overflow: hidden; }
.aside { background-color: #1e222d; display: flex; flex-direction: column; height: 100%; }
.logo-box { height: 60px; display: flex; align-items: center; padding: 0 20px; gap: 10px; color: #fff; font-size: 18px; font-weight: bold; background: #161a23; flex-shrink: 0; }
.menu { border-right: none; flex: 1; overflow-y: auto; }
.aside-footer { padding: 20px; border-top: 1px solid #2d323f; flex-shrink: 0; }
.header { background: #fff; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #f0f2f5; height: 60px; flex-shrink: 0; }
.user-info { display: flex; align-items: center; gap: 10px; }
.main-container { height: 100vh; display: flex; flex-direction: column; }
.main-content { background-color: #f5f7fa; padding: 20px; overflow-y: auto; flex: 1; }
</style>