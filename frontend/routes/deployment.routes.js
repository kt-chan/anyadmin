const express = require('express');
const router = express.Router();
const deploymentController = require('../controllers/deployment.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// Page
router.get('/', requireLogin, deploymentController.showDeployment);

// API Routes
router.post('/api/generate', requireLogin, deploymentController.generate);
router.get('/api/ssh-key', requireLogin, deploymentController.getSSHKey);
router.post('/api/test-connection', requireLogin, deploymentController.testConnection);
router.post('/api/discover-models', requireLogin, deploymentController.discoverModels);

router.get('/api/nodes', requireLogin, deploymentController.getNodes);
router.post('/api/nodes', requireLogin, deploymentController.saveNodes);
router.post('/api/detect-hardware', requireLogin, deploymentController.detectHardware);

module.exports = router;