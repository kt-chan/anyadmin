const { getServicesData, getDashboardServices } = require('../data/mockData');
const logger = require('../utils/logger');

const servicesService = {
  getServicesList: async () => {
    logger.debug('Fetching services list');
    try {
      return {
        services: getServicesData()
      };
    } catch (error) {
      logger.error('Error fetching services data', error);
      throw error;
    }
  },

  getServicesStatus: async () => {
    return getDashboardServices();
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
