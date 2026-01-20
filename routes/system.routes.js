const express = require('express');
const router = express.Router();
const systemController = require('../controllers/system.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// 系统管理页面（需要登录）
router.get('/', requireLogin, systemController.showSystem);

module.exports = router;