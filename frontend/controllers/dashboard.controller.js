const dashboardService = require('../services/dashboard.service');
const logger = require('../utils/logger');

const dashboardController = {
  // 显示仪表板
  showDashboard: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const data = await dashboardService.getOverviewData(token);
      
      res.render('dashboard', {
        user: req.session.user,
        page: 'dashboard',
        ...data
      });
    } catch (error) {
      logger.error('Error rendering dashboard', error);
      res.status(500).render('error', {
        message: '无法加载仪表板数据',
        error: process.env.NODE_ENV === 'development' ? error : {}
      });
    }
  }
};

module.exports = dashboardController;