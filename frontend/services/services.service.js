const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const servicesService = {
  getServicesList: async (token) => {
    logger.debug('Fetching services list');
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      // Use dashboard stats as it contains the services health info
      const response = await apiClient.get('/api/v1/dashboard/stats', config);
      return {
        services: response.data.services
      };
    } catch (error) {
      logger.error('Error fetching services data', error);
      throw error;
    }
  },

  getServicesStatus: async (token) => {
    const config = {
      headers: { Authorization: `Bearer ${token}` }
    };
    const response = await apiClient.get('/api/v1/dashboard/stats', config);
    return response.data.services;
  },

  restartService: async (serviceName, nodeIP, token, serviceType) => {
    logger.info(`Restarting service: ${serviceName} on node: ${nodeIP} (Type: ${serviceType})`);
    const config = {
      headers: { Authorization: `Bearer ${token}` }
    };
    
    if (serviceType === 'Agent') {
      await apiClient.post('/api/v1/deploy/agent/control', { ip: nodeIP, action: 'restart' }, config);
    } else {
      await apiClient.post('/api/v1/container/control', { name: serviceName, action: 'restart', node_ip: nodeIP }, config);
    }
    return true;
  },

  stopService: async (serviceName, nodeIP, token, serviceType) => {
    logger.info(`Stopping service: ${serviceName} on node: ${nodeIP} (Type: ${serviceType})`);
    const config = {
      headers: { Authorization: `Bearer ${token}` }
    };

    if (serviceType === 'Agent') {
      await apiClient.post('/api/v1/deploy/agent/control', { ip: nodeIP, action: 'stop' }, config);
    } else {
      await apiClient.post('/api/v1/container/control', { name: serviceName, action: 'stop', node_ip: nodeIP }, config);
    }
    return true;
  }
};

module.exports = servicesService;