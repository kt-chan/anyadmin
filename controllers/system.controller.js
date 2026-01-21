const systemService = require('../services/system.service');
const logger = require('../utils/logger');

const systemController = {
  // 显示系统管理页面
  showSystem: async (req, res) => {
    try {
      const data = await systemService.getSystemData();
      
      res.render('system', {
        user: req.session.user,
        users: data.users,
        auditLogs: data.auditLogs,
        page: 'system'
      });
    } catch (error) {
      logger.error('Error rendering system page', error);
      res.status(500).render('error', {
        message: '无法加载系统管理数据',
        error: process.env.NODE_ENV === 'development' ? error : {}
      });
    }
  }
};

module.exports = systemController;