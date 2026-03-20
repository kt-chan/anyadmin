const express = require('express');
const session = require('express-session');
const path = require('path');
require('dotenv').config({ path: path.join(__dirname, '../.env') });
const routes = require('./routes');
const appConfig = require('./config/app.config');
const sessionConfig = require('./config/session.config');
const logger = require('./utils/logger');

const app = express();
const PORT = appConfig.port;

// Trust first proxy (required for session cookies behind a proxy/container)
app.set('trust proxy', 1);

// ==================== 中间件配置 ====================
app.set('view engine', 'pug');
app.set('views', path.join(__dirname, 'views'));

// 静态文件服务
app.use(express.static(path.join(__dirname, 'public')));

// 解析请求体
app.use(express.urlencoded({ extended: true }));
app.use(express.json());

// Request logging middleware
app.use((req, res, next) => {
  logger.info(`${req.method} ${req.url}`);
  next();
});

// Session配置
app.use(session(sessionConfig));

// ==================== 注册路由 ====================
// 所有路由通过routes/index.js统一管理
app.use('/', routes);

// ==================== 错误处理 ====================

// 404处理 - 在所有路由之后
app.use((req, res) => {
  res.status(404).render('404', {
    message: '页面未找到',
    user: req.session.user || null
  });
});

// 全局错误处理中间件
app.use((err, req, res, next) => {
  logger.error('Server error:', err);
  
  // 根据环境决定是否暴露错误详情
  const errorDetails = appConfig.env === 'development' ? err.message : {};
  
  res.status(500).render('error', {
    user: req.session.user || null,
    message: '服务器内部错误',
    error: errorDetails
  });
});

// ==================== 启动服务器 ====================
const server = app.listen(PORT, () => {
  const address = server.address();
  const host = address.address === '::' ? 'localhost' : address.address;
  const port = address.port;
  
  logger.info(`Server started on http://${host}:${port}`);
  console.log(`
╔══════════════════════════════════════════════════════╗
║       知识库管理平台 MVP - 已成功启动                ║
╠══════════════════════════════════════════════════════╣
║ 🌐 访问地址: http://${host}:${port}                  ║
║ 📊 仪表板:   http://${host}:${port}/dashboard        ║
║ 🔧 服务管理: http://${host}:${port}/services         ║
║ 🚀 部署配置: http://${host}:${port}/deployment       ║
║ 💾 备份恢复: http://${host}:${port}/backup           ║
║ ⚙️  系统管理: http://${host}:${port}/system           ║
╠══════════════════════════════════════════════════════╣
║ 🔑 登录凭据:                                         ║
║   • 管理员:   admin / password                       ║
║   • 操作员:   operator_01 / password                 ║
╚══════════════════════════════════════════════════════╝
  `);
});

// 优雅关闭处理
process.on('SIGTERM', () => {
  logger.info('Received SIGTERM, shutting down...');
  server.close(() => {
    logger.info('Server closed');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  logger.info('Received SIGINT, shutting down...');
  server.close(() => {
    logger.info('Server closed');
    process.exit(0);
  });
});

// 导出app用于测试或其他模块
module.exports = app;