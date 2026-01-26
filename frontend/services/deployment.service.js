const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const deploymentService = {
  // Generate deployment artifacts
  generateDeployment: async (token, config) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/deploy/generate', config, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error generating deployment:', error);
      throw error;
    }
  },

  // Test service connection
  testConnection: async (token, serviceDetails) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      // Mocking test connection since backend doesn't have it yet
      // const response = await apiClient.post('/api/v1/deploy/test-connection', serviceDetails, axiosConfig);
      // return response.data;
      return { success: true, message: 'Connection successful' };
    } catch (error) {
      logger.error('Error testing connection:', error);
      throw error;
    }
  },

  // Get list of models
  getModels: async (token) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.get('/api/v1/configs/inference', axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error fetching models:', error);
      throw error;
    }
  },

  // Save model configuration
  saveModelConfig: async (token, config) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/configs/inference', config, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error saving model config:', error);
      throw error;
    }
  },

  // Save target nodes
  saveNodes: async (token, nodes) => {
    try {
      // Backend doesn't have /deploy/nodes yet
      return { success: true };
    } catch (error) {
      logger.error('Error saving nodes:', error);
      throw error;
    }
  },

  // Get target nodes
  getNodes: async (token) => {
    try {
      // Backend doesn't have /deploy/nodes yet
      return { nodes: [] };
    } catch (error) {
      logger.error('Error fetching nodes:', error);
      throw error;
    }
  },

  // Detect hardware
  detectHardware: async (token, nodes) => {
    try {
      // Backend doesn't have /deploy/detect-hardware yet
      return { success: true, hardware: 'Ascend NPU' };
    } catch (error) {
      logger.error('Error detecting hardware:', error);
      throw error;
    }
  }
};

module.exports = deploymentService;