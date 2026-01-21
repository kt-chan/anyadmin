const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const servicesService = {
  getServicesList: async () => {
    logger.debug('Fetching services list');
    try {
      const response = await apiClient.get('/services');
      return {
        services: response.data
      };
    } catch (error) {
      logger.error('Error fetching services data', error);
      throw error;
    }
  },

  getServicesStatus: async () => {
    const response = await apiClient.get('/dashboard/services');
    return response.data;
  },

  restartService: async (serviceId) => {
    logger.info(`Restarting service: ${serviceId}`);
    return true;
  },

  stopService: async (serviceId) => {
    logger.info(`Stopping service: ${serviceId}`);
    return true;
  }
};

module.exports = servicesService;
