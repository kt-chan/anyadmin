const authService = require('../services/auth.service');

const authController = {
  // 显示登录页面
  showLogin: (req, res) => {
    // 如果用户已登录，重定向到仪表板
    if (req.session.user) {
      return res.redirect('/dashboard');
    }
    
    res.render('login');
  },
  
  // 处理登录
  handleLogin: async (req, res) => {
    try {
      const { username, password } = req.body;
      const user = await authService.authenticate(username, password);
      
      if (user) {
        req.session.user = user;
        return res.json({ 
          success: true, 
          redirect: '/dashboard',
          user: user
        });
      } else {
        return res.json({ 
          success: false, 
          message: '用户名或密码错误' 
        });
      }
    } catch (error) {
      console.error('Login error:', error);
      return res.status(500).json({
        success: false,
        message: '登录处理失败'
      });
    }
  },
  
  // 处理注销
  handleLogout: (req, res) => {
    req.session.destroy((err) => {
      if (err) {
        console.error('注销时发生错误:', err);
        return res.status(500).send('注销失败');
      }
      res.redirect('/login');
    });
  }
};

module.exports = authController;