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

  // Get System SSH Key
  getSSHKey: async (token) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` },
        responseType: 'text'
      };
      const response = await apiClient.get('/api/v1/deploy/ssh-key', axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error getting SSH key:', error);
      throw error;
    }
  },

  // Test service connection
  testConnection: async (token, serviceDetails) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      
      const response = await apiClient.post('/api/v1/deploy/test-connection', serviceDetails, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error testing connection:', error);
      throw error;
    }
  },

  // Discover models from remote vLLM service
  discoverModels: async (token, { host, port }) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/deploy/vllm-models', { host, port }, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error discovering models:', error);
      throw error;
    }
  },

  // Save target nodes
  saveNodes: async (token, nodes) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/deploy/nodes', { nodes }, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error saving nodes:', error);
      throw error;
    }
  },

  // Get target nodes
  getNodes: async (token) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.get('/api/v1/deploy/nodes', axiosConfig);
      return response.data;
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