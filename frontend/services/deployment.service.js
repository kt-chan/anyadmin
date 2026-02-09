const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');
const fs = require('fs').promises;
const path = require('path');

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

  // Discover models
  discoverModels: async (token, { host, port, mode }) => {
    try {
      if (mode === 'local' || mode === 'new_deployment') {
        // Local scan
        const modelsDir = path.join(__dirname, '../../backend/deployments/models');
        try {
          const files = await fs.readdir(modelsDir, { withFileTypes: true });
          const models = files
            .filter(dirent => dirent.isDirectory() && dirent.name !== '.tmp')
            .map(dirent => ({ id: dirent.name }));

          return { data: models };
        } catch (err) {
          logger.error('Error scanning local models:', err);
          return { data: [] }; // Return empty if dir doesn't exist or error
        }
      } else {
        // Remote vLLM service
        const axiosConfig = {
          headers: { Authorization: `Bearer ${token}` }
        };
        const response = await apiClient.post('/api/v1/deploy/vllm-models', { host, port }, axiosConfig);
        return response.data;
      }
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
  },

  // Check agent status
  checkAgentStatus: async (token, ip) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` },
        params: { ip }
      };
      const response = await apiClient.get('/api/v1/deploy/status', axiosConfig);
      return response.data;
    } catch (error) {
      // Don't log full error for 404 (not found yet) to avoid noise
      if (error.response && error.response.status === 404) {
        return { success: false, message: "Agent not yet online" };
      }
      logger.error('Error checking agent status:', error);
      throw error;
    }
  },

  // Control remote agent
  controlAgent: async (token, { ip, action }) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/deploy/agent/control', { ip, action }, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error(`Error ${action}ing agent on ${ip}:`, error);
      throw error;
    }
  },

  // Update vLLM Config
  updateVLLMConfig: async (token, { node_ip, config }) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/services/vllm/config', { node_ip, config }, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error updating vLLM config:', error);
      throw error;
    }
  }
};

module.exports = deploymentService;
