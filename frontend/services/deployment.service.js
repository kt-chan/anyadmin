const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const deploymentService = {
  // Generate deployment artifacts
  generateDeployment: async (config) => {
    try {
      const response = await apiClient.post('/deploy/generate', config);
      return response.data;
    } catch (error) {
      logger.error('Error generating deployment:', error);
      throw error;
    }
  },

  // Test service connection
  testConnection: async (serviceDetails) => {
    try {
      const response = await apiClient.post('/deploy/test-connection', serviceDetails);
      return response.data;
    } catch (error) {
      logger.error('Error testing connection:', error);
      throw error;
    }
  },

  // Get list of models
  getModels: async () => {
    try {
      const response = await apiClient.get('/models');
      return response.data;
    } catch (error) {
      logger.error('Error fetching models:', error);
      throw error;
    }
  },

  // Save model configuration
  saveModelConfig: async (config) => {
    try {
      const response = await apiClient.post('/models', config);
      return response.data;
    } catch (error) {
      logger.error('Error saving model config:', error);
      throw error;
    }
  }
};

module.exports = deploymentService;
