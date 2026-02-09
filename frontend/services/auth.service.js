const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const authService = {
  authenticate: async (username, password) => {
    const loginUrl = `${apiClient.defaults.baseURL}/api/v1/login`;
    logger.debug(`Attempting login for user: ${username} at URL: ${loginUrl}`);
    logger.debug(`Password sent to backend service (first 20 chars): ${password.substring(0, 20)}...`);
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
  },

  getPublicKey: async () => {
    try {
      const response = await apiClient.get('/api/v1/public-key');
      if (response.data && response.data.publicKey) {
        return response.data.publicKey;
      }
    } catch (error) {
      logger.warn(`Failed to fetch public key: ${error.message}`);
    }
    return null;
  }
};

module.exports = authService;
