const express = require('express');
const router = express.Router();
const authController = require('../controllers/auth.controller');

// 登录页面
router.get('/login', authController.showLogin);

// 登录处理
router.post('/login', authController.handleLogin);

// 注销
router.get('/logout', authController.handleLogout);

module.exports = router;