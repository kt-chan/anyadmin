const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const backupService = {
  getBackupData: async () => {
    logger.debug('Fetching backup page data');
    try {
      const response = await apiClient.get('/backups');
      return {
        backups: response.data
      };
    } catch (error) {
      logger.error('Error fetching backup data', error);
      throw error;
    }
  },

  createBackup: async (type) => {
    logger.info(`Creating backup of type: ${type}`);
    return {
      backupId: `bk_${Date.now()}`,
      startTime: new Date().toLocaleTimeString()
    };
  },

  restoreFromBackup: async (backupId) => {
    logger.info(`Restoring from backup: ${backupId}`);
    return {
      restoreStartTime: new Date().toLocaleTimeString()
    };
  },

  deleteBackup: async (backupId) => {
    logger.info(`Deleting backup: ${backupId}`);
    return true;
  }
};

module.exports = backupService;
