/**
 * 登录检查中间件
 * 如果用户未登录且访问的页面不是登录页，则重定向到登录页
 */
const requireLogin = (req, res, next) => {
  // 白名单：不需要登录的路径
  const whitelist = ['/login', '/api/login'];
  
  // 检查用户是否已登录
  if (!req.session.user && !whitelist.includes(req.path)) {
    return res.redirect('/login');
  }
  
  // 如果用户已登录但访问登录页，则重定向到仪表板
  if (req.session.user && req.path === '/login') {
    return res.redirect('/dashboard');
  }
  
  next();
};

module.exports = {
  requireLogin
};