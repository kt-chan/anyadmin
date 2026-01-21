const express = require('express');
const router = express.Router();
const dashboardController = require('../controllers/dashboard.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// 仪表板页面（需要登录）
router.get('/', requireLogin, dashboardController.showDashboard);

module.exports = router;