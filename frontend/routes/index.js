const express = require('express');
const router = express.Router();

// 导入路由模块
const authRoutes = require('./auth.routes');
const dashboardRoutes = require('./dashboard.routes');
const deploymentRoutes = require('./deployment.routes');
const servicesRoutes = require('./services.routes');
const backupRoutes = require('./backup.routes');
const systemRoutes = require('./system.routes');
const importRoutes = require('./import.routes');
const modelsRoutes = require('./models.routes');
const apiRoutes = require('./api.routes');

// 注册路由
router.use('/', authRoutes);
router.use('/dashboard', dashboardRoutes);
router.use('/deployment', deploymentRoutes);
router.use('/models', modelsRoutes);
router.use('/services', servicesRoutes);
router.use('/import', importRoutes);
router.use('/backup', backupRoutes);
router.use('/system', systemRoutes);
router.use('/api', apiRoutes);

// 首页重定向到仪表板
router.get('/', (req, res) => {
  res.redirect('/dashboard');
});

module.exports = router;