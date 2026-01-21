// middleware/auth.middleware.js

/**
 * 认证中间件模块
 * 提供用户登录状态检查和权限验证功能
 */

/**
 * 检查用户是否已登录
 * 如果未登录且访问的页面不是白名单中的路径，则重定向到登录页
 * @returns {Function} Express中间件函数
 */
const requireLogin = (req, res, next) => {
  // 不需要登录验证的路径（白名单）
  const publicPaths = [
    '/login',           // 登录页面
    '/api/login',       // 登录API
    '/favicon.ico',     // 网站图标
    '/css/',           // CSS静态资源
    '/js/',            // JavaScript静态资源
    '/webfonts/',      // 字体文件
    '/fonts/'          // 字体文件
  ];
  
  // 检查当前路径是否在白名单中
  const isPublicPath = publicPaths.some(path => 
    req.path.startsWith(path) || req.path === path
  );
  
  // 获取用户信息
  const user = req.session.user;
  
  // 如果是API路径且需要验证
  if (req.path.startsWith('/api/') && !req.path.startsWith('/api/login')) {
    if (!user) {
      return res.status(401).json({
        success: false,
        message: '请先登录',
        code: 'UNAUTHORIZED'
      });
    }
    // API请求继续
    return next();
  }
  
  // 如果是公共路径或用户已登录，继续执行
  if (isPublicPath || user) {
    // 如果已登录用户访问登录页，重定向到仪表板
    if (user && req.path === '/login') {
      return res.redirect('/dashboard');
    }
    return next();
  }
  
  // 用户未登录且访问非公共路径，重定向到登录页
  // 保存原始请求URL以便登录后重定向
  req.session.returnTo = req.originalUrl || req.url;
  return res.redirect('/login');
};

/**
 * 检查用户角色权限
 * @param {Array} allowedRoles - 允许访问的角色数组
 * @returns {Function} Express中间件函数
 */
const requireRole = (allowedRoles) => {
  return (req, res, next) => {
    const user = req.session.user;
    
    // 首先检查是否登录
    if (!user) {
      if (req.xhr || req.headers.accept?.includes('application/json')) {
        return res.status(401).json({
          success: false,
          message: '请先登录',
          code: 'UNAUTHORIZED'
        });
      }
      req.session.returnTo = req.originalUrl || req.url;
      return res.redirect('/login');
    }
    
    // 检查角色权限
    if (!allowedRoles.includes(user.role)) {
      console.warn(`用户 ${user.username} (${user.role}) 试图访问需要 ${allowedRoles.join(', ')} 角色的资源: ${req.path}`);
      
      if (req.xhr || req.headers.accept?.includes('application/json')) {
        return res.status(403).json({
          success: false,
          message: '权限不足',
          code: 'FORBIDDEN'
        });
      }
      
      // 页面请求：显示权限不足页面或重定向
      return res.status(403).render('error', {
        user: req.session.user,
        message: '权限不足',
        error: {
          code: 'FORBIDDEN',
          message: '您没有权限访问此页面'
        }
      });
    }
    
    next();
  };
};

/**
 * 获取当前用户信息的中间件
 * 将用户信息添加到res.locals以便模板直接使用
 * @returns {Function} Express中间件函数
 */
const currentUser = (req, res, next) => {
  // 将用户信息添加到res.locals，这样所有模板都可以访问
  res.locals.user = req.session.user || null;
  
  // 将用户角色常量添加到模板变量中
  res.locals.ROLES = {
    ADMIN: 'admin',
    OPERATOR: 'operator'
  };
  
  // 添加辅助函数判断当前用户角色
  res.locals.isAdmin = () => {
    return req.session.user?.role === 'admin';
  };
  
  res.locals.isOperator = () => {
    return req.session.user?.role === 'operator';
  };
  
  // 添加当前页面路径，用于高亮导航菜单
  res.locals.currentPath = req.path;
  
  next();
};

/**
 * 检查是否已登录的重定向中间件
 * 与requireLogin不同，这个中间件不重定向，而是返回状态
 * 用于AJAX请求或需要返回JSON的情况
 * @returns {Function} Express中间件函数
 */
const checkLogin = (req, res, next) => {
  if (!req.session.user) {
    // 如果不是AJAX请求，直接重定向
    if (!req.xhr && !req.headers.accept?.includes('application/json')) {
      req.session.returnTo = req.originalUrl || req.url;
      return res.redirect('/login');
    }
    
    // 如果是AJAX或API请求，返回JSON错误
    return res.status(401).json({
      success: false,
      message: '会话已过期，请重新登录',
      code: 'SESSION_EXPIRED'
    });
  }
  
  next();
};

/**
 * 会话续期中间件
 * 每次请求都更新会话过期时间
 * @returns {Function} Express中间件函数
 */
const renewSession = (req, res, next) => {
  if (req.session.user) {
    // 每次请求都重置cookie过期时间
    req.session.cookie.maxAge = 24 * 60 * 60 * 1000; // 24小时
    req.session.touch(); // 更新会话时间
  }
  next();
};

/**
 * 记录用户活动的中间件
 * 记录用户访问的页面和操作
 * @returns {Function} Express中间件函数
 */
const logActivity = (req, res, next) => {
  // 跳过静态资源请求
  if (req.path.startsWith('/css/') || 
      req.path.startsWith('/js/') || 
      req.path.startsWith('/webfonts/') ||
      req.path === '/favicon.ico') {
    return next();
  }
  
  const user = req.session.user;
  const timestamp = new Date().toISOString();
  const method = req.method;
  const path = req.path;
  const ip = req.ip || req.connection.remoteAddress;
  
  // 记录用户活动
  if (user) {
    console.log(`[${timestamp}] ${user.username} (${user.role}) ${method} ${path} - IP: ${ip}`);
    
    // 可以将活动记录到数据库中
    if (req.session.activities) {
      req.session.activities.push({
        timestamp,
        method,
        path,
        ip
      });
      
      // 只保留最近50条活动记录
      if (req.session.activities.length > 50) {
        req.session.activities = req.session.activities.slice(-50);
      }
    } else {
      req.session.activities = [{
        timestamp,
        method,
        path,
        ip
      }];
    }
  } else {
    console.log(`[${timestamp}] Guest ${method} ${path} - IP: ${ip}`);
  }
  
  next();
};

/**
 * 验证CSRF令牌的中间件（可选）
 * 对于生产环境建议启用
 * @returns {Function} Express中间件函数
 */
const csrfProtection = (req, res, next) => {
  // 跳过GET、HEAD、OPTIONS请求
  if (['GET', 'HEAD', 'OPTIONS'].includes(req.method)) {
    return next();
  }
  
  // 跳过API登录请求
  if (req.path === '/api/login') {
    return next();
  }
  
  // 检查CSRF令牌
  const csrfToken = req.headers['x-csrf-token'] || req.body._csrf;
  const sessionToken = req.session.csrfToken;
  
  // 如果会话中没有CSRF令牌，生成一个
  if (!sessionToken) {
    req.session.csrfToken = require('crypto').randomBytes(16).toString('hex');
    return next();
  }
  
  // 验证CSRF令牌
  if (csrfToken !== sessionToken) {
    console.warn(`CSRF验证失败: 请求令牌=${csrfToken}, 会话令牌=${sessionToken}`);
    
    if (req.xhr || req.headers.accept?.includes('application/json')) {
      return res.status(403).json({
        success: false,
        message: 'CSRF令牌验证失败',
        code: 'CSRF_FAILED'
      });
    }
    
    return res.status(403).send('CSRF令牌验证失败');
  }
  
  next();
};

// 导出所有中间件
module.exports = {
  requireLogin,
  requireRole,
  currentUser,
  checkLogin,
  renewSession,
  logActivity,
  csrfProtection
};