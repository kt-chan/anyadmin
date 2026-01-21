const express = require('express');
const router = express.Router();
const backupController = require('../controllers/backup.controller');
const { requireLogin } = require('../middleware/auth.middleware');

// 备份恢复页面（需要登录）
router.get('/', requireLogin, backupController.showBackup);

module.exports = router;