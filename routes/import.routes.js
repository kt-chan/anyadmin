const express = require('express');
const router = express.Router();
const importController = require('../controllers/import.controller');
const authMiddleware = require('../middleware/auth.middleware');

// 页面路由
router.get('/', authMiddleware.requireLogin, importController.getImportPage);

// API 路由
router.get('/api/tasks', authMiddleware.requireLogin, importController.getTasks);
router.post('/api/create', authMiddleware.requireLogin, importController.createTask);
router.post('/api/operate', authMiddleware.requireLogin, importController.operateTask);
router.post('/api/global', authMiddleware.requireLogin, importController.operateGlobal);

module.exports = router;