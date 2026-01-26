const servicesService = require('../services/services.service');
const logger = require('../utils/logger');

const servicesController = {
  // 显示服务管理页面
  showServices: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const data = await servicesService.getServicesList(token);
      
      res.render('services', {
        user: req.session.user,
        services: data.services,
        page: 'services'
      });
    } catch (error) {
      logger.error('Error rendering services page', error);
      res.status(500).render('error', {
        message: '无法加载服务数据',
        error: process.env.NODE_ENV === 'development' ? error : {}
      });
    }
  }
};

module.exports = servicesController;