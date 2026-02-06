const servicesService = require('../services/services.service');
const logger = require('../utils/logger');

const servicesController = {
  // 显示服务管理页面
  showServices: async (req, res) => {
    try {
      const token = req.session.user?.token;
      // Fetch full config which includes nodes, inference_cfgs, rag_app_cfgs, system settings
      const fullConfig = await servicesService.getFullConfig(token);
      
      // We might still want status info if separate, but let's rely on config for now or merge if needed.
      // The previous view used 'services' array which was simplified.
      // The new view will need the raw structure.
      
      res.render('services', {
        user: req.session.user,
        config: fullConfig, // Pass the whole config tree
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