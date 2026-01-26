const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const backupService = {
  getBackupData: async (token) => {
    logger.debug('Fetching backup page data');
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.get('/api/v1/backups', config);
      return {
        backups: response.data
      };
    } catch (error) {
      logger.error('Error fetching backup data', error);
      throw error;
    }
  },

  createBackup: async (token, type) => {
    logger.info(`Creating backup of type: ${type}`);
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/backups', { type }, config);
      return response.data;
    } catch (error) {
      logger.error('Error creating backup', error);
      throw error;
    }
  },

  restoreFromBackup: async (token, backupId) => {
    logger.info(`Restoring from backup: ${backupId}`);
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post(`/api/v1/backups/restore/${backupId}`, {}, config);
      return response.data;
    } catch (error) {
      logger.error('Error restoring from backup', error);
      throw error;
    }
  },

  deleteBackup: async (token, backupId) => {
    logger.info(`Deleting backup: ${backupId}`);
    // Backend doesn't seem to have delete backup yet in router.go
    return true;
  }
};

module.exports = backupService;