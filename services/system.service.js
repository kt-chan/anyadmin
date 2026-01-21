const { getSystemUsersData, getSystemAuditLogs } = require('../data/mockData');
const logger = require('../utils/logger');

const systemService = {
  getSystemData: async () => {
    logger.debug('Fetching system page data');
    try {
      return {
        users: getSystemUsersData(),
        auditLogs: getSystemAuditLogs()
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
