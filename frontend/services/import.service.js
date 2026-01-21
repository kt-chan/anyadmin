const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const importService = {
  getTasks: async () => {
    try {
      const response = await apiClient.get('/import/tasks');
      return response.data;
    } catch (error) {
      logger.error('Error fetching import tasks', error);
      throw error;
    }
  },

  createTask: async (task) => {
    try {
      const response = await apiClient.post('/import/tasks', task);
      return response.data;
    } catch (error) {
      logger.error('Error creating import task', error);
      throw error;
    }
  },

  updateTask: async (id, updates) => {
    try {
      const response = await apiClient.put(`/import/tasks/${id}`, updates);
      return response.data;
    } catch (error) {
      logger.error('Error updating import task', error);
      throw error;
    }
  },

  deleteTask: async (id) => {
    try {
      const response = await apiClient.delete(`/import/tasks/${id}`);
      return response.data;
    } catch (error) {
      logger.error('Error deleting import task', error);
      throw error;
    }
  }
};

module.exports = importService;
