const { users } = require('../data/mockData');

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
  handleLogin: (req, res) => {
    const { username, password } = req.body;
    const user = users.find(u => u.username === username && u.password === password);
    
    if (user) {
      req.session.user = user;
      return res.json({ 
        success: true, 
        redirect: '/dashboard',
        user: {
          username: user.username,
          role: user.role
        }
      });
    } else {
      return res.json({ 
        success: false, 
        message: '用户名或密码错误' 
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