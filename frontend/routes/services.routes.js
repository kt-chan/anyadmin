const express = require('express');
const router = express.Router();
const servicesController = require('../controllers/services.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// 服务管理页面（需要登录）
router.get('/', requireLogin, servicesController.showServices);

module.exports = router;