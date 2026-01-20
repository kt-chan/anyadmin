const express = require('express');
const session = require('express-session');
const path = require('path');
const routes = require('./routes');

const app = express();
const PORT = process.env.PORT || 3000;

// ==================== 中间件配置 ====================
app.set('view engine', 'pug');
app.set('views', path.join(__dirname, 'views'));

// 静态文件服务
app.use(express.static(path.join(__dirname, 'public')));

// 解析请求体
app.use(express.urlencoded({ extended: true }));
app.use(express.json());

// Session配置
app.use(session({
  secret: 'knowledgebase-secret-key',
  resave: false,
  saveUninitialized: true,
  cookie: { 
    secure: false, // 生产环境应设置为true并使用HTTPS
    maxAge: 24 * 60 * 60 * 1000, // 24小时
    httpOnly: true
  }
}));

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
  console.error('服务器错误:', err.stack);
  
  // 根据环境决定是否暴露错误详情
  const errorDetails = process.env.NODE_ENV === 'development' ? err.message : {};
  
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
╠══════════════════════════════════════════════════════╣
║ 📁 项目结构:                                         ║
║   • app.js             - 主应用入口                  ║
║   • routes/            - 路由模块目录                ║
║   • controllers/       - 控制器目录                  ║
║   • views/             - 模板文件目录                ║
║   • public/            - 静态资源目录                ║
╠══════════════════════════════════════════════════════╣
║ ⚠️  注意: 应用使用Tailwind CSS和Font Awesome CDN     ║
║     请确保网络连接正常                               ║
╚══════════════════════════════════════════════════════╝
  `);
});

// 优雅关闭处理
process.on('SIGTERM', () => {
  console.log('收到SIGTERM信号，正在关闭服务器...');
  server.close(() => {
    console.log('服务器已关闭');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('收到SIGINT信号，正在关闭服务器...');
  server.close(() => {
    console.log('服务器已关闭');
    process.exit(0);
  });
});

// 导出app用于测试或其他模块
module.exports = app;