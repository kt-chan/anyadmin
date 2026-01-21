const express = require('express');
const router = express.Router();
const apiController = require('../controllers/api.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// 配置相关API
router.post('/config/save', requireLogin, apiController.saveConfig);
router.post('/config/concurrency', requireLogin, apiController.updateConcurrency);
router.post('/config/token-limit', requireLogin, apiController.updateTokenLimit);

// 服务操作API
router.post('/service/restart', requireLogin, apiController.restartService);
router.post('/service/stop', requireLogin, apiController.stopService);
router.get('/services/status', requireLogin, apiController.getServicesStatus);

// 备份操作API
router.post('/backup/create', requireLogin, apiController.createBackup);
router.post('/backup/restore', requireLogin, apiController.restoreFromBackup);
router.post('/backup/delete', requireLogin, apiController.deleteBackup);

// 系统操作API
router.post('/system/reflash', requireLogin, apiController.appReflash);

// 部署操作API
router.post('/deployment/next', requireLogin, apiController.nextDeploymentStep);

// 用户管理API
router.post('/user/create', requireLogin, apiController.createUser);

// 系统指标API
router.get('/metrics', requireLogin, apiController.getSystemMetrics);

// 调试API (公开访问)
router.get('/debug', apiController.debugInfo);

module.exports = router;