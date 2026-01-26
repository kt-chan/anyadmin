const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const authService = {
  authenticate: async (username, password) => {
    logger.debug(`Attempting login for user: ${username}`);
    try {
      const response = await apiClient.post('/api/v1/login', { username, password });
      const data = response.data;
      if (data && data.user) {
        const user = { ...data.user, token: data.token };
        logger.info(`User logged in successfully: ${username}`);
        return user;
      }
    } catch (error) {
       logger.warn(`Login failed for user: ${username} - ${error.response?.status} ${error.response?.statusText}`);
    }
    return null;
  }
};

module.exports = authService;
