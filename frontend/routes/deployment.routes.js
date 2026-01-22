const express = require('express');
const router = express.Router();
const deploymentController = require('../controllers/deployment.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// Page
router.get('/', requireLogin, deploymentController.showDeployment);

// API Routes
router.post('/api/generate', requireLogin, deploymentController.generate);
router.post('/api/test-connection', requireLogin, deploymentController.testConnection);
router.get('/api/models', requireLogin, deploymentController.getModels);
router.post('/api/models', requireLogin, deploymentController.saveModel);

module.exports = router;