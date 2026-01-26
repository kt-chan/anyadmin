const backupService = require('../services/backup.service');
const logger = require('../utils/logger');

const backupController = {
  // 显示备份恢复页面
  showBackup: async (req, res) => {
    try {
      const token = req.session.user?.token;
      const data = await backupService.getBackupData(token);
      
      res.render('backup', {
        user: req.session.user,
        backups: data.backups,
        page: 'backup'
      });
    } catch (error) {
      logger.error('Error rendering backup page', error);
      res.status(500).render('error', {
        message: '无法加载备份数据',
        error: process.env.NODE_ENV === 'development' ? error : {}
      });
    }
  }
};

module.exports = backupController;