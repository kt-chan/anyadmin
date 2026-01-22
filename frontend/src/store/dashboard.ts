import { defineStore } from 'pinia'
import request from '../api/request'

export const useDashboardStore = defineStore('dashboard', {
  state: () => ({
    system: {
      cpuUsage: 0,
      memoryTotal: 0,
      memoryUsed: 0,
      diskTotal: 0,
      diskUsed: 0,
      gpuUsage: 0,
      gpuMemUsed: 0,
      gpuMemTotal: 0,
      gpuDevices: [] as any[],
      npuUsage: 0,
      npuMemUsed: 0,
      npuMemTotal: 0,
      npuDevices: [] as any[]
    },
    services: [] as any[],
    logs: [] as any[],
    loading: false
  }),
  actions: {
    async fetchStats() {
      this.loading = true
      try {
        const res = await request.get('/dashboard/stats')
        this.system = res.data.system
        this.services = res.data.services
        this.logs = res.data.logs
      } catch (error) {
        console.error('Failed to fetch dashboard stats', error)
      } finally {
        this.loading = false
      }
    }
  }
})