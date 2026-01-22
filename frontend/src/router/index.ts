import { createRouter, createWebHistory } from 'vue-router'
import Layout from '../layout/index.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/login/index.vue')
    },
    {
      path: '/',
      component: Layout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('../views/dashboard/index.vue')
        },
        {
          path: 'deploy',
          name: 'Deploy',
          component: () => import('../views/deploy/index.vue')
        },
        {
          path: 'import',
          name: 'Import',
          component: () => import('../views/import/index.vue')
        },
        {
          path: 'backup',
          name: 'Backup',
          component: () => import('../views/backup/index.vue')
        },
        {
          path: 'services',
          name: 'Services',
          component: () => import('../views/services/index.vue')
        },
        {
          path: 'system',
          name: 'System',
          component: () => import('../views/system/index.vue')
        },
      ]
    }
  ]
})

export default router
