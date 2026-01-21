const {
  getDashboardMetrics,
  getDashboardServices,
  getBackupInfo,
  getDashboardConfig,
  getDashboardAuditLogs
} = require('../data/mockData');
const logger = require('../utils/logger');

const dashboardService = {
  getOverviewData: async () => {
    logger.debug('Fetching dashboard overview data');
    try {
      // In a real app, these might be parallel DB calls
      const metrics = getDashboardMetrics();
      const services = getDashboardServices();
      const backupInfo = getBackupInfo();
      const config = getDashboardConfig();
      const auditLogs = getDashboardAuditLogs();

      return {
        metrics,
        services,
        backupInfo,
        config,
        auditLogs
      };
    } catch (error) {
      logger.error('Error fetching dashboard data', error);
      throw error;
    }
  },

  getMetrics: async () => {
      return getDashboardMetrics();
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
