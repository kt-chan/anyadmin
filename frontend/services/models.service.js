const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');
const FormData = require('form-data');
const fs = require('fs');

const modelsService = {
  // List models
  getModels: async (token) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` },
        timeout: 300000
      };
      const response = await apiClient.get('/api/v1/models', axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error fetching models:', error);
      throw error;
    }
  },

  // Init Upload
  initUpload: async (token, data) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` },
        timeout: 300000
      };
      const response = await apiClient.post('/api/v1/models/upload/init', data, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error init upload:', error);
      throw error;
    }
  },

  // Upload Chunk
  uploadChunk: async (token, uploadId, chunkFile) => {
    try {
      const form = new FormData();
      form.append('upload_id', uploadId);
      form.append('chunk', fs.createReadStream(chunkFile.path));

      const axiosConfig = {
        headers: { 
            Authorization: `Bearer ${token}`,
            ...form.getHeaders()
        },
        maxContentLength: Infinity,
        maxBodyLength: Infinity,
        timeout: 300000
      };

      const response = await apiClient.post('/api/v1/models/upload/chunk', form, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error uploading chunk:', error);
      throw error;
    }
  },

  // Finalize Upload
  finalizeUpload: async (token, data) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` },
        timeout: 300000
      };
      const response = await apiClient.post('/api/v1/models/finalize', data, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error('Error finalizing upload:', error);
      throw error;
    }
  },

  // Delete model
  deleteModel: async (token, name) => {
    try {
      const axiosConfig = {
        headers: { Authorization: `Bearer ${token}` },
        timeout: 300000
      };
      const response = await apiClient.delete(`/api/v1/models/${name}`, axiosConfig);
      return response.data;
    } catch (error) {
      logger.error(`Error deleting model ${name}:`, error);
      throw error;
    }
  }
};

module.exports = modelsService;