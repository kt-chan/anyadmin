const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const importService = {
  getTasks: async (token) => {
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.get('/api/v1/import/tasks', config);
      return response.data;
    } catch (error) {
      logger.error('Error fetching import tasks', error);
      throw error;
    }
  },

  createTask: async (token, task) => {
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      const response = await apiClient.post('/api/v1/import/tasks', task, config);
      return response.data;
    } catch (error) {
      logger.error('Error creating import task', error);
      throw error;
    }
  },

  updateTask: async (token, id, updates) => {
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      // Backend doesn't have PUT /import/tasks/:id in router.go yet
      // return response.data;
      return true;
    } catch (error) {
      logger.error('Error updating import task', error);
      throw error;
    }
  },

  deleteTask: async (token, id) => {
    try {
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };
      // Backend doesn't have DELETE /import/tasks/:id in router.go yet
      return true;
    } catch (error) {
      logger.error('Error deleting import task', error);
      throw error;
    }
  }
};

module.exports = importService;