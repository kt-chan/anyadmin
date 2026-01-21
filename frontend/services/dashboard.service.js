const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const dashboardService = {
  getOverviewData: async () => {
    logger.debug('Fetching dashboard overview data');
    try {
      const [metricsRes, servicesRes, backupRes, configRes, auditRes] = await Promise.all([
        apiClient.get('/dashboard/metrics'),
        apiClient.get('/dashboard/services'),
        apiClient.get('/backup/info'),
        apiClient.get('/dashboard/config'),
        apiClient.get('/dashboard/audit-logs')
      ]);

      return {
        metrics: metricsRes.data,
        services: servicesRes.data,
        backupInfo: backupRes.data,
        config: configRes.data,
        auditLogs: auditRes.data
      };
    } catch (error) {
      logger.error('Error fetching dashboard data', error);
      throw error;
    }
  },

  getMetrics: async () => {
      const response = await apiClient.get('/dashboard/metrics');
      return response.data;
  },

  saveConfig: async (configData) => {
    logger.info('Saving config', configData);
    // Mock saving
    return true;
  },

  updateConcurrency: async (value) => {
    logger.info(`Updating concurrency to ${value}`);
    return value;
  },

  updateTokenLimit: async (value) => {
    logger.info(`Updating token limit to ${value}`);
    return value;
  }
};

module.exports = dashboardService;
