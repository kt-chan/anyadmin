const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const dashboardService = {
  getOverviewData: async (token) => {
    logger.debug('Fetching dashboard overview data from Go backend');
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };

      const response = await apiClient.get('/api/v1/dashboard/stats', config);
      const data = response.data;

      // Map backend data to frontend expected structure
      const system = data.system;
      const services = data.services;
      const logs = data.logs;

      const runningCount = services.filter(s => s.status.toLowerCase() === 'running' || s.status.toLowerCase() === 'healthy').length;
      
      const metrics = {
        runningServices: {
          current: runningCount,
          total: services.length,
          onlineRate: ((runningCount / services.length) * 100).toFixed(0) + '%'
        },
        computeLoad: {
          percentage: (system.npuUsage || system.gpuUsage || system.cpuUsage).toFixed(1) + '%',
          type: system.npuUsage > 0 ? 'NPU' : (system.gpuUsage > 0 ? 'GPU' : 'CPU')
        },
        memoryUsage: {
          percentage: system.npuMemTotal > 0 
            ? ((system.npuMemUsed / system.npuMemTotal) * 100).toFixed(1) + '%'
            : ((system.memoryUsed / system.memoryTotal) * 100).toFixed(1) + '%',
          used: system.npuMemTotal > 0 
            ? (system.npuMemUsed / (1024 * 1024 * 1024)).toFixed(1)
            : (system.memoryUsed / (1024 * 1024 * 1024)).toFixed(1),
          total: system.npuMemTotal > 0
            ? (system.npuMemTotal / (1024 * 1024 * 1024)).toFixed(0)
            : (system.memoryTotal / (1024 * 1024 * 1024)).toFixed(0)
        },
        taskQueue: {
          count: 5, // Mocked
          status: '正常'
        }
      };

      const auditLogs = logs.map(l => ({
        user: l.username,
        action: l.action,
        time: new Date(l.createdAt).toLocaleTimeString(),
        details: l.detail,
        type: l.username === 'system' ? 'system' : 'user'
      }));

      // Mocked backup info and config as they are not in GetDashboardStats
      // In a real app, these might be separate calls or included in stats
      const backupInfo = {
        lastBackup: { time: '2024-05-20 02:00', type: '全量' },
        availablePoints: 12
      };

      const configData = {
        concurrency: 64,
        tokenOptions: [
          { value: 4096, label: '4K', selected: false },
          { value: 8192, label: '8K', selected: true },
          { value: 16384, label: '16K', selected: false }
        ],
        dynamicBatching: true,
        hardwareAcceleration: system.npuUsage > 0 ? 'Ascend NPU' : (system.gpuUsage > 0 ? 'NVIDIA GPU' : 'CPU')
      };

      return {
        metrics,
        services,
        backupInfo,
        config: configData,
        auditLogs
      };
    } catch (error) {
      logger.error('Error fetching dashboard data from backend', error);
      throw error;
    }
  },

  getMetrics: async (token) => {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.get('/api/v1/system/stats', config);
      return response.data;
  },

  saveConfig: async (token, configData) => {
    logger.info('Saving config', configData);
    // Mock saving
    return true;
  }
};

module.exports = dashboardService;