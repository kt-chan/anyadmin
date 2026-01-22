const deploymentService = require('../services/deployment.service');
const logger = require('../utils/logger');

const deploymentController = {
  // Render Deployment Page
  showDeployment: async (req, res) => {
    try {
      res.render('deployment', {
        user: req.session.user,
        page: 'deployment'
      });
    } catch (error) {
      logger.error('Error rendering deployment page:', error);
      res.status(500).render('error', { message: 'Failed to load deployment page' });
    }
  },

  // API: Generate Deployment
  generate: async (req, res) => {
    try {
      const result = await deploymentService.generateDeployment(req.body);
      res.json({ success: true, data: result });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Test Connection
  testConnection: async (req, res) => {
    try {
      const result = await deploymentService.testConnection(req.body);
      res.json(result);
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Get Models
  getModels: async (req, res) => {
    try {
      const models = await deploymentService.getModels();
      res.json({ success: true, data: models });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Save Model Config
  saveModel: async (req, res) => {
    try {
      const result = await deploymentService.saveModelConfig(req.body);
      res.json({ success: true, data: result });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  }
};

module.exports = deploymentController;