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

      // Mocked backup info
      const backupInfo = {
        lastBackup: { time: '2024-05-20 02:00', type: '全量' },
        availablePoints: 12
      };

      // Fetch inference configs
      let inferenceConfig = {
        mode: 'balanced',
        concurrency: 64,
        tokenLimit: 8192
      };
      try {
        const configResp = await apiClient.get('/api/v1/configs/inference', config);
        if (configResp.data && configResp.data.length > 0) {
          // Use the first config or specific one if we had logic to select
          const backendCfg = configResp.data[0];
          inferenceConfig = {
            mode: backendCfg.mode || 'balanced',
            concurrency: backendCfg.max_num_seqs || 64,
            tokenLimit: backendCfg.max_model_len || 8192
          };
        }
      } catch (err) {
        logger.warn('Failed to fetch inference config, using default', err.message);
      }

      const configData = {
        mode: inferenceConfig.mode,
        concurrency: inferenceConfig.concurrency,
        tokenOptions: [
          { value: 4096, label: '4K', selected: inferenceConfig.tokenLimit === 4096 },
          { value: 8192, label: '8K', selected: inferenceConfig.tokenLimit === 8192 },
          { value: 16384, label: '16K', selected: inferenceConfig.tokenLimit === 16384 },
          { value: 32768, label: '32K', selected: inferenceConfig.tokenLimit === 32768 }
        ],
        dynamicBatching: true
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
    logger.info('Saving config to backend', configData);
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };

      // Map the simple frontend data to the backend structure if necessary,
      // or just send what the backend api.SaveInferenceConfig expects (global.InferenceConfig).
      // Based on frontend/public/js/dashboard.js, it sends { name, mode }.
      
      const response = await apiClient.post('/api/v1/configs/inference', configData, config);
      return response.data;
    } catch (error) {
      logger.error('Error saving config to backend', error);
      throw error;
    }
  },

  calculateVllmConfig: async (token, data) => {
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/configs/vllm-calculate', data, config);
      return response.data;
    } catch (error) {
      logger.error('Error calculating vLLM config', error);
      throw error;
    }
  }
};

module.exports = dashboardService;