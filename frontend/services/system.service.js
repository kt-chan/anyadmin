const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const systemService = {
  getSystemData: async () => {
    logger.debug('Fetching system page data');
    try {
      const [usersRes, auditRes] = await Promise.all([
        apiClient.get('/system/users'),
        apiClient.get('/system/audit-logs')
      ]);
      return {
        users: usersRes.data,
        auditLogs: auditRes.data
      };
    } catch (error) {
      logger.error('Error fetching system data', error);
      throw error;
    }
  },

  createUser: async (username, role) => {
    logger.info(`Creating user: ${username} with role: ${role}`);
    return {
      userId: `user_${Date.now()}`
    };
  },

  appReflash: async () => {
    logger.info('Starting app reflash');
    return {
      estimatedTime: '10分钟'
    };
  }
};

module.exports = systemService;
