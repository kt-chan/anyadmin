const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const systemService = {
  getSystemData: async (token) => {
    logger.debug('Fetching system page data');
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const [usersRes, statsRes] = await Promise.all([
        apiClient.get('/api/v1/users', config),
        apiClient.get('/api/v1/dashboard/stats', config)
      ]);

      const logs = statsRes.data.logs.map(l => ({
        user: l.username,
        action: l.action,
        time: new Date(l.createdAt).toLocaleString(),
        details: l.detail,
        type: l.username === 'system' ? 'system' : 'user'
      }));

      return {
        users: usersRes.data,
        auditLogs: logs
      };
    } catch (error) {
      logger.error('Error fetching system data', error);
      throw error;
    }
  },

  createUser: async (token, username, role, password) => {
    logger.info(`Creating user: ${username} with role: ${role}`);
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/users', { username, role, password }, config);
      return response.data;
    } catch (error) {
      logger.error('Error creating user', error);
      throw error;
    }
  },

  appReflash: async (token) => {
    logger.info('Starting app reflash');
    return {
      estimatedTime: '10分钟'
    };
  }
};

module.exports = systemService;