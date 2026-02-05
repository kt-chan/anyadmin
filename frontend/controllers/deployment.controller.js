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
      const token = req.session.user?.token;
      const result = await deploymentService.generateDeployment(token, req.body);
      res.json({ success: true, data: result });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Get SSH Key
  getSSHKey: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const key = await deploymentService.getSSHKey(token);
      res.set('Content-Type', 'text/plain');
      res.send(key);
    } catch (error) {
      res.status(500).send('Failed to fetch SSH key');
    }
  },

  // API: Test Connection
  testConnection: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const result = await deploymentService.testConnection(token, req.body);
      res.json(result);
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Discover Models (Remote)
  discoverModels: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const result = await deploymentService.discoverModels(token, req.body);
      res.json({ success: true, data: result });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Save Nodes
  saveNodes: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const result = await deploymentService.saveNodes(token, req.body.nodes);
      res.json({ success: true, data: result });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Get Nodes
  getNodes: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const result = await deploymentService.getNodes(token);
      res.json({ success: true, data: result.nodes });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Detect Hardware
  detectHardware: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const result = await deploymentService.detectHardware(token, req.body.nodes);
      res.json(result);
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

  // API: Check Agent Status
  checkStatus: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const ip = req.query.ip;
      if (!ip) {
          return res.status(400).json({ success: false, message: "IP required" });
      }
      const result = await deploymentService.checkAgentStatus(token, ip);
      res.json(result);
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  },

    // API: Control Agent

    controlAgent: async (req, res) => {

      try {

        const token = req.session.user?.token;

        const result = await deploymentService.controlAgent(token, req.body);

        res.json({ success: true, data: result });

      } catch (error) {

        res.status(500).json({ success: false, message: error.message });

      }

    },

  // API: Update vLLM Config
  updateVLLMConfig: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const { node_ip, config } = req.body;
      const result = await deploymentService.updateVLLMConfig(token, { node_ip, config });
      res.json({ success: true, data: result });
    } catch (error) {
      res.status(500).json({ success: false, message: error.message });
    }
  }

  };

  

  module.exports = deploymentController;

  