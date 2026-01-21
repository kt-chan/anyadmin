const express = require('express');
const router = express.Router();
const deploymentController = require('../controllers/deployment.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// 部署配置页面（需要登录）
router.get('/', requireLogin, deploymentController.showDeployment);

module.exports = router;